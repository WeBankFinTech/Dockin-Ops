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

	"github.com/webankfintech/dockin-opserver/internal/model"
)

type DockinExecParam struct {
	UserName      string   `json:"userName"`
	Password      string   `json:"password"`
	AccessToken   string   `json:"accessToken"`
	Env           []string `json:"env"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	PodName       string   `json:"podName"`
	ContainerName string   `json:"containerName"`
	Rule          string   `json:"rule"`
	Namespace     string   `json:"namespace"`
	User          string   `json:"user"`
	WorkDir       string   `json:"workDir"`
	Cmd           []string `json:"cmd"`
	TTY           bool     `json:"tty"`
	HostIP        string   `json:"hostIp"`
	Image         string   `json:"image"`
}

func OpsOption2ExecParam(opsOpts *model.OpsOption) *DockinExecParam {
	execParam := &DockinExecParam{
		UserName:      opsOpts.UserName,
		Password:      opsOpts.Password,
		AccessToken:   opsOpts.AccessToken,
		Env:           opsOpts.Env,
		PodName:       opsOpts.Name,
		ContainerName: opsOpts.Container,
		Rule:          opsOpts.Rule,
		Namespace:     opsOpts.Namespace,
		User:          opsOpts.User,
		WorkDir:       opsOpts.WorkDir,
		Cmd:           opsOpts.Flags,
		Image:         opsOpts.Image,
	}

	tty := opsOpts.TTY()
	if tty {
		width, _ := opsOpts.Params["width"].(float64)
		height, _ := opsOpts.Params["height"].(float64)
		execParam.Width = int(width)
		execParam.Height = int(height)
		execParam.TTY = tty
	}

	return execParam
}

type ExecConfig struct {
	Command string            `json:"command" validate:"required"`
	Env     map[string]string `json:"env"`
	WorkDir string            `json:"workDir"`
	User    string            `json:"user"`
	Tty     bool              `json:"tty"`
}

type SSHConfig struct {
	Cols    int
	Rows    int
	Timeout time.Duration
	Address string
	Port    int
}

type VmConfig struct {
	SSHConfig
	UserName  string
	PassWord  string
	PublicKey string
}

type DockerConfig struct {
	SSHConfig
	ContainerName string
	Address       string
	Port          int32
}
