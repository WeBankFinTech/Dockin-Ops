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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/config"

	"github.com/pkg/errors"

	"github.com/webankfintech/dockin-opserver/internal/log"
)

type FilterChan struct {
	filters []Filter
}

func NewFilterChan() *FilterChan {
	return &FilterChan{}
}

func (fc *FilterChan) AddFilter(cf Filter) {
	fc.filters = append(fc.filters, cf)
}

func (fc *FilterChan) Do(cmdLine string) error {
	cul, err := ParseCmdlineToCmdUnitList(cmdLine)
	if err != nil {
		log.Logger.Warnf("failed to parse the command:%s, err:%s", cmdLine, err)
		return nil
	}
	for _, f := range fc.filters {
		if err := f.Do(cul); err != nil {
			return err
		}
	}
	return nil
}

type FuncGetCandidates func() []string

type CommandFilter struct {
	candidates        []string
	funcGetCandidates FuncGetCandidates
	updateTicker      *time.Ticker
}

func (b *CommandFilter) updateCandidates() {
	log.Logger.Infof("start to updateCandidates")
	if b.funcGetCandidates == nil {
		return
	}
	b.candidates = b.funcGetCandidates()

	/*
		go func() {
			defer log.Logger.Infof("end to funcGetCandidates")
			select {
				case <-b.updateTicker.C:
										b.candidates = b.funcGetCandidates()
			}
		}()
	*/
}

type Filter interface {
	Do([]*CmdUnit) error
}

type WhitelistFilter struct {
	CommandFilter
}

func NewWhitelistFilter(funcs FuncGetCandidates) Filter {
	wl := &WhitelistFilter{
		CommandFilter: CommandFilter{
			candidates:        []string{},
			updateTicker:      time.NewTicker(time.Minute),
			funcGetCandidates: funcs,
		},
	}
	wl.updateCandidates()
	return wl
}

func (w *WhitelistFilter) Do(cua []*CmdUnit) error {
	log.Logger.Infof("check whitelist=%v", w.candidates)

	check := func(cmd string) bool {
		log.Logger.Infof("cmd=%s", cmd)
		for _, c := range w.candidates {
			if strings.EqualFold(strings.TrimSpace(cmd), c) {
				return true
			}
		}
		return false
	}

	for _, cu := range cua {
		if check(cu.cmd) {
			continue
		} else {
			return fmt.Errorf("command [%s] is not allowd to exec, please contact Dockin_helper for help.", cu.cmd)
		}
	}

	return nil
}

type BlacklistFilter struct {
	CommandFilter
}

func NewBlacklistFilter(funcs FuncGetCandidates) Filter {
	bl := &BlacklistFilter{
		CommandFilter: CommandFilter{
			candidates:        []string{},
			updateTicker:      time.NewTicker(time.Minute),
			funcGetCandidates: funcs,
		},
	}
	bl.updateCandidates()
	return bl
}

func (w *BlacklistFilter) Do(cua []*CmdUnit) error {
	log.Logger.Infof("check blacklist=%v", w.candidates)

	for _, cu := range cua {
		log.Logger.Infof("cmd=%s", cu.cmd)
		for _, c := range w.candidates {
			if strings.EqualFold(strings.TrimSpace(cu.cmd), c) {
				return fmt.Errorf("command [%s] is not allowd to exec, please contact Dockin_helper for help.", cu.cmd)
			}
		}
	}

	return nil
}

type FuncGetCurrentWorkDir func() string

type VIMFilter struct {
	Executor   Executor
	Limit      int64
	CurrentDir FuncGetCurrentWorkDir
	Container  string
}

func NewVimFilter(container string, executor Executor, funs FuncGetCurrentWorkDir) Filter {
	vf := &VIMFilter{
		Executor:   executor,
		Limit:      config.OpsConfig.Limits.ViFileMaxSize * 1024 * 1024,
		CurrentDir: funs,
		Container:  container,
	}
	return vf
}

func (vf *VIMFilter) Do(cua []*CmdUnit) error {
	log.Logger.Infof("check vi")
	var (
		targetFile string
	)
	isViCmd := false
	for _, cu := range cua {
		log.Logger.Infof("cmd=%s", cu.cmd)
		if strings.EqualFold(cu.cmd, "vi") || strings.EqualFold(cu.cmd, "vim") {
			isViCmd = true
			targetFile = cu.target
		}
	}
	if !isViCmd {
		return nil
	}

	ios, _, stdout, stderr := NewIOStreams()
	err := vf.Executor.Exec(&DockinExecParam{
		Cmd:  []string{"du", "-shb", targetFile},
		User: "root",
		//WorkDir: vf.CurrentDir(),
		WorkDir:       vf.CurrentDir(),
		ContainerName: vf.Container,
	}, ios)
	if err != nil {
		log.Logger.Warnf("failed to vi file as stderr:%s, err:%v", stderr.String(), err)
		stderrstr := stderr.String()
		if strings.HasSuffix(stderrstr, "\n") {
			stderrstr = strings.TrimSuffix(stderrstr, "\n") + ", if "
		}
		return errors.New(stderrstr)
	}

	size, err := strconv.ParseInt(strings.Split(stdout.String(), "\t")[0], 10, 64)
	if err != nil {
		log.Logger.Warnf("failed to parse the stdout to file size, stdout:%s, err:%s", stdout.String(), err)
		return errors.Errorf("ParseInt vi file=%s size failed", targetFile)
	}
	if size > vf.Limit {
		log.Logger.Infof("the file edited is larger than %d, and it's size to %d", vf.Limit, size)
		return fmt.Errorf("the file edited is larger than %d, and it's size to %d", vf.Limit, size)
	}

	return nil
}
