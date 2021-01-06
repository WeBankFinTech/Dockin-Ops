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
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/webankfintech/dockin-opserver/internal/model"

	"github.com/webankfintech/dockin-opserver/internal/log"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type DockerExecutor struct {
	ts      chan *remotecommand.TerminalSize
	Address string
	Port    int32
}

func NewDockerExecutor(address string, port int32) Executor {
	return &DockerExecutor{
		Address: address,
		Port:    port,
		ts:      make(chan *remotecommand.TerminalSize),
	}
}

func (de *DockerExecutor) Exec(execParam *DockinExecParam, streams *IOStreams) error {
	uri, err := url.Parse(fmt.Sprintf("http://%s:%d/dockin/opagent/exec/serverexec", de.Address, de.Port))
	if err != nil {
		return err
	}

	log.Logger.Infof("Exec opagent url=http://%s%s", uri.Host, uri.Path)

	params := url.Values{}
	//params.Add("input", "1")
	params.Add("output", "1")
	params.Add("error", "1")
	execParam.Env = append(execParam.Env, "LANG=en_US.utf8")
	execParam.Env = append(execParam.Env, "TERM=xterm")
	execParam.Env = append(execParam.Env, "LC_ALL=zh_CN.UTF-8")
	for _, e := range execParam.Env {
		params.Add("env", e)
	}
	params.Add("user", execParam.User)
	params.Add("workDir", execParam.WorkDir)
	for _, c := range execParam.Cmd {
		params.Add("command", c)
	}
	params.Add("containerId", execParam.ContainerName)
	params.Add("access-token", model.OpagentAccessToken())
	uri.RawQuery = params.Encode()

	exec, err := remotecommand.NewSPDYExecutor(&restclient.Config{}, "POST", uri)
	if err != nil {
		log.Logger.Warnf("create spdy executor with opagent failed %v", err)
		return err
	}

	if err := exec.Stream(remotecommand.StreamOptions{
		Stderr:            streams.ErrOut,
		Stdout:            streams.Out,
		Tty:               false,
		TerminalSizeQueue: de,
	}); err != nil {
		log.Logger.Warnf("stream with opagent failed %v", err)
		return err
	}
	return nil
}

func (de *DockerExecutor) ExecInteractive(execParam *DockinExecParam, cmdStream *InteractStream) error {

	uri, err := url.Parse(fmt.Sprintf("http://%s:%d/dockin/opagent/exec/serverexec", de.Address, de.Port))
	if err != nil {
		return err
	}

	log.Logger.Infof("ExecInteractive opagent url=http://%s%s", uri.Host, uri.Path)

	params := url.Values{}
	params.Add("input", "1")
	params.Add("output", "1")
	params.Add("error", "1")
	if execParam.TTY {
		params.Add("tty", "1")
	} else {
		params.Add("tty", "0")
	}

	execParam.Env = append(execParam.Env, "LANG=en_US.utf8")
	execParam.Env = append(execParam.Env, "TERM=xterm")
	execParam.Env = append(execParam.Env, "LC_ALL=zh_CN.UTF-8")
	execParam.Env = append(execParam.Env, "LESSOPEN=||/usr/bin/lesspipe.sh %s")

	for _, e := range execParam.Env {
		params.Add("env", e)
	}
	params.Add("user", execParam.User)
	params.Add("workDir", execParam.WorkDir)
	for _, c := range execParam.Cmd {
		params.Add("command", c)
	}
	params.Add("containerId", execParam.ContainerName)
	params.Add("access-token", model.OpagentAccessToken())
	uri.RawQuery = params.Encode()

	exec, err := remotecommand.NewSPDYExecutor(&restclient.Config{}, "POST", uri)
	if err != nil {
		log.Logger.Warnf("create spdy executor with opagent failed %v", err)
		return err
	}

	go func() {
		defer log.Logger.Infof("end chan TerminalSize")
		de.ts <- &remotecommand.TerminalSize{
			Width:  uint16(execParam.Width),
			Height: uint16(execParam.Height),
		}
		log.Logger.Infof("after chan TerminalSize")
	}()

	if err := exec.Stream(remotecommand.StreamOptions{
		Stdin:             cmdStream,
		Stderr:            cmdStream,
		Stdout:            cmdStream,
		Tty:               execParam.TTY,
		TerminalSizeQueue: de,
	}); err != nil {
		log.Logger.Warnf("stream with opagent failed %v", err)
		return err
	}

	log.Logger.Infof("end to ExecInteractive")
	return nil
}

