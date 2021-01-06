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

package wsstream

import (
	"encoding/base64"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"k8s.io/apimachinery/pkg/util/runtime"
)

const binaryWebSocketProtocol = "binary.k8s.io"

const base64BinaryWebSocketProtocol = "base64.binary.k8s.io"

type ReaderProtocolConfig struct {
	Binary bool
}

func NewDefaultReaderProtocols() map[string]ReaderProtocolConfig {
	return map[string]ReaderProtocolConfig{
		"":                            {Binary: true},
		binaryWebSocketProtocol:       {Binary: true},
		base64BinaryWebSocketProtocol: {Binary: false},
	}
}

type Reader struct {
	err              chan error
	r                io.Reader
	ping             bool
	timeout          time.Duration
	protocols        map[string]ReaderProtocolConfig
	selectedProtocol string

	handleCrash func() // overridable for testing
}

func NewReader(r io.Reader, ping bool, protocols map[string]ReaderProtocolConfig) *Reader {
	return &Reader{
		r:           r,
		err:         make(chan error),
		ping:        ping,
		protocols:   protocols,
		handleCrash: func() { runtime.HandleCrash() },
	}
}

func (r *Reader) SetIdleTimeout(duration time.Duration) {
	r.timeout = duration
}

func (r *Reader) handshake(config *websocket.Config, req *http.Request) error {
	supportedProtocols := make([]string, 0, len(r.protocols))
	for p := range r.protocols {
		supportedProtocols = append(supportedProtocols, p)
	}
	return handshake(config, req, supportedProtocols)
}

func (r *Reader) Copy(w http.ResponseWriter, req *http.Request) error {
	go func() {
		defer r.handleCrash()
		websocket.Server{Handshake: r.handshake, Handler: r.handle}.ServeHTTP(w, req)
	}()
	return <-r.err
}

func (r *Reader) handle(ws *websocket.Conn) {
	// Close the connection when the client requests it, or when we finish streaming, whichever happens first
	closeConnOnce := &sync.Once{}
	closeConn := func() {
		closeConnOnce.Do(func() {
			ws.Close()
		})
	}

	negotiated := ws.Config().Protocol
	r.selectedProtocol = negotiated[0]
	defer close(r.err)
	defer closeConn()

	go func() {
		defer runtime.HandleCrash()
		// This blocks until the connection is closed.
		// Client should not send anything.
		IgnoreReceives(ws, r.timeout)
		// Once the client closes, we should also close
		closeConn()
	}()

	r.err <- messageCopy(ws, r.r, !r.protocols[r.selectedProtocol].Binary, r.ping, r.timeout)
}

func resetTimeout(ws *websocket.Conn, timeout time.Duration) {
	if timeout > 0 {
		ws.SetDeadline(time.Now().Add(timeout))
	}
}

func messageCopy(ws *websocket.Conn, r io.Reader, base64Encode, ping bool, timeout time.Duration) error {
	buf := make([]byte, 2048)
	if ping {
		resetTimeout(ws, timeout)
		if base64Encode {
			if err := websocket.Message.Send(ws, ""); err != nil {
				return err
			}
		} else {
			if err := websocket.Message.Send(ws, []byte{}); err != nil {
				return err
			}
		}
	}
	for {
		resetTimeout(ws, timeout)
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if n > 0 {
			if base64Encode {
				if err := websocket.Message.Send(ws, base64.StdEncoding.EncodeToString(buf[:n])); err != nil {
					return err
				}
			} else {
				if err := websocket.Message.Send(ws, buf[:n]); err != nil {
					return err
				}
			}
		}
	}
}
