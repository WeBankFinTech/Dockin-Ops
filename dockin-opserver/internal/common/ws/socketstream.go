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

package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

type xtermMessage struct {
	MsgType string `json:"type"`  	Content string `json:"input"` 	Rows    uint16 `json:"rows"`  	Cols    uint16 `json:"cols"`  }

type SocketStream struct {
	WebsocketConnection *WebsocketConnection
	ResizeEvent         chan remotecommand.TerminalSize
}

func (stream *SocketStream) Read(p []byte) (size int, err error) {
	var (
		msg      *WsMessage
		xtermMsg xtermMessage
	)

		if msg, err = stream.WebsocketConnection.WsRead(); err != nil {
		return
	}

		if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil {
		return
	}
	//web ssh调整了终端大小
	if xtermMsg.MsgType == "resize" {
				stream.ResizeEvent <- remotecommand.TerminalSize{
			Width:  xtermMsg.Cols,
			Height: xtermMsg.Rows,
		}
	} else if xtermMsg.MsgType == "input" {
				size = len(xtermMsg.Content)
		copy(p, xtermMsg.Content)
	}
	return
}

func (stream *SocketStream) Write(p []byte) (size int, err error) {
	var (
		copyData []byte
	)
	copyData = make([]byte, len(p))
	copy(copyData, p)
	size = len(p)

	err = stream.WebsocketConnection.WsWrite(websocket.TextMessage, copyData)
	return
}

func (stream *SocketStream) Next() (size *remotecommand.TerminalSize) {
	ret := <-stream.ResizeEvent
	size = &ret

	return
}
