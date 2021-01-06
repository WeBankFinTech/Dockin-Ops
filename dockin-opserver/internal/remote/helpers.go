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
	"time"

	"github.com/webankfintech/dockin-opserver/internal/log"

	"github.com/gorilla/websocket"
)

var (
	CommandParseErrTemplate = "command format is invalid, please check the input command: %s"
)

func HandleWSError(ws *websocket.Conn, err error) bool {
	if err != nil {
		log.Logger.Warnf("handler ws err:%v", err)
		dt := time.Now().Add(time.Second)
		if err := ws.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			//if err := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); err != nil {
			log.Logger.Warnf("websocket writes control message failed:")
		}
		ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt)
		return true
	}
	return false
}
