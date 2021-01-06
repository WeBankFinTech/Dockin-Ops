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

package model

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

const (
	OK       = 0
	ErrParam = -1
	ErrExec  = -2
)

type AgentResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (a *AgentResult) ToString() string {
	return fmt.Sprintf("%+v", a)
}

func (a *AgentResult) ToJSONString() string {
	d, _ := jsoniter.MarshalToString(a)
	return d
}

func (a *AgentResult) ToJSONByte() []byte {
	d, _ := jsoniter.Marshal(a)
	return d
}

func NewSuccessAgentResult(data interface{}) *AgentResult {
	return &AgentResult{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func NewErrorAgentResult(err error) *AgentResult {
	return &AgentResult{
		Code:    -1,
		Message: err.Error(),
	}
}

func NewErrorAgentResultWithCode(err error, code int) *AgentResult {
	return &AgentResult{
		Code:    -1,
		Message: err.Error(),
	}
}
