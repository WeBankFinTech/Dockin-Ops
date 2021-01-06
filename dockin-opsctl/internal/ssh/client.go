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

package ssh

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/webankfintech/dockin-opsctl/internal/log"
)

type WindowSize struct {
	Width  int
	Height int
}

type Client struct {
	Config *DockinExecConfig

	WsConn *websocket.Conn

	Stream ReadWriter

	Send chan rune

	WinSize chan WindowSize

	CloseChan chan byte
}

func NewClient(wsconn *websocket.Conn) *Client {
	return &Client{
		WsConn:    wsconn,
		Send:      make(chan rune),
		CloseChan: make(chan byte),
		WinSize:   make(chan WindowSize),
		Stream:    ReadWriter{},
	}
}

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		log.Debugf("exit read pump")
		c.Close()
	}()

	c.WsConn.SetReadDeadline(time.Now().Add(pongWait))
	c.WsConn.SetPongHandler(func(string) error {
		c.WsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		msgType, message, err := c.WsConn.ReadMessage()
		if err != nil {
			log.Debugf("read msg err:%v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debugf("error: %v", err)
			}
			return
		}

		c.Stream.Write(message)
		switch msgType {
		case websocket.CloseMessage:
			log.Debugf("remote proxy send close msg")
			return
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Debugf("exit write pump")
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case rece, ok := <-c.Send:
			c.WsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Debugf("send chan is closed, exit")
				c.WsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.WsConn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Debugf("create writer err:%v", err)
				return
			}

			msg := newCmdMessage(rece)
			buf := msg.ToByte()
			log.Debugf("write message to opserver:%s", string(buf))
			w.Write(buf)

			if err := w.Close(); err != nil {
				log.Debugf("try to close write err:%v", err)
				return
			}
		case rece, ok := <-c.WinSize:
			c.WsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Debugf("win size chan is closed, exit")
				return
			}

			w, err := c.WsConn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Debugf("create writer err:%v", err)
				return
			}

			msg := newResizeMessage(rece.Width, rece.Height)
			w.Write(msg.ToByte())

			if err := w.Close(); err != nil {
				log.Debugf("try to close write err:%v", err)
				return
			}
		case <-ticker.C:
			c.WsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Debugf("write ping msg failed, close write err:%v", err)
				return
			}
		case <-ctx.Done():
			log.Debugf("in write pump context is done")
			return
		}
	}
}

func (c *Client) Close() {
	c.WsConn.Close()
	c.CloseChan <- 0
}