func (de *DockerExecutor) Resize(width, height int) error {
	de.ts <- &remotecommand.TerminalSize{
		Width:  uint16(width),
		Height: uint16(height),
	}
	return nil
}

func (de *DockerExecutor) Shell(ctx context.Context, execParam *DockinExecParam, cmdStream *InteractStream) error {
	uri, err := url.Parse(fmt.Sprintf("http://%s:%d/dockin/opagent/exec/serverexec", de.Address, de.Port))
	if err != nil {
		return err
	}

	log.Logger.Infof("Shell opagent url=http://%s%s", uri.Host, uri.Path)
	params := url.Values{}
	params.Add("input", "1")
	params.Add("output", "1")
	params.Add("error", "1")
	params.Add("tty", "1")
	params.Add("command", "/bin/bash")
	params.Add("containerId", execParam.ContainerName)
	params.Add("access-token", model.OpagentAccessToken())

	execParam.Env = append(execParam.Env, "TERM=xterm")
	execParam.Env = append(execParam.Env, "LANG=en_US.utf8")
	execParam.Env = append(execParam.Env, "LC_ALL=zh_CN.UTF-8")
	for _, e := range execParam.Env {
		params.Add("env", e)
	}
	params.Add("workDir", execParam.WorkDir)
	uri.RawQuery = params.Encode()

	exec, err := remotecommand.NewSPDYExecutor(&restclient.Config{}, "POST", uri)
	if err != nil {
		log.Logger.Warnf("create spdy executor with opagent failed %v", err)
		return err
	}
	go func() {
		de.ts <- &remotecommand.TerminalSize{
			Width:  uint16(execParam.Width),
			Height: uint16(execParam.Height),
		}
	}()
	if err := exec.Stream(remotecommand.StreamOptions{
		Stdin:             cmdStream,
		Stderr:            cmdStream,
		Stdout:            cmdStream,
		Tty:               true,
		TerminalSizeQueue: de,
	}); err != nil {
		log.Logger.Warnf("stream with opagent failed %v", err)
		return err
	}
	log.Logger.Infof("end to Shell")
	return nil
}

func (de *DockerExecutor) DebugShell(ctx context.Context, execParam *DockinExecParam, cmdStream *InteractStream) error {
	uri, err := url.Parse(fmt.Sprintf("http://%s:%d/dockin/opagent/debug", de.Address, de.Port))
	if err != nil {
		return err
	}

	log.Logger.Infof("DebugShell opagent url=http://%s%s", uri.Host, uri.Path)
	params := url.Values{}
	params.Add("image", execParam.Image)
	params.Add("podName", execParam.PodName)
	params.Add("containerId", execParam.ContainerName)
	bytes, err := json.Marshal(execParam.Cmd)
	if err != nil {
		return err
	}

	params.Add("command", string(bytes))
	params.Add("access-token", model.OpagentAccessToken())
	uri.RawQuery = params.Encode()

	exec, err := remotecommand.NewSPDYExecutor(&restclient.Config{}, "POST", uri)
	if err != nil {
		log.Logger.Warnf("create spdy executor with opagent failed %v", err)
		return err
	}
	go func() {
		de.ts <- &remotecommand.TerminalSize{
			Width:  uint16(execParam.Width),
			Height: uint16(execParam.Height),
		}
	}()

	if err := exec.Stream(remotecommand.StreamOptions{
		Stdin:             cmdStream,
		Stderr:            cmdStream,
		Stdout:            cmdStream,
		Tty:               true,
		TerminalSizeQueue: de,
	}); err != nil {
		log.Logger.Warnf("stream with opagent failed %v", err)
		return err
	}

	log.Logger.Infof("end to DebugShell")
	return nil
}

func (de *DockerExecutor) Next() *remotecommand.TerminalSize {
	select {
	case ts, ok := <-de.ts:
		if !ok {
			log.Logger.Infof("terminal size change close")
			return nil
		}
		return ts
	}
}
