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
	"bytes"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/remote/prompt"
)

const (
	CR     byte = 13
	LF     byte = 10
	ESC    byte = 27
	Escape byte = 91
	Cursor byte = 75
)

type IOFilter struct {
	parser            prompt.ConsoleParser
	buf               *prompt.Buffer
	keyBindings       []prompt.KeyBind
	ASCIICodeBindings []prompt.ASCIICodeBind
	LastCtrlKey       prompt.Key
	InBytes           []byte
	isEsc             bool
	isEscape          bool
}

func NewIOFilter() *IOFilter {
	return &IOFilter{
		parser: prompt.NewStandardInputParser(),
		buf:    prompt.NewBuffer(),
	}
}

func (io *IOFilter) handleInput(data []byte) {
	io.LastCtrlKey = io.parser.GetKey(data)
	log.Logger.Infof("user type:%v", data)
}

func (io *IOFilter) handleOutput(data []byte) {
	log.Logger.Infof("receive output:%v, byte:%v", string(data), data)
	switch io.LastCtrlKey {
	case prompt.Enter, prompt.ControlC, prompt.ControlJ:
		log.Logger.Infof("last click is enter or contro-c, ignore the response:%s", string(data))
		return
	}
	if IsShellPrompt(data) {

		return
	}

	log.Logger.Infof("before filter output is :%v, index:%v", io.buf.Text(), io.buf.Document().CursorPositionCol())
	prompt.WalkRemoteOutput(data, io.buf)
	log.Logger.Infof("after filter output is :%v, index:%v", io.buf.Text(), io.buf.Document().CursorPositionCol())
}

func (io *IOFilter) handleASCIICodeBinding(b []byte) bool {
	checked := false
	for _, kb := range io.ASCIICodeBindings {
		if bytes.Compare(kb.ASCIICode, b) == 0 {
			kb.Fn(io.buf)
			checked = true
		}
	}
	return checked
}

func (io *IOFilter) handleKeyBinding(key prompt.Key) {
	for i := range prompt.CommonKeyBindings {
		kb := prompt.CommonKeyBindings[i]
		if kb.Key == key {
			kb.Fn(io.buf)
		}
	}

	for i := range io.keyBindings {
		kb := io.keyBindings[i]
		if kb.Key == key {
			kb.Fn(io.buf)
		}
	}
}

func IsShellPrompt(buffer []byte) bool {
	//if !strings.HasPrefix(string(buffer), "[") {
	//	return false
	//}
	if strings.Contains(string(buffer), AppSuffix) {
		return true
	} else if strings.Contains(string(buffer), RootSuffix) {
		return true
	}
	return false
}
