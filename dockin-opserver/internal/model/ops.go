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
	jsoniter "github.com/json-iterator/go"
)

var (
	Success      = 0
	Failed       = -1
	WcsAesHeader = "access-token"
)

type OpsResult struct {
	Code    int
	Message string
	Data    interface{}
}

func FailedOpsResult(err error) *OpsResult {
	return &OpsResult{
		Code:    Failed,
		Message: err.Error(),
	}
}

func SuccessOpsResult(data interface{}) *OpsResult {
	return &OpsResult{
		Code:    Success,
		Message: "Success",
		Data:    data,
	}
}

func (o *OpsResult) ToByte() []byte {
	temp, _ := jsoniter.MarshalToString(o)
	return []byte(temp)
}

func (o *OpsResult) ToString() string {
	temp, _ := jsoniter.MarshalToString(o)
	return temp
}

type OpsOption struct {
	Command   string
	Resource  string
	Name      string
	PodIp     string
	PrintType string
	Flags     []string
	Params    map[string]interface{}

	Container string
	Namespace string
	Rule      string
	Operator  string
	ClusterId string
	HostIP    string

	UserName    string
	Password    string
	AccessToken string
	Env         []string
	User        string
	WorkDir     string
	Image       string
}

func (o *OpsOption) IsAdmin() bool {
	return o.Rule == "admin"
}

func (o *OpsOption) TTY() bool {
	value, exist := o.Params["tty"]
	if !exist {
		return false
	}

	return value.(bool)
}

func (o *OpsOption) Force() bool {
	value, exist := o.Params["force"]
	if !exist {
		return false
	}

	return value.(bool)
}

func (o *OpsOption) String() string {
	str, _ := jsoniter.MarshalToString(o)
	return str
}
