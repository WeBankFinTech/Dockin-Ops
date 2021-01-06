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

type ProcessReqDTO struct {
	PodName  string   `json:"podName"`
	UserName string   `json:"userName"`
	Include  []string `json:"include"`
	Exclude  []string `json:"exclude"`
	UID      string   `json:"uid"`
}

type ProcessDTO struct {
	ProcessName string  `json:"processName"`
	Status      string  `json:"status"`
	CPU         float64 `json:"cpu"`
	Mem         float32 `json:"mem"`
	Rss         uint64  `json:"rss"`
	Vms         uint64  `json:"vms"`
	Fd          int32   `json:"fd"`
	UserName    string  `json:"userName"`
	CreateTime  int64   `json:"createTime"`
}
