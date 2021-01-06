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
	"fmt"
	"net/http"
	"sync"

	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Terminal struct {
	Cm *client.Manager
}

func NewTerminal(cm *client.Manager) *Terminal {
	t := &Terminal{
		Cm: cm,
	}
	http.HandleFunc("/v1/dockin/opserver/terminal", t.handler)
	return t
}

func (t *Terminal) handler(writer http.ResponseWriter, req *http.Request) {
	var (
		opsctlConn *websocket.Conn
		err        error
	)

	traceId := trace.TraceID()
	if opsctlConn, err = wsUpgrader.Upgrade(writer, req, nil); err != nil {
		log.Logger.Warnf("failed to upgrade connection to websocket, err=%v", err)
		return
	}
	log.Logger.Infof("make connection with opserver success, traceId=%s", traceId)

	ctx, cancel := context.WithCancel(req.Context())
	wg := new(sync.WaitGroup)
	wg.Add(4)
	defer cancel()

	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}
	opagentConn, _, err := dialer.Dial(fmt.Sprintf("%s:%d", "", config.OpsConfig.OpAgentPort), nil)
	client := &Client{
		opsctlConn:        opsctlConn,
		opagentConn:       opagentConn,
		bufferFromOpagent: make(chan []byte, 1024),
		bufferToOpagent:   make(chan []byte, 1024),
	}

	go client.opsctlMessageWriteLoop(ctx, wg, traceId)
	go client.opsctlMessageReadLoop(ctx, wg, traceId)
	go client.opagentMessageWriteLoop(ctx, wg, traceId)
	go client.opagentMessageReadLoop(ctx, wg, traceId)

	wg.Wait()
	log.Logger.Infof("close connection traceId=%s", traceId)
}
