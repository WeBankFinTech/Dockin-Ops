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

import jsoniter "github.com/json-iterator/go"

const (
	MsgCmd    = "cmd"
	MsgResize = "resize"
)

type Message struct {
	Type    string `json:"type"`
	Cmd     string `json:"cmd"`
	Cols    int    `json:"Cols"`
	Rows    int    `json:"Rows"`
	CmdLine string `json:"cmdLine"`
}

func ParserToMessage(data []byte) (*Message, error) {
	msg := &Message{}
	if err := jsoniter.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
