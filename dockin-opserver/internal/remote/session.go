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

package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/config"

	"github.com/webankfintech/dockin-opserver/internal/log"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	defaultSSHTimeout = time.Second * 15
)

type Resize interface {
	OnChange(winSize *WindowSize)
}

type WindowSize struct {
	rows int
	cols int
}

type ExecSession struct {
	conn           *websocket.Conn
	send           chan []byte
	Executor       Executor
	dockinExecParm *DockinExecParam
	sshIOManager   *SSHIOManager
	InterStream    *InteractStream
	clientIp       string
	isClosed       bool
	execMode       ExecMode
	IsRemoteClosed bool
	dockinParm     *DockinExecParam
}

func (base *ExecSession) HandleReceiveClientMsg(ctx context.Context, traceId string) {
	log.Logger.Infof("start to HandleReceiveClientMsg traceId=%s", traceId)
	defer func() {
		base.Close()
		log.Logger.Infof("close goroutine handle websocket read,traceId=%s", traceId)
	}()

	//base.conn.SetReadLimit(maxMessageSize)
	base.conn.SetReadDeadline(time.Now().Add(pongWait))
	base.conn.SetPongHandler(func(string) error { base.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, body, err := base.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Logger.Warnf("read message error: %v, traceId=%s", err, traceId)
			}
			break
		}
		log.Logger.Infof("receive request msg, clientIp:%s, msg:%s, traceId=%s", base.clientIp, string(body), traceId)

		msg, err := ParserToMessage(body)
		if err != nil {
			log.Logger.Warnf("failed to parse the message, body=%s, err=%v, traceId=%s", string(body), err, traceId)
			break
		}

		switch msg.Type {
		case MsgCmd:
			if err := base.handleClientInput(msg); err != nil {
				log.Logger.Warnf("write byte to ssh remote:%s, err:%v, traceId=%s", base.clientIp, err, traceId)
				return
			}
		case MsgResize:
			base.Executor.Resize(msg.Cols, msg.Rows)
		}
	}
}

func (base *ExecSession) HandleWriteClientMsg(ctx context.Context, traceId string) {
	log.Logger.Infof("start to HandleWriteClientMsg traceId=%s", traceId)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Logger.Infof("close goroutine handle websocket write,traceId=%s", traceId)
		ticker.Stop()
		base.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Logger.Infof("recv ctx done,exit HandleWriteClientMsg, traceId:%s", traceId)
			return
		case buf := <-base.InterStream.outBuffer:
			if len(buf) > 0 {
				base.handleRemoteOutput([]byte(buf))
				base.conn.SetWriteDeadline(time.Now().Add(writeWait))
				w, err := base.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Logger.Infof("create writer err:%v, traceId:%s", err, traceId)
					return
				}
				if _, err := w.Write([]byte(buf)); err != nil {
					log.Logger.Warnf("write date to client, err:%v, traceId:%s", err, traceId)
					return
				}

				if err := w.Close(); err != nil {
					log.Logger.Infof("close writer err:%v, traceId:%s", err, traceId)
					return
				}
				//base.InterStream.outBuffer.Reset()
			}
		case <-ticker.C:
			base.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := base.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (base *ExecSession) Close() {
	time.Sleep(100 * time.Millisecond)
	if base.isClosed {
		return
	}
	base.isClosed = true
	if !base.IsRemoteClosed {
		base.InterStream.inputBuffer <- []byte("exit")
	}
	base.conn.Close()
}

func (base *ExecSession) handleRemoteOutput(buffer []byte) {
	if base.execMode != SSHExecMode {
		return
	}
	if len(buffer) == 0 {
		return
	}

	base.sshIOManager.OnOutput(buffer)
}

func (base *ExecSession) handleClientInput(msg *Message) error {
	if base.execMode == SSHExecMode {
		log.CommandLogger.Info("ssh",
			zap.String("operator", base.dockinParm.UserName),
			zap.String("ip", base.clientIp),
			zap.String("command", msg.CmdLine),
			zap.String("timestamp", fmt.Sprintf("%d", time.Now().Unix())),
			zap.String("containerName", base.dockinParm.ContainerName),
			zap.String("rule", base.dockinParm.Rule))

		if err := base.sshIOManager.OnInput([]byte(msg.Cmd)); err != nil {
			base.InterStream.inputBuffer <- []byte{CharInterrupt}
			time.Sleep(time.Millisecond * 100)
			base.InterStream.inputBuffer <- []byte(err.Error())
			time.Sleep(time.Millisecond * 100)
			base.InterStream.inputBuffer <- []byte{CharInterrupt}
			return nil
		}
	}

	base.InterStream.inputBuffer <- []byte(msg.Cmd)
	return nil
}

func (base *ExecSession) Start(ctx context.Context, traceId string) error {
	go base.HandleReceiveClientMsg(ctx, traceId)
	go base.HandleWriteClientMsg(ctx, traceId)
	return nil
}

func (base *ExecSession) AddFilter(cf Filter) {
	base.sshIOManager.AddFilter(cf)
}

func CreateExecSession(ctx context.Context, dockinParm *DockinExecParam, conn *websocket.Conn, mode ExecMode) (*ExecSession, error) {
	log.Logger.Infof("start to CreateExecSession")
	docker := &ExecSession{
		conn: conn,
		InterStream: &InteractStream{
			outBuffer:   make(chan string),
			ctx:         ctx,
			inputBuffer: make(chan []byte, 32),
		},
		dockinParm: dockinParm,
	}
	ioFilter := NewIOFilter()
	sshContext := NewSSHContext()
	if dockinParm.WorkDir != "" {
		sshContext.WorkingDir = dockinParm.WorkDir
	}
	docker.execMode = mode
	docker.Executor = NewDockerExecutor(dockinParm.HostIP, config.OpsConfig.OpAgentPort)
	docker.sshIOManager = NewSSHIOManager(sshContext, ioFilter)
	docker.AddFilter(NewVimFilter(dockinParm.ContainerName, docker.Executor, sshContext.getCurrentWorkingDir))
	log.Logger.Infof("success to CreateExecSession %v", docker)
	return docker, nil
}
