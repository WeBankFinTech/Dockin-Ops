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
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/log"
)

var (
	NoSuchFileOrDirectory    = "No such file or directory"
	ErrNoSuchFileOrDirectory = fmt.Errorf("No such file or directory")

	AppUserDir  = "/data/app"
	RootUserDir = "/root"
)

type ShellPrompt struct {
	user     string
	instance string
	dir      string
}

func (sp *ShellPrompt) String() string {
	return fmt.Sprintf("user:%s, podName:%s,dir:%s", sp.user, sp.instance, sp.dir)
}

func ParserShellPrompt(content string) (*ShellPrompt, error) {
	if !IsShellPrompt([]byte(content)) {
		return nil, fmt.Errorf("not a validate shell prompt")
	}
	index := strings.Index(content, "]")
	if index == -1 {
		return nil, fmt.Errorf("not right ] exist")
	}
	buf := content[1:index]
	indexAt := strings.Index(buf, "@")
	if indexAt == -1 {
		return nil, fmt.Errorf("no @ exist in shell prompt")
	}

	sp := &ShellPrompt{}
	sp.user = buf[:indexAt]
	buf = buf[indexAt+1:]

	indexSpace := strings.Index(buf, " ")
	if indexSpace == -1 {
		return nil, fmt.Errorf("no space exist in shell prompt")
	}
	sp.instance = buf[:indexSpace]
	buf = buf[indexSpace+1:]

	if len(buf) == 0 {
		return nil, fmt.Errorf("no work dir exist in shell prompt")
	}
	sp.dir = buf

	return sp, nil
}

type SSHContext struct {
	LastInputRune         func() rune
	CurrentRunningCommand string
	CommandFinished       bool
	WorkingDir            string
	LastDir               string
}

func NewSSHContext() *SSHContext {
	sshc := &SSHContext{}
	//sshc.LastInputRune = input.LastInputRune
	sshc.WorkingDir = "/data"
	return sshc
}

func (s *SSHContext) setEnvironment(output string) {
	if IsShellPrompt([]byte(output)) {
		if ok, cu := isCd(s.CurrentRunningCommand); ok {
			sp, err := ParserShellPrompt(output)
			if err != nil {
				return
			}
			log.Logger.Infof("update environment shell prompt:%s", sp.String())
			s.setCurrentWorkingDir(sp, cu.target)
		}
	}
}

func isCd(cmd string) (bool, *CmdUnit) {
	unitList, err := ParseCmdlineToCmdUnitList(cmd)
	if err != nil {
		return false, nil
	}

	for _, c := range unitList {
		if c.cmd == "cd" {
			return true, c
		}
	}
	return false, nil
}

func isExecSuccess(output string) bool {
	return !strings.Contains(output, NoSuchFileOrDirectory)
}

func (s *SSHContext) setCurrentWorkingDir(ps *ShellPrompt, target string) {
	if ps.dir == "~" {
		s.LastDir = s.WorkingDir
		if ps.user == "app" {
			s.WorkingDir = AppUserDir
		} else {
			s.WorkingDir = RootUserDir
		}
		log.Logger.Infof("~, update current dir from [%s] to [%s], click [%s]", s.LastDir, s.WorkingDir, target)
		return
	}
	if ps.dir == "/" {
		s.LastDir = s.WorkingDir
		s.WorkingDir = "/"
		log.Logger.Infof("/, update current dir from [%s] to [%s], click [%s]", s.LastDir, s.WorkingDir, target)
		return
	}

	if target == "-" {
		s.LastDir, s.WorkingDir = s.WorkingDir, s.LastDir
		log.Logger.Infof("-, update current dir from [%s] to [%s], click [%s]", s.LastDir, s.WorkingDir, target)
		return
	}

	maybe := getAbsolutePath(s.WorkingDir, target)
	if strings.HasSuffix(maybe, "/") {
		maybe = maybe[:len(maybe)-1]
	}
	if strings.HasSuffix(maybe, ps.dir) {
		s.LastDir = s.WorkingDir
		s.WorkingDir = maybe
		log.Logger.Infof("has suffix, update current dir from [%s] to [%s], click [%s]", s.LastDir, s.WorkingDir, target)
	}
}

func (s *SSHContext) getCurrentWorkingDir() string {
	return s.WorkingDir
}

func (sc *SSHContext) setCurrentCommand(data string) {
	sc.CurrentRunningCommand = data
}

func getAbsolutePath(basePath, givenPath string) string {
	if givenPath == "" || givenPath == "~" || givenPath == "." {
		log.Logger.Infof("given path =%s is empty or ~, return the base path %s", givenPath, basePath)
		return basePath
	}
	if strings.HasPrefix(givenPath, "/") {
		log.Logger.Infof("given path is a absolute path, return the given path %s", givenPath)
		return givenPath
	}
	if strings.HasPrefix(givenPath, "..") && len(givenPath) > 2 {
		lastIndex := strings.LastIndex(givenPath, "..")
		right := givenPath[lastIndex+2:]
		cnt := strings.Count(givenPath, "..")
		for i := 0; i < cnt; i++ {
			basePath = getParentDirectory(basePath)
		}

		if right != "" {
			basePath = basePath + "/" + right
		}
		log.Logger.Infof("given path need to get parent path, result is %s, given path %s", basePath, givenPath)
		return basePath
	}
	if strings.HasPrefix(givenPath, ".") && len(givenPath) > 2 {
		log.Logger.Infof("given path is the current path, return the current base path %s", givenPath)
		return basePath + "/" + givenPath[2:]
	}

	return basePath + "/" + givenPath
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(directory string) string {
	if strings.TrimSpace(directory) == "" {
		return "."
	}

	if strings.Count(directory, "/") == 0 {
		return "."
	}
	index := strings.LastIndex(directory, "/")
	if index == -1 {
		return strings.TrimSpace(directory)
	}
	sub := substr(directory, 0, index)
	if sub == "" {
		return "/"
	}
	return sub
}
