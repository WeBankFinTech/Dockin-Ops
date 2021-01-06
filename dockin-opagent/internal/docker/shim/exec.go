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

package dockershim

import (
	"fmt"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/docker/shim/libdocker"
	"github.com/webankfintech/dockin-opagent/internal/log"
	dockertypes "github.com/docker/docker/api/types"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type NativeExecHandler struct{}

func (n *NativeExecHandler) ExecInContainer(client libdocker.Interface,
	containerId, traceId, podName string,
	createOpts dockertypes.ExecConfig,
	iostream *IOStreams,
	resize <-chan TerminalSize) (*dockertypes.ContainerExecInspect, error) {
	done := make(chan struct{})
	defer close(done)
	createOpts.AttachStdin = iostream.In != nil
	createOpts.AttachStdout = iostream.Out != nil
	createOpts.AttachStderr = iostream.ErrOut != nil

	execObj, err := client.CreateExec(containerId, createOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to exec in container - Exec setup failed - %v", err)
	}

	// Have to start this before the call to client.StartExec because client.StartExec is a blocking
	// call :-( Otherwise, resize events don't get processed and the terminal never resizes.
	//
	// We also have to delay attempting to send a terminal resize request to docker until after the
	// exec has started; otherwise, the initial resize request will fail.
	execStarted := make(chan struct{})
	go func() {
		select {
		case <-execStarted:
			// client.StartExec has started the exec, so we can start resizing
		case <-done:
			// ExecInContainer has returned, so short-circuit
			return
		}

		n.HandleResizing(resize, func(size TerminalSize) {
			client.ResizeExecTTY(execObj.ID, uint(size.Height), uint(size.Width))
		})
	}()

	startOpts := dockertypes.ExecStartCheck{Detach: false, Tty: createOpts.Tty}
	streamOpts := libdocker.StreamOptions{
		InputStream:  iostream.In,
		OutputStream: iostream.Out,
		ErrorStream:  iostream.ErrOut,
		RawTerminal:  createOpts.Tty,
		ExecStarted:  execStarted,
	}
	err = client.StartExec(execObj.ID, startOpts, streamOpts)
	if err != nil {
		log.Logger.Warnf("StartExec err=%s",err.Error())
		return nil, err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	count := 0
	inspect := &dockertypes.ContainerExecInspect{}
	for {
		insp, err2 := client.InspectExec(execObj.ID)
		inspect = insp
		if err2 != nil {
			return inspect, err2
		}
		if !inspect.Running {
			if inspect.ExitCode != 0 {
				err = fmt.Errorf("err executing in Docker Container: %d", inspect.ExitCode)
			}
			break
		}

		count++
		if count == 5 {
			log.Logger.Warnf("Exec session %s in container %s terminated but process still running!", execObj.ID, containerId)
			break
		}

		<-ticker.C
	}

	log.Logger.Infof("end to ExecInContainer,traceId=%s",traceId)
	return inspect, err
}

func (n *NativeExecHandler)HandleResizing(resize <-chan TerminalSize, resizeFunc func(size TerminalSize)) {
	if resize == nil {
		return
	}

	go func() {
		defer runtime.HandleCrash()

		for size := range resize {
			if size.Height < 1 || size.Width < 1 {
				continue
			}
			resizeFunc(size)
		}
	}()
}
