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

package exec

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/remote"

	"github.com/gorilla/websocket"
)

var (
	ErrNoPodsInfo       = fmt.Errorf("could not get pod info")
	ErrNoContainerExec  = fmt.Errorf("could not find container to exec")
	DefaultTimeout      = time.Duration(15 * time.Minute)
	DefaultNoTtyTimeout = time.Duration(5 * time.Second)
	forbiddenExecCmd    []string
)

func init() {
	forbiddenExecCmd = config.OpsConfig.Limits.ExecForbidden
	log.Logger.Infof("load forbidden command %#v", forbiddenExecCmd)
}

type ExecCommand struct {
	OpsOpts *model.OpsOption
	Conn    *websocket.Conn
	HostIp  string
}

func (e *ExecCommand) RunWithTty(traceId string) error {
	log.Logger.Infof("start to RunWithTty")
	execParam := remote.OpsOption2ExecParam(e.OpsOpts)
	execParam.HostIP = e.HostIp

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	session, err := remote.CreateExecSession(cancelCtx, execParam, e.Conn, remote.InteractExecMode)
	if err != nil {
		log.Logger.Warnf("failed to create interactive exec session, err=%v, traceId=%s", err, traceId)
		remote.HandleWSError(e.Conn, err)
		return err
	}
	session.Start(cancelCtx, traceId)
	if err := session.Executor.ExecInteractive(execParam, session.InterStream); err != nil {
		log.Logger.Warnf("run interactive exec err:%v, traceId=%s", err, traceId)
		//remote.HandleWSError(e.Conn, err)
		return err
	}

	return nil
}

func (e *ExecCommand) RunNoTty(traceId string, ioStream *remote.IOStreams) error {
	execParam := remote.OpsOption2ExecParam(e.OpsOpts)
	execParam.HostIP = e.HostIp

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	session, err := remote.CreateExecSession(cancelCtx, execParam, e.Conn, remote.CommonExecMode)
	if err != nil {
		log.Logger.Warnf("failed to create common exec session, err=%v, traceId=%s", err, traceId)
		return err
	}

	//session.Start(cancelCtx)
	err = session.Executor.Exec(execParam, ioStream)
	if err != nil {
		log.Logger.Warnf("run common exec err:%v, stderr=%s, traceId=%s", err, ioStream.ErrOut.(*bytes.Buffer).String(), traceId)
		return err
	}

	return nil
}
