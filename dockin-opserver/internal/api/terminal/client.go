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

package terminal

import (
	"context"
	"sync"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	opsctlConn        *websocket.Conn
	opagentConn       *websocket.Conn
	bufferFromOpagent chan []byte
	bufferToOpagent   chan []byte
	isLogin           bool
	runningCmd        *CommandMessage
	identity          *model.UserIdentity
}

func (c *Client) opsctlMessageReadLoop(ctx context.Context, wg *sync.WaitGroup, traceId string) {
	defer func() {
		wg.Done()
		c.opsctlConn.Close()
	}()
	c.opsctlConn.SetReadLimit(maxMessageSize)
	c.opsctlConn.SetReadDeadline(time.Now().Add(pongWait))
	c.opsctlConn.SetPongHandler(func(string) error { c.opsctlConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.opsctlConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Logger.Infof("unexpected close error, ignore it, err=%v, traceId=%s", err, traceId)
			}
			break
		}
		msg, err := ParseAESMessage(message)
		if err != nil {
			log.Logger.Warnf("failed to parse message, err=%v, traceId=%s", err, traceId)
			continue
		}
		log.Logger.Infof("receive opsctl message=%#v, traceId=%s", msg, traceId)
		if msg.MessageType != LoginType && c.isLogin {
			log.Logger.Infof("not a login user, traceId=%s", traceId)
			c.opsctlConn.WriteMessage(websocket.CloseMessage, []byte("not a login user, traceId="+traceId))
			return
		}
		switch msg.MessageType {
		case LoginType:
			loginMsg := msg.Data.(*LoginMessage)
			if err := c.handleLogin(loginMsg); err != nil {
				log.Logger.Infof("login check err=%v, traceId=%s", err, traceId)
				c.opsctlConn.WriteMessage(websocket.CloseMessage, []byte(err.Error()+traceId))
				return
			}
			log.Logger.Infof("login success, login msg=%#v, traceId=%s", loginMsg, traceId)
		case ResizeType:
		case StartCommandType:
			cmdMsg := msg.Data.(*CommandMessage)
			if err := c.handleCommand(cmdMsg); err != nil {
				c.opsctlConn.WriteMessage(websocket.TextMessage, []byte(err.Error()+traceId))
				return
			}

			c.runningCmd = cmdMsg
			c.bufferToOpagent <- message
		case EndCommandType:
		case RawInputType:
			c.bufferToOpagent <- message
		}
	}
}

func (c *Client) opsctlMessageWriteLoop(ctx context.Context, wg *sync.WaitGroup, traceId string) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		wg.Done()
		ticker.Stop()
		c.opsctlConn.Close()
	}()
	for {
		select {
		case content, ok := <-c.bufferFromOpagent:
			c.opsctlConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.opsctlConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			msg, err := ParseMessage(content)
			if err != nil {
				continue
			}

			log.Logger.Infof("receive message from opagent=%#v, traceId=%s", msg, traceId)
			w, err := c.opsctlConn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Logger.Warnf("failed to returns a writer for the next message, traceId={}", traceId)
				return
			}
			switch msg.MessageType {
			case EndCommandType:
				c.runningCmd = nil
				fallthrough
			default:
				if _, err := w.Write(content); err != nil {
					log.Logger.Warnf("failed to write message from opagent to opsctl, err=%v, traceId=%s", err, traceId)
					return
				}
				if err := w.Close(); err != nil {
					log.Logger.Warnf("failed to close the writer about websocket, err=%v, traceId=%s", err, traceId)
					return
				}
			}
		case <-ticker.C:
			c.opsctlConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.opsctlConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-ctx.Done():
			log.Logger.Infof("exit the write to opsctl, traceId=%s", traceId)
		}
	}
}

func (c *Client) opagentMessageWriteLoop(ctx context.Context, wg *sync.WaitGroup, traceId string) {
	defer func() {
		wg.Done()
		c.opagentConn.Close()
	}()
	for {
		select {
		case content, ok := <-c.bufferToOpagent:
			c.opagentConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.opagentConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.opagentConn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Logger.Warnf("failed to returns a writer for the next message, err=%v, traceId={}", err, traceId)
				return
			}
			if _, err := w.Write(content); err != nil {
				log.Logger.Warnf("failed to write message to opagent, err=%v, traceId={}", err, traceId)
				return
			}
			if err := w.Close(); err != nil {
				log.Logger.Warnf("failed to close the writer about websocket, err=%v, traceId=%s", err, traceId)
				return
			}
		case <-ctx.Done():
			log.Logger.Infof("exit the write to opagent, traceId=%s", traceId)
		}
	}
}

func (c *Client) opagentMessageReadLoop(ctx context.Context, wg *sync.WaitGroup, traceId string) {
	defer func() {
		wg.Done()
		c.opagentConn.Close()
	}()

	for {
		_, content, err := c.opagentConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Logger.Infof("unexpected close error, ignore it, err=%v, traceId=%s", err, traceId)
			}
			break
		}
		log.Logger.Infof("read message buffer from opagent, message=%s, traceId=%s", string(content), traceId)
		c.bufferFromOpagent <- content
	}
}

func (c *Client) handleLogin(message *LoginMessage) error {
	log.Logger.Infof("handle login request, userName=%s, accessToken=%s, rule=%s",
		message.UserName, message.AccessToken, message.Rule)
	if message.Rule == "" {
		message.Rule = "default"
	}

	c.isLogin = true
	return nil
}

func (c *Client) handleCommand(msg *CommandMessage) error {
	return nil
}
