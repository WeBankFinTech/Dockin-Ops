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

package remotecommand

import (
	"fmt"
	"net/http"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/log"

	srccontext "context"

	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"

	dockertypes "github.com/docker/docker/api/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	remotecommandconsts "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilexec "k8s.io/utils/exec"
)

type Executor interface {
	// ExecInContainer executes a command in a container in the pod, copying data
	// between in/out/err and the container's stdin/stdout/stderr.
	ExecInContainer(ctx srccontext.Context, podName string, podUID types.UID, uid string, container string, execCfg dockertypes.ExecConfig, iostream *dockershim.IOStreams, resize <-chan dockershim.TerminalSize) (*dockertypes.ContainerExecInspect, error) //ExecInContainer(name string, uid types.UID, container string, cmd []string, in io.Reader, out, err io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize, timeout time.Duration) (*dockertypes.ContainerExecInspect, error
}

func ServeExec(w http.ResponseWriter, req *http.Request, executor Executor, uid string, container string, execCfg dockertypes.ExecConfig, streamOpts *Options, idleTimeout, streamCreationTimeout time.Duration, supportedProtocols []string) {
	ctx, ok := createStreams(req, w, streamOpts, supportedProtocols, idleTimeout, streamCreationTimeout)
	if !ok {
		// error is handled by createStreams
		log.Logger.Warnf("createStreams failed")
		w.Write([]byte("createStreams failed"))
		return
	}
	defer ctx.conn.Close()

	ioStream := &dockershim.IOStreams{
		In:     ctx.stdinStream,
		Out:    ctx.stdoutStream,
		ErrOut: ctx.stderrStream,
	}

	execCfg.Tty = ctx.tty
	_, err := executor.ExecInContainer(req.Context(), "", "", uid, container, execCfg, ioStream, ctx.resizeChan)
	if err != nil {
		if exitErr, ok := err.(utilexec.ExitError); ok && exitErr.Exited() {
			log.Logger.Warnf("ExecInContainer err=%s", exitErr.Error())
			rc := exitErr.ExitStatus()
			ctx.writeStatus(&apierrors.StatusError{ErrStatus: metav1.Status{
				Status: metav1.StatusFailure,
				Reason: remotecommandconsts.NonZeroExitCodeReason,
				Details: &metav1.StatusDetails{
					Causes: []metav1.StatusCause{
						{
							Type:    remotecommandconsts.ExitCodeCauseType,
							Message: fmt.Sprintf("%d", rc),
						},
					},
				},
				Message: fmt.Sprintf("command terminated with non-zero exit code: %v", exitErr),
			}})
		} else {
			err = fmt.Errorf("error executing command in container: %v", err)
			log.Logger.Warnf(err.Error())
			runtime.HandleError(err)
			ctx.writeStatus(apierrors.NewInternalError(err))
		}
	} else {
		ctx.writeStatus(&apierrors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusSuccess,
		}})
	}
}
