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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/webankfintech/dockin-opagent/internal/log"

	jsoniter "github.com/json-iterator/go"
)

const (
	ResizeType       = 2
	StartCommandType = 3
	EndCommandType   = 4
	RawInputType     = 5
)

type Message struct {
	MessageType int         `json:"messageType"`
	Data        interface{} `json:"data"`
}

func (m *Message) ToJSONBytes() []byte {
	b, _ := jsoniter.Marshal(m)
	return b
}

type LoginMessage struct {
	UserName    string `json:"userName"`
	Password    string `json:"password"`
	Rule        string `json:"rule"`
	AccessToken string `json:"accessToken"`
	PodName     string `json:"podName"`
}

type ResizeMessage struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ExecMessage struct {
	PodName    string   `json:"podName"`    // pod which want to execed in
	User       string   `json:"user"`       // User that will run the command
	Privileged bool     `json:"privileged"` // Is the container in privileged mode
	Tty        bool     `json:"tty"`        // Attach standard streams to a tty.
	Env        []string `json:"env"`        // Environment variables
	Cmd        []string `json:"cmd"`        // Execution commands and args
	WorkingDir string   `json:"workingDir"` // Working directory
}

func parseMessageFromHttpReq(req *http.Request, uid string) (*ExecMessage, error) {
	return &ExecMessage{}, nil
	paramStr, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Logger.Warnf("failed to parse the body about ExecParam, err=%v, uid=%s", err, uid)
		return nil, err
	}
	log.Logger.Infof("common exec, req param str=%s, uid=%s", string(paramStr), err)

	ep := &ExecMessage{}
	if err := jsoniter.Unmarshal(paramStr, ep); err != nil {
		log.Logger.Warnf("failed to serialize body to ExecParam, err=%v, uid=%s", err, uid)
		return nil, err
	}

	if ep.PodName == "" {
		return nil, fmt.Errorf("pod name is empty")
	}
	if len(ep.Cmd) == 0 {
		return nil, fmt.Errorf("command line is empty")
	}
	return ep, nil
}

func (e *ExecMessage) ToJSONString() string {
	str, _ := jsoniter.MarshalToString(e)
	return str
}

type CommandMessage struct {
}

func ParseMessage(content []byte) (*Message, error) {
	msg := &Message{}
	err := jsoniter.Unmarshal(content, msg)
	if err != nil {
		log.Logger.Warnf("unmarshal read data error, check the input data, err=%s", err.Error())
		return nil, err
	}

	return msg, nil
}
