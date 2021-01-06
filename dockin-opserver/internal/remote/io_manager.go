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
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/remote/prompt"
)

type SSHIOManager struct {
	filterChan *FilterChan
	sshContext *SSHContext
	inputMode  InputMode
	ioFilter   *IOFilter
}

func NewSSHIOManager(ctx *SSHContext, ioFilter *IOFilter) *SSHIOManager {
	sm := &SSHIOManager{
		sshContext: ctx,
		inputMode:  CollectMode,
		filterChan: NewFilterChan(),
		ioFilter:   ioFilter,
	}

	return sm
}

func (im *SSHIOManager) AddFilter(cf Filter) {
	im.filterChan.AddFilter(cf)
}

func (im *SSHIOManager) OnInput(data []byte) error {
	im.ioFilter.handleInput(data)
	if im.inputMode == CollectMode {
		switch im.ioFilter.LastCtrlKey {
		case prompt.Enter, prompt.ControlJ:
			err := im.handleCommand()
			log.Logger.Infof("after handle the command:%v", err)
			im.inputMode = InteractMode
			return err
		case prompt.ControlC:
			im.ioFilter.buf = prompt.NewBuffer()
			log.Logger.Infof("after ControlC, clear the buffer")
		}
	}

	return nil
}

func (im *SSHIOManager) handleCommand() error {
	cmd := im.ioFilter.buf.Text()
	im.ioFilter.buf = prompt.NewBuffer()
	log.Logger.Infof("handle command:%s", cmd)

	im.sshContext.setCurrentCommand(cmd)

	if strings.TrimSpace(cmd) == "" {
		return nil
	}
	if im.IsExitCmd(cmd) {
		return nil
	}
	if err := im.filterChan.Do(cmd); err != nil {
		log.Logger.Infof(err.Error())
		return err
	}

	return nil
}

func (im *SSHIOManager) IsExitCmd(cmd string) bool {
	return cmd == "exit" || cmd == "quit"
}

func (im *SSHIOManager) OnOutput(buffer []byte) {
	if im.inputMode == CollectMode {
		im.ioFilter.handleOutput(buffer)
	}

	im.updateCommandExecFinished(buffer)

	switch im.ioFilter.LastCtrlKey {
	case prompt.Enter, prompt.ControlC, prompt.ControlJ:
		im.sshContext.setEnvironment(string(buffer))
	}

	if im.sshContext.CommandFinished {
		log.Logger.Infof("command exec finished, set to CollectMode")
		im.inputMode = CollectMode
		//im.ioFilter.buf = prompt.NewBuffer()
	}
}

func (im *SSHIOManager) updateCommandExecFinished(buffer []byte) {
	im.sshContext.CommandFinished = IsShellPrompt(buffer)
}
