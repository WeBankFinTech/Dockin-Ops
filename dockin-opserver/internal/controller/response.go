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

package controller

import "github.com/json-iterator/go"

var (
	Success = 0
	RedisFail = -1
)

type Response struct {
	Code int 			`json:"code"`
	Msg string 			`json:"msg"`
	TaskId string       `json:"taskId"`
	Data interface{}	`json:"data"`
}

func (r *Response)ToJSONBytes() []byte {
	json, _ := jsoniter.Marshal(r)
	return json
}