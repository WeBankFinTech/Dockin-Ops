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

package exec

import (
	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/log"

	"github.com/gorilla/websocket"
)

type SocketStream struct {
	OpserverWsConn *websocket.Conn
	TerminalSize   chan dockershim.TerminalSize
	// buffer from opserver, need to write to docker
	BufferWriteToDocker chan []byte
	// buffer from docker ,need to write to opagent
	BufferWriteToOpagent chan []byte
	UID                  string
}

func (stream *SocketStream) Read(p []byte) (size int, err error) {
	// read will blocking until the data is ready or the connection is closed
	buf := <-stream.BufferWriteToDocker

	log.Logger.Infof("read message from opserver %#v, uid=%s", string(buf), stream.UID)
	size = len(buf)
	copy(p, buf)
	log.Logger.Infof("success handle opagent buffer, data=%v, uid=%s", string(buf), stream.UID)
	return
}

func (stream *SocketStream) Write(p []byte) (size int, err error) {
	var (
		copyData []byte
	)
	copyData = make([]byte, len(p))
	copy(copyData, p)
	size = len(p)

	log.Logger.Infof("write data from docker to opagent, data=%v, uid=%s", string(copyData), stream.UID)
	stream.BufferWriteToOpagent <- copyData
	log.Logger.Infof("write data from docker to opserver, data=%v, uid=%s", string(copyData), stream.UID)
	return
}
