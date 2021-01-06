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
	"net/http"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/log"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type AgentClient struct {
	opserverConn *websocket.Conn // 底层websocket
	podName      string
	loginUser    string
	send         chan []byte
	inChan       chan *Message       // 读取队列
	outChan      chan *OutputMessage // 发送队列
	closeChan    chan byte           // 关闭通知
}

func (c *AgentClient) readPump() {
	defer func() {
		c.opserverConn.Close()
	}()
	c.opserverConn.SetReadLimit(maxMessageSize)
	c.opserverConn.SetReadDeadline(time.Now().Add(pongWait))
	c.opserverConn.SetPongHandler(func(string) error { c.opserverConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	//for {
	//	_, message, err := c.opserverConn.ReadMessage()
	//	if err != nil {
	//		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
	//			log.Logger.Infof("un expected close error, podName=%s, loginUser=%s", c.podName, c.loginUser)
	//		}
	//		break
	//	}
	//	message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	//	c.hub.broadcast <- message
	//}
}

func (c *AgentClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.opserverConn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.opserverConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.opserverConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.opserverConn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.opserverConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.opserverConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *AgentClient) WsWrite(messageType int, data []byte) (err error) {
	select {
	case c.outChan <- &OutputMessage{
		MessageType: messageType,
		Data:        data,
	}:
	case <-c.closeChan:
		err = errors.New("exit write as websocket closed")
	}
	return
}

func (c *AgentClient) WsRead() (msg *Message, err error) {
	select {
	case msg = <-c.inChan:
		return
	case <-c.closeChan:
		err = errors.New("exit read as websocket closed")
	}
	return
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Logger.Warnf("failed to update connection to websocket, err=%v", err)
		return
	}
	client := &AgentClient{opserverConn: conn, send: make(chan []byte, 256)}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
