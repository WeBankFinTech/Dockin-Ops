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

package ssh

import "time"

var DockinAesHeader = "access-token"

type PodInfo struct {
	PodName   string `json:"podName"`
	ClusterId string `json:"clusterId"`
	PodIp     string `json:"podIp"`
	HostIp    string `json:"hostIp"`
}

type OPServerResult struct {
	Code    int
	Message string
	Data    interface{}
}

type UserIdentity struct {
	UserName    string    `json:"userName"`
	Password    string    `json:"password"`
	Rule        string    `json:"rule"`
	Expire      int64     `json:"expire"`
	CreateTime  time.Time `json:"createTime"`
	AccessToken string    `json:"accessToken"`
}
