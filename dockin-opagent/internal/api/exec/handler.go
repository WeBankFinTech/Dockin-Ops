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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/common"
	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/server/remotecommand"
	"github.com/webankfintech/dockin-opagent/internal/server/streaming"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/webankfintech/dockin-opagent/internal/docker"
	"github.com/webankfintech/dockin-opagent/internal/log"
	"github.com/webankfintech/dockin-opagent/internal/model"
	api "github.com/webankfintech/dockin-opagent/internal/server/api/core"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	remoteapi "k8s.io/apimachinery/pkg/util/remotecommand"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type DockerHandler struct {
	dockerService *docker.DockerService
}

func NewDockerHandler(d *docker.DockerService) *DockerHandler {
	return &DockerHandler{
		dockerService: d,
	}
}

func (d *DockerHandler) SSHHandle(writer http.ResponseWriter, req *http.Request) {
	uid := uuid.New().String()
	log.Logger.Infof("receive ssh exec request, uid=%s", uid)

	var (
		opserverConn *websocket.Conn
		err          error
		resizeChan   = make(chan dockershim.TerminalSize)
	)

	if opserverConn, err = wsUpgrader.Upgrade(writer, req, nil); err != nil {
		log.Logger.Warnf("failed to upgrade connection to websocket, err=%v, uid=%s", err, uid)
		return
	}
	log.Logger.Infof("make connection with opserver success, uid=%s", uid)
	ctx, cancel := context.WithCancel(req.Context())
	wg := new(sync.WaitGroup)
	wg.Add(2)
	defer cancel()
	client := &Client{
		OpserverConn:    opserverConn,
		MsgFromOpserver: make(chan []byte, 1024),
		MsgFromDocker:   make(chan []byte, 1024),
		ResizeChan:      resizeChan,
		DockerService:   d.dockerService,
	}
	go client.OPServerMessageReadLoop(ctx, wg, uid)
	go client.OPServerMessageWriteLoop(ctx, wg, uid)
	wg.Wait()
}

func (d *DockerHandler) InteractiveHandle(writer http.ResponseWriter, req *http.Request) {
	uid := uuid.New().String()
	log.Logger.Infof("receive interactive exec request, uid=%s", uid)

	var (
		opserverConn *websocket.Conn
		err          error
		resizeChan   = make(chan dockershim.TerminalSize)
	)

	traceId := uuid.New().String()
	if opserverConn, err = wsUpgrader.Upgrade(writer, req, nil); err != nil {
		log.Logger.Warnf("failed to upgrade connection to websocket, err=%v", err)
		return
	}
	log.Logger.Infof("make connection with opserver success, traceId=%s", traceId)
	execParam, err := parseMessageFromHttpReq(req, uid)
	if err != nil {
		log.Logger.Warnf("failed to parse the body to ExecParam, err=%v, uid=%s", err, uid)
		writer.Write(model.NewErrorAgentResultWithCode(err, model.ErrParam).ToJSONByte())
		return
	}
	stream := &SocketStream{
		OpserverWsConn: opserverConn,
		TerminalSize:   resizeChan,
		UID:            uid,
	}
	ios := &dockershim.IOStreams{
		In:     stream,
		Out:    stream,
		ErrOut: stream,
	}

	log.Logger.Infof("send interactive command to docker, uid=%s", uid)
	// this will block the exec until the ssh connection stop
	insp, err := d.dockerService.Exec(req.Context(), uid, execParam.PodName, dockertypes.ExecConfig{
		User:         execParam.User,
		Privileged:   execParam.Privileged,
		Tty:          false,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          execParam.Env,
		WorkingDir:   execParam.WorkingDir,
		Cmd:          execParam.Cmd,
	}, ios, resizeChan)

	if err != nil {
		log.Logger.Warnf("exec interactive command in docker err=%v, inspect=%v, uid=%s", err, insp, uid)
		return
	}
	log.Logger.Infof("success exec interactive command in docker, uid=%s", uid)
}

func (d *DockerHandler) CommonHandle(writer http.ResponseWriter, req *http.Request) {
	uid := uuid.New().String()
	log.Logger.Infof("receive common exec request, uid=%s", uid)

	execParam, err := parseMessageFromHttpReq(req, uid)
	if err != nil {
		log.Logger.Warnf("failed to parse the body to ExecParam, err=%v, uid=%s", err, uid)
		writer.Write(model.NewErrorAgentResultWithCode(err, model.ErrParam).ToJSONByte())
		return
	}
	ios, _, bufout, buferr := dockershim.NewBufferIOStreams()

	log.Logger.Infof("send common command to docker, uid=%s", uid)
	insp, err := d.dockerService.Exec(req.Context(), uid, execParam.PodName, dockertypes.ExecConfig{
		User:         execParam.User,
		Privileged:   execParam.Privileged,
		Tty:          false,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          execParam.Env,
		WorkingDir:   execParam.WorkingDir,
		Cmd:          execParam.Cmd,
	}, ios, nil)
	if err != nil {
		log.Logger.Warnf("exec command in docker err=%v, stderr=%s, inspect=%v, uid=%s", err, buferr.String(), insp, uid)
		writer.Write(model.NewErrorAgentResultWithCode(err, model.ErrParam).ToJSONByte())
	}

	log.Logger.Infof("success exec comm and in docker, uid=%s", uid)
	writer.Write(model.NewSuccessAgentResult(bufout.String()).ToJSONByte())
}

