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
	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/utils/aes"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	LoginType        = 1
	ResizeType       = 2
	StartCommandType = 3
	EndCommandType   = 4
	RawInputType     = 5
)

type Message struct {
	MessageType int         `json:"messageType"`
	Data        interface{} `json:"data"`
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

type CommandMessage struct {
}

func ParseAESMessage(content []byte) (*Message, error) {
	aes, err := aes.NewAes(common.ResKey)
	decoded, err := aes.AesDecrypt(string(content))
	if err != nil {
		err = errors.Errorf("decrypt err=%s", err.Error())
		log.Logger.Warnf(err.Error())
		return nil, err
	}
	msg := &Message{}
	err = jsoniter.Unmarshal([]byte(decoded), msg)
	if err != nil {
		log.Logger.Warnf("unmarshal read data error, check the input data, err=%s", err.Error())
		return nil, err
	}

	return msg, nil
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
