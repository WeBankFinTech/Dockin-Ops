/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package remotecommand

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/log"

	api "github.com/webankfintech/dockin-opagent/internal/server/api/core"
	"github.com/webankfintech/dockin-opagent/internal/ws/wsstream"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
	remotecommandconsts "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
)

type Options struct {
	Stdin  bool
	Stdout bool
	Stderr bool
	TTY    bool
}

func NewOptions(req *http.Request) (*Options, error) {
	tty := req.FormValue(api.ExecTTYParam) == "1"
	stdin := req.FormValue(api.ExecStdinParam) == "1"
	stdout := req.FormValue(api.ExecStdoutParam) == "1"
	stderr := req.FormValue(api.ExecStderrParam) == "1"

	if tty && stderr {
		// TODO: make this an error before we reach this method
		klog.V(4).Infof("Access to exec with tty and stderr is not supported, bypassing stderr")
		stderr = false
	}

	if !stdin && !stdout && !stderr {
		return nil, fmt.Errorf("you must specify at least 1 of stdin, stdout, stderr")
	}

	return &Options{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		TTY:    tty,
	}, nil
}

type context struct {
	conn         io.Closer
	stdinStream  io.ReadCloser
	stdoutStream io.WriteCloser
	stderrStream io.WriteCloser
	writeStatus  func(status *apierrors.StatusError) error
	resizeStream io.ReadCloser
	resizeChan   chan dockershim.TerminalSize
	tty          bool
}

type streamAndReply struct {
	httpstream.Stream
	replySent <-chan struct{}
}

func waitStreamReply(replySent <-chan struct{}, notify chan<- struct{}, stop <-chan struct{}) {
	select {
	case <-replySent:
		notify <- struct{}{}
	case <-stop:
	}
}

func createStreams(req *http.Request, w http.ResponseWriter, opts *Options, supportedStreamProtocols []string, idleTimeout, streamCreationTimeout time.Duration) (*context, bool) {
	log.Logger.Infof("start createStreams")
	var ctx *context
	var ok bool

	if wsstream.IsWebSocketRequest(req) {
		ctx, ok = createWebSocketStreams(req, w, opts, idleTimeout)
	} else {
		ctx, ok = createHttpStreamStreams(req, w, opts, supportedStreamProtocols, idleTimeout, streamCreationTimeout)
	}

	if !ok {
		return nil, false
	}

	if ctx.resizeStream != nil {
		ctx.resizeChan = make(chan dockershim.TerminalSize)
		go handleResizeEvents(ctx.resizeStream, ctx.resizeChan)
	}

	return ctx, true
}

func createHttpStreamStreams(req *http.Request, w http.ResponseWriter, opts *Options, supportedStreamProtocols []string, idleTimeout, streamCreationTimeout time.Duration) (*context, bool) {
	log.Logger.Infof("start createHttpStreamStreams")
	protocol, err := httpstream.Handshake(req, w, supportedStreamProtocols)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, false
	}

	streamCh := make(chan streamAndReply)

	upgrader := spdy.NewResponseUpgrader()
	conn := upgrader.UpgradeResponse(w, req, func(stream httpstream.Stream, replySent <-chan struct{}) error {
		streamCh <- streamAndReply{Stream: stream, replySent: replySent}
		return nil
	})
	// from this point on, we can no longer call methods on response
	if conn == nil {
		// The upgrader is responsible for notifying the client of any errors that
		// occurred during upgrading. All we can do is return here at this point
		// if we weren't successful in upgrading.
		return nil, false
	}

	conn.SetIdleTimeout(idleTimeout)

	var handler protocolHandler
	switch protocol {
	case remotecommandconsts.StreamProtocolV4Name:
		handler = &v4ProtocolHandler{}
	case remotecommandconsts.StreamProtocolV3Name:
		handler = &v3ProtocolHandler{}
	case remotecommandconsts.StreamProtocolV2Name:
		handler = &v2ProtocolHandler{}
	case "":
		klog.V(4).Infof("Client did not request protocol negotiation. Falling back to %q", remotecommandconsts.StreamProtocolV1Name)
		fallthrough
	case remotecommandconsts.StreamProtocolV1Name:
		handler = &v1ProtocolHandler{}
	}

	// count the streams client asked for, starting with 1
	expectedStreams := 1
	if opts.Stdin {
		expectedStreams++
	}
	if opts.Stdout {
		expectedStreams++
	}
	if opts.Stderr {
		expectedStreams++
	}
	if opts.TTY && handler.supportsTerminalResizing() {
		expectedStreams++
	}

	expired := time.NewTimer(streamCreationTimeout)
	defer expired.Stop()

	ctx, err := handler.waitForStreams(streamCh, expectedStreams, expired.C)
	if err != nil {
		runtime.HandleError(err)
		return nil, false
	}

	ctx.conn = conn
	ctx.tty = opts.TTY

	return ctx, true
}

type protocolHandler interface {
	// waitForStreams waits for the expected streams or a timeout, returning a
	// remoteCommandContext if all the streams were received, or an error if not.
	waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error)
	// supportsTerminalResizing returns true if the protocol handler supports terminal resizing
	supportsTerminalResizing() bool
}

type v4ProtocolHandler struct{}

func (*v4ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error) {
	ctx := &context{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v4WriteStatusFunc(stream) // write json errors
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeResize:
				ctx.resizeStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("Unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

func (*v4ProtocolHandler) supportsTerminalResizing() bool { return true }

type v3ProtocolHandler struct{}

func (*v3ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error) {
	ctx := &context{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeResize:
				ctx.resizeStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("Unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

func (*v3ProtocolHandler) supportsTerminalResizing() bool { return true }

type v2ProtocolHandler struct{}

func (*v2ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error) {
	ctx := &context{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("Unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

func (*v2ProtocolHandler) supportsTerminalResizing() bool { return false }

type v1ProtocolHandler struct{}

func (*v1ProtocolHandler) waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*context, error) {
	ctx := &context{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streams:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)

				// This defer statement shouldn't be here, but due to previous refactoring, it ended up in
				// here. This is what 1.0.x kubelets do, so we're retaining that behavior. This is fixed in
				// the v2ProtocolHandler.
				defer stream.Reset()

				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("Unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	if ctx.stdinStream != nil {
		ctx.stdinStream.Close()
	}

	return ctx, nil
}

func (*v1ProtocolHandler) supportsTerminalResizing() bool { return false }

func handleResizeEvents(stream io.Reader, channel chan<- dockershim.TerminalSize) {
	defer runtime.HandleCrash()

	decoder := json.NewDecoder(stream)
	for {
		size := dockershim.TerminalSize{}
		if err := decoder.Decode(&size); err != nil {
			break
		}
		channel <- size
	}
}

func v1WriteStatusFunc(stream io.Writer) func(status *apierrors.StatusError) error {
	return func(status *apierrors.StatusError) error {
		if status.Status().Status == metav1.StatusSuccess {
			return nil // send error messages
		}
		_, err := stream.Write([]byte(status.Error()))
		return err
	}
}

func v4WriteStatusFunc(stream io.Writer) func(status *apierrors.StatusError) error {
	return func(status *apierrors.StatusError) error {
		bs, err := json.Marshal(status.Status())
		if err != nil {
			return err
		}
		_, err = stream.Write(bs)
		return err
	}
}