func (d *DockerHandler) ServerExec(writer http.ResponseWriter, req *http.Request) {
	uid := uuid.New().String()
	log.Logger.Infof("receive ssh exec request, uid=%s", uid)

	if err := common.ValidateRequestV2(req, uid); err != nil {
		res := model.NewErrorAgentResult(err)
		writer.Write(res.ToJSONByte())
		log.Logger.Warnf("ValidateRequest err=%s", err.Error())
		return
	}
	execRequest, err := getExecRequest(req)
	if err != nil {
		log.Logger.Warnf("getExecRequest err=%s,uid=%s", err.Error(), uid)
		nerr := fmt.Sprintf("parse request param err=%s,uid=%s", err.Error(), uid)
		writer.Write([]byte(nerr))
		return
	}

	log.Logger.Infof("server exec request=%#v", execRequest)
	if err = validateExecRequest(execRequest); err != nil {
		log.Logger.Warnf("validateExecRequest err=%s,uid=%s", err.Error(), uid)
		nerr := fmt.Sprintf("valid request param err=%s,uid=%s", err.Error(), uid)
		writer.Write([]byte(nerr))
		return
	}

	/*
		containerId, err := d.dockerService.GetDockerIdByPodName(execRequest.ContainerId)
		if err != nil{
			log.Logger.Warnf("GetDockerIdByPodName err=%s",err.Error())
			nerr := fmt.Sprintf("get container by pod:%s err=%s,uid=%s",execRequest.ContainerId,err.Error(),uid)
			writer.Write([]byte(nerr))
			return
		}
	*/

	streamOpts := &remotecommand.Options{
		Stdin:  execRequest.Stdin,
		Stdout: execRequest.Stdout,
		Stderr: execRequest.Stderr,
		TTY:    execRequest.Tty,
	}

	log.Logger.Infof("streamOpts=%#v", streamOpts)
	log.Logger.Infof("exec param:containerId=%s,cmd=%v", execRequest.ContainerId, execRequest.Cmd)

	config := streaming.Config{
		StreamIdleTimeout:               15 * time.Minute,
		StreamCreationTimeout:           15 * time.Second,
		SupportedRemoteCommandProtocols: remoteapi.SupportedStreamingProtocols,
	}
	runtime := d.dockerService.GetClient()

	exeCfg := streaming.GetExecConfig(execRequest)

	s := streaming.NewServer(config, runtime)
	s.ServeExec(
		writer,
		req,
		execRequest.ContainerId,
		exeCfg,
		streamOpts,
		uid)
}

func getExecRequest(req *http.Request) (*streaming.ExecRequest, error) {
	streamOpts, err := remotecommand.NewOptions(req)
	if err != nil {
		log.Logger.Warnf("getExecRequestParams new streamOpts err=%s", err)
		return nil, err
	}

	query := req.URL.Query()
	container := query.Get("containerId")
	cmd := query[api.ExecCommandParam]
	user := query.Get(api.User)
	workDir := query.Get(api.WorkDir)
	env := query[api.Env]
	privileged := query.Get(api.Privileged) == "1"
	log.Logger.Infof("user=%s,workDir=%s,env=%s", user, workDir, env)
	return &streaming.ExecRequest{
		Stdin:       streamOpts.Stdin,
		Stdout:      streamOpts.Stdout,
		Stderr:      streamOpts.Stderr,
		Tty:         streamOpts.TTY,
		ContainerId: container,
		Cmd:         cmd,
		User:        user,
		Env:         env,
		WorkingDir:  workDir,
		Privileged:  privileged,
	}, nil
}

func validateExecRequest(req *streaming.ExecRequest) error {
	if req.ContainerId == "" {
		return status.Errorf(codes.InvalidArgument, "missing required container_id")
	}
	if req.Tty && req.Stderr {
		// If TTY is set, stderr cannot be true because multiplexing is not
		// supported.
		return status.Errorf(codes.InvalidArgument, "tty and stderr cannot both be true")
	}
	if !req.Stdin && !req.Stdout && !req.Stderr {
		return status.Errorf(codes.InvalidArgument, "one of stdin, stdout, or stderr must be set")
	}
	return nil
}
