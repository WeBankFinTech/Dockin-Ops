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

type RmResultDto struct {
	Code    int             `json:"code"`
	Data    []*RmResultData `json:"data"`
	Message string          `json:"message"`
}

type OneRmResultDto struct {
	Code    int           `json:"code"`
	Data    *RmResultData `json:"data"`
	Message string        `json:"message"`
}

type RmResultData struct {
	PodName     string `json:"podName"`
	SubSystem   string `json:"subSystem"`
	SubSystemId string `json:"subSystemId"`
	Dcn         string `json:"dcn"`
	PodIP       string `json:"podIp"`
	Gateway     string `json:"gateway"`
	SubnetMask  string `json:"subnetMask"`
	Namespace   string `json:"namespace"`
	HostIP      string `json:"hostIp"`
	CPU         string `json:"cpu"`
	Mem         string `json:"mem"`
	Type        string `json:"type"`
	Port        int    `json:"port"`
	State       string `json:"state"`
	ClusterID   string `json:"clusterId"`
	Status      string `json:"status"`
}

type RmClusterData struct {
	Code int `json:"code"`
	Data struct {
		ClusterID string `json:"clusterId"`
	} `json:"data"`
}
