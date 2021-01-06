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

package common

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsMessage struct {
	MessageType int
	Data        []byte
}

type WsConnection struct {
	wsSocket *websocket.Conn
	inChan   chan *WsMessage
	outChan  chan *WsMessage

	mutex     sync.Mutex
	isClosed  bool
	closeChan chan byte
}

func (wsConn *WsConnection) handleReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *WsMessage
		err     error
	)
	for {
		if msgType, data, err = wsConn.wsSocket.ReadMessage(); err != nil {
			goto ERROR
		}
		msg = &WsMessage{
			msgType,
			data,
		}
		wsConn.inChan <- msg
	}
ERROR:
	wsConn.WsClose()
}

func (wsConn *WsConnection) handleWriteLoop() {
	var (
		msg *WsMessage
		err error
	)
	for {
		select {
		case msg = <-wsConn.outChan:
			if err = wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				goto ERROR
			}
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}

func InitWebsocket(resp http.ResponseWriter, req *http.Request) (wsConn *WsConnection, err error) {
	var (
		wsSocket *websocket.Conn
	)
	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		return
	}
	wsConn = &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *WsMessage),
		outChan:   make(chan *WsMessage),
		closeChan: make(chan byte),
		isClosed:  false,
	}

	go wsConn.handleReadLoop()

	go wsConn.handleWriteLoop()

	return
}

func (wsConn *WsConnection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case wsConn.outChan <- &WsMessage{messageType, data}:
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg = <-wsConn.inChan:
		fmt.Println(string(msg.Data))
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

func (wsConn *WsConnection) WsClose() {
	wsConn.wsSocket.Close()

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}
