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
	"fmt"
	"net/http"
	"time"

	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type Attacher interface {
	// AttachContainer attaches to the running container in the pod, copying data between in/out/err
	// and the container's stdin/stdout/stderr.
	AttachContainer(container string, iostream *dockershim.IOStreams, tty bool, resize <-chan dockershim.TerminalSize, uid string) error
}

func ServeAttach(w http.ResponseWriter, req *http.Request, attacher Attacher, container string, streamOpts *Options, idleTimeout, streamCreationTimeout time.Duration, supportedProtocols []string, uid string) {
	ctx, ok := createStreams(req, w, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// error is handled by createStreams
		return
	}
	defer ctx.conn.Close()

	ioStream := &dockershim.IOStreams{
		In:     ctx.stdinStream,
		Out:    ctx.stdoutStream,
		ErrOut: ctx.stderrStream,
	}

	err := attacher.AttachContainer(container, ioStream, ctx.tty, ctx.resizeChan, uid)
	if err != nil {
		err = fmt.Errorf("error attaching to container: %v", err)
		runtime.HandleError(err)
		ctx.writeStatus(apierrors.NewInternalError(err))
	} else {
		ctx.writeStatus(&apierrors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusSuccess,
		}})
	}
}

func DebugAttach(w http.ResponseWriter, req *http.Request, attacher Attacher, container string, streamOpts *Options, idleTimeout, streamCreationTimeout time.Duration, supportedProtocols []string, uid string) {
	ctx, ok := createStreams(req, w, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// error is handled by createStreams
		return
	}
	defer ctx.conn.Close()

	ioStream := &dockershim.IOStreams{
		In:     ctx.stdinStream,
		Out:    ctx.stdoutStream,
		ErrOut: ctx.stderrStream,
	}

	err := attacher.AttachContainer(container, ioStream, ctx.tty, ctx.resizeChan, uid)
	if err != nil {
		err = fmt.Errorf("error attaching to container: %v", err)
		runtime.HandleError(err)
		ctx.writeStatus(apierrors.NewInternalError(err))
	} else {
		ctx.writeStatus(&apierrors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusSuccess,
		}})
	}
}
