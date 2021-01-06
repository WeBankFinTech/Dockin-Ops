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
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/log"

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
	Done        chan bool
}

type WebsocketConnection struct {
	wsSocket  *websocket.Conn 	inChan    chan *WsMessage 	outChan   chan *WsMessage 	mutex     sync.Mutex      	isClosed  bool            	closeChan chan byte       }

func (websocketconn *WebsocketConnection) handleReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *WsMessage
		err     error
	)
	for {
				if msgType, data, err = websocketconn.wsSocket.ReadMessage(); err != nil {
			log.Logger.Infof("exit handleReadLoop as read socket msg %s", err.Error())
			return
		}
		msg = &WsMessage{
			MessageType: msgType,
			Data:        data,
		}
				select {
		case websocketconn.inChan <- msg:
			break
		case <-websocketconn.closeChan:
			log.Logger.Warnf("exit handleReadLoop as normal close")
			return
		}
	}
}

func (websocketconn *WebsocketConnection) handleWriteLoop() {
	var (
		msg *WsMessage
		err error
	)
	for {
		select {
				case msg = <-websocketconn.outChan:
						if err = websocketconn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				return
			}
			if msg.Done != nil {
				msg.Done <- true
			}
		case <-websocketconn.closeChan:
			log.Logger.Infof("exit handleWriteLoop as normal close")
			return
		}
	}
}

func InitWebsocket(resp http.ResponseWriter, req *http.Request) (wsConn *WebsocketConnection, err error) {
	var (
		wsSocket *websocket.Conn
	)
		if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		return
	}
	wsConn = &WebsocketConnection{
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

func (websocketconn *WebsocketConnection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case websocketconn.outChan <- &WsMessage{
		MessageType: messageType,
		Data:        data}:
	case <-websocketconn.closeChan:
		err = errors.New("exit write as websocket closed")
	}
	return
}

func (websocketconn *WebsocketConnection) WsBlockingWrite(messageType int, data []byte, ch chan bool) (err error) {
	select {
	case websocketconn.outChan <- &WsMessage{
		MessageType: messageType,
		Data:        data,
		Done:        ch}:
	case <-websocketconn.closeChan:
		err = errors.New("exit write as websocket closed")
	}

	select {
	case <-ch:
		log.Logger.Infof("blocking message %s write success", string(data))
	case <-time.After(time.Second * 2):
		log.Logger.Warnf("wait blocking message %s write timeout", string(data))
	}
	return
}

func (websocketconn *WebsocketConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg = <-websocketconn.inChan:
		return
	case <-websocketconn.closeChan:
		err = errors.New("exit read as websocket closed")
	}
	return
}

func (websocketconn *WebsocketConnection) WsClose() {
	websocketconn.wsSocket.Close()

	websocketconn.mutex.Lock()
	defer websocketconn.mutex.Unlock()
	if !websocketconn.isClosed {
		websocketconn.isClosed = true
		close(websocketconn.closeChan)
	}
}
