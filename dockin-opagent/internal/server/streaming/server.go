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

package streaming

import (
	"context"
	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/log"
	"io"
	"net/http"
	"time"

	remotecommandserver "github.com/webankfintech/dockin-opagent/internal/server/remotecommand"

	dockertypes "github.com/docker/docker/api/types"
	"k8s.io/apimachinery/pkg/types"
)

type Runtime interface {
	//Exec(containerID string, cmd []string, in io.Reader, out, err io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize) error
	Exec(ctx context.Context, traceId, containerId, podName string, execCfg dockertypes.ExecConfig, iostream *dockershim.IOStreams, resize <-chan dockershim.TerminalSize) (*dockertypes.ContainerExecInspect, error)
	Attach(containerID string, iostream *dockershim.IOStreams, tty bool, resize <-chan dockershim.TerminalSize, uid string) error
}

type Config struct {
	// How long to leave idle connections open for.
	StreamIdleTimeout time.Duration
	// How long to wait for clients to create streams. Only used for SPDY streaming.
	StreamCreationTimeout time.Duration

	// The streaming protocols the server supports (understands and permits).  See
	// k8s.io/kubernetes/pkg/kubelet/server/remotecommand/constants.go for available protocols.
	// Only used for SPDY streaming.
	SupportedRemoteCommandProtocols []string

	// The streaming protocols the server supports (understands and permits).  See
	// k8s.io/kubernetes/pkg/kubelet/server/portforward/constants.go for available protocols.
	// Only used for SPDY streaming.
	SupportedPortForwardProtocols []string
}

type ExecRequest struct {
	// ID of the container in which to execute the command.
	ContainerId string `protobuf:"bytes,1,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	// Command to execute.
	Cmd []string `protobuf:"bytes,2,rep,name=cmd" json:"cmd,omitempty"`
	// Whether to exec the command in a TTY.
	Tty bool `protobuf:"varint,3,opt,name=tty,proto3" json:"tty,omitempty"`
	// Whether to stream stdin.
	// One of `stdin`, `stdout`, and `stderr` MUST be true.
	Stdin bool `protobuf:"varint,4,opt,name=stdin,proto3" json:"stdin,omitempty"`
	// Whether to stream stdout.
	// One of `stdin`, `stdout`, and `stderr` MUST be true.
	Stdout bool `protobuf:"varint,5,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// Whether to stream stderr.
	// One of `stdin`, `stdout`, and `stderr` MUST be true.
	// If `tty` is true, `stderr` MUST be false. Multiplexing is not supported
	// in this case. The output of stdout and stderr will be combined to a
	// single stream.
	Stderr     bool     `protobuf:"varint,6,opt,name=stderr,proto3" json:"stderr,omitempty"`
	User       string   // User that will run the command
	WorkingDir string   `json:"workingDir"` // Working directory
	Env        []string `json:"env"`        // Environment variables
	Privileged bool     `json:"privileged"` // Is the container in privileged mode
}

type Server struct {
	runtime *criAdapter
	config  Config
}

func NewServer(config Config, runtime Runtime) *Server {
	return &Server{
		config:  config,
		runtime: &criAdapter{runtime},
	}
}

func (s *Server) ServeExec(writer http.ResponseWriter, req *http.Request, container string, execCfg dockertypes.ExecConfig, streamOpts *remotecommandserver.Options, uid string) {
	log.Logger.Infof("start to ServeExec")
	remotecommandserver.ServeExec(
		writer,
		req,
		s.runtime,
		uid,
		container,
		execCfg,
		streamOpts,
		s.config.StreamIdleTimeout,
		s.config.StreamCreationTimeout,
		s.config.SupportedRemoteCommandProtocols)
}

func (s *Server) ServeAttach(writer http.ResponseWriter, req *http.Request, container string, execCfg dockertypes.ExecConfig, streamOpts *remotecommandserver.Options, uid string) {
	log.Logger.Infof("start to ServeAttach")

	remotecommandserver.ServeAttach(
		writer,
		req,
		s.runtime,
		container,
		streamOpts,
		s.config.StreamIdleTimeout,
		s.config.StreamCreationTimeout,
		s.config.SupportedRemoteCommandProtocols,
		uid)
}

type Attacher interface {
	// AttachContainer attaches to the running container in the pod, copying data between in/out/err
	// and the container's stdin/stdout/stderr.
	AttachContainer(name string, uid types.UID, container string, in io.Reader, out, err io.WriteCloser, tty bool, resize <-chan dockershim.TerminalSize) error
	CleanContainer(id string)
}

type criAdapter struct {
	Runtime
}

var _ remotecommandserver.Executor = &criAdapter{}
var _ remotecommandserver.Attacher = &criAdapter{}

func (a *criAdapter) ExecInContainer(ctx context.Context, podName string, podUID types.UID, traceId string, container string, execCfg dockertypes.ExecConfig, iostream *dockershim.IOStreams, resize <-chan dockershim.TerminalSize) (*dockertypes.ContainerExecInspect, error) {
	log.Logger.Infof("start to ExecInContainer")
	return a.Runtime.Exec(ctx, traceId, container, podName, execCfg, iostream, resize)
}

func (a *criAdapter) AttachContainer(container string, iostream *dockershim.IOStreams, tty bool, resize <-chan dockershim.TerminalSize, uid string) error {
	log.Logger.Infof("start to AttachContainer")
	return a.Runtime.Attach(container, iostream, tty, resize, uid)
}

func GetExecConfig(req *ExecRequest) dockertypes.ExecConfig {
	exeCfg := dockertypes.ExecConfig{
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Privileged:   req.Privileged,
		Cmd:          req.Cmd,
	}
	if req.User != "" {
		exeCfg.User = req.User
	}

	if req.WorkingDir != "" {
		exeCfg.WorkingDir = req.WorkingDir
	}

	if len(req.Env) != 0 {
		exeCfg.Env = req.Env
	}
	return exeCfg
}
