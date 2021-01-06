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

type DockinExecConfig struct {
	UserName    string `json:"userName"`
	Password    string `json:"password"`
	AccessToken string `json:"accessToken"`

	Env []string `json:"env"`

	Width int `json:"width"`

	Height int `json:"height"`

	PodName string `json:"podName"`

	ContainerName string `json:"containerName"`

	Rule string `json:"rule"`

	Namespace string `json:"namespace"`

	User string `json:"user"`

	WorkDir string `json:"workDir"`

	Cmd string `json:"cmd"`

	TTY bool `json:"tty"`
}
