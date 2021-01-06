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

import "time"

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 1024
)

const (
	CharLineStart = 1
	CharBackward  = 2
	CharInterrupt = 3
	CharDelete    = 4
	CharLineEnd   = 5
	CharForward   = 6
	CharBell      = 7
	CharCtrlH     = 8
	CharTab       = 9
	CharCtrlJ     = 10
	CharKill      = 11
	CharCtrlL     = 12
	CharEnter     = 13
	CharNext      = 14
	CharPrev      = 16
	CharBckSearch = 18
	CharFwdSearch = 19
	CharTranspose = 20
	CharCtrlU     = 21
	CharCtrlW     = 23
	CharCtrlY     = 25
	CharCtrlZ     = 26
	CharEsc       = 27
	CharEscapeEx  = 91
	CharBackspace = 127
)

var (
	PromptPrefix = []rune{27, 93, 48, 59}
	AppSuffix    = "]$ "
	SuffixLen    = 3
	RootSuffix   = "]# "
)

type InputMode int

const (
	CollectMode  InputMode = 0
	InteractMode InputMode = 1
)

const (
	BlacklistMode string = "blacklist"
	WhitelistMode string = "whitelist"
)

type MachineType int

const (
	VirtualMachine MachineType = 0
	DockerMachine  MachineType = 1
)

type ExecMode int

const (
	InteractExecMode ExecMode = 0
	SSHExecMode      ExecMode = 1
	CommonExecMode   ExecMode = 2
)
