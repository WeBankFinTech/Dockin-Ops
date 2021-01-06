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
	"context"
	"sync"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/docker"
	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/log"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/gorilla/websocket"
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

type Client struct {
	// The websocket connection with opserver.
	OpserverConn *websocket.Conn
	// received msg from opserver, need to send to docker
	MsgFromOpserver chan []byte
	// received msg from docker, need to send to opserver
	MsgFromDocker chan []byte
	// resize buffer event
	ResizeChan chan dockershim.TerminalSize
	// command which is running, will set in the message start, and set to nil in the message end
	socketStream *SocketStream
	// DockerService
	DockerService *docker.DockerService
}

func (c *Client) OPServerMessageReadLoop(ctx context.Context, wg *sync.WaitGroup, uid string) {
	defer func() {
		wg.Done()
		c.OpserverConn.Close()
	}()
	c.OpserverConn.SetReadLimit(maxMessageSize)
	c.OpserverConn.SetReadDeadline(time.Now().Add(pongWait))
	c.OpserverConn.SetPongHandler(func(string) error { c.OpserverConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		msg := &Message{}
		err := c.OpserverConn.ReadJSON(msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Logger.Infof("unexpected close error, ignore it, err=%v, uid=%s", err, uid)
			}
			break
		}
		log.Logger.Infof("receive opserver message=%#v, uid=%s", msg, uid)
		// indicate the command is running, just sent to socket stream, otherwise means a new commands,
		// need to create a new docker exec process and bind a new socket stream
		if c.socketStream == nil {
			switch msg.MessageType {
			case ResizeType:
				resizeMsg := msg.Data.(*ResizeMessage)
				c.socketStream.TerminalSize <- dockershim.TerminalSize{
					Width:  resizeMsg.Width,
					Height: resizeMsg.Height,
				}
			case RawInputType:
				buf := msg.Data.([]byte)
				c.socketStream.BufferWriteToDocker <- buf
			default:
				log.Logger.Warnf("illegal message type=%d, uid=%s", msg.MessageType, uid)
			}
			return
		}

		switch msg.MessageType {
		case StartCommandType:
			c.socketStream = &SocketStream{
				OpserverWsConn: c.OpserverConn,
				TerminalSize:   c.ResizeChan,
				UID:            uid,
			}
			ios := &dockershim.IOStreams{
				In:     c.socketStream,
				Out:    c.socketStream,
				ErrOut: c.socketStream,
			}
			cmdMsg := msg.Data.(*ExecMessage)
			log.Logger.Infof("send interactive command to docker, uid=%s", uid)
			// this will block the exec until the ssh connection stop
			insp, err := c.DockerService.Exec(ctx, uid, cmdMsg.PodName, dockertypes.ExecConfig{
				User:         cmdMsg.User,
				Privileged:   cmdMsg.Privileged,
				Tty:          cmdMsg.Tty,
				AttachStdin:  true,
				AttachStderr: true,
				AttachStdout: true,
				Env:          cmdMsg.Env,
				WorkingDir:   cmdMsg.WorkingDir,
				Cmd:          cmdMsg.Cmd,
			}, ios, c.ResizeChan)

			if err != nil {
				log.Logger.Warnf("failed exec ssh command to docker, err=%v, uid=%s", err, uid)
			}
			log.Logger.Infof("exec ssh command to docker finished, err=%v, insp=%v, uid=%s", err, insp, uid)
			// write exit command type to opserver
			// todo
			exitCmdMsg := &Message{
				MessageType: EndCommandType,
			}
			c.MsgFromDocker <- exitCmdMsg.ToJSONBytes()
			c.socketStream = nil
		default:
			log.Logger.Warnf("illegal message type=%d, uid=%s", msg.MessageType, uid)
		}
	}
}

func (c *Client) OPServerMessageWriteLoop(ctx context.Context, wg *sync.WaitGroup, uid string) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		wg.Done()
		ticker.Stop()
		c.OpserverConn.Close()
	}()
	for {
		select {
		case content, ok := <-c.MsgFromDocker:
			c.OpserverConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.OpserverConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			msg, err := ParseMessage(content)
			if err != nil {
				continue
			}

			log.Logger.Infof("receive message from opagent=%#v, uid=%s", msg, uid)
			w, err := c.OpserverConn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Logger.Warnf("failed to returns a writer for the next message, uid={}", uid)
				return
			}
			switch msg.MessageType {
			case EndCommandType:
				log.Logger.Infof("write exit command type to opserver, uid=%s", uid)
				fallthrough
			default:
				// just write the content to opsctl
				if _, err := w.Write(content); err != nil {
					log.Logger.Warnf("failed to write message from opagent to opserver, err=%v, uid=%s", err, uid)
					return
				}
				if err := w.Close(); err != nil {
					log.Logger.Warnf("failed to close the writer about websocket, err=%v, uid=%s", err, uid)
					return
				}
			}
		case <-ticker.C:
			c.OpserverConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.OpserverConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-ctx.Done():
			log.Logger.Infof("exit the write to opserver, uid=%s", uid)
		}
	}
}
