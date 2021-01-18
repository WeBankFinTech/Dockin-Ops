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

package ssh

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/tabwriter"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/remote"
	"github.com/webankfintech/dockin-opserver/internal/utils/ip"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Ssh struct {
	Cm          *client.Manager
	RedisClient *redis.RedisClient
}

func NewSsh(cm *client.Manager, rs *redis.RedisClient) *Ssh {
	ssh := &Ssh{
		Cm:          cm,
		RedisClient: rs,
	}
	http.HandleFunc("/v1/dockin/opserver/ssh-v2", ssh.HandleV2)
	return ssh
}

func (s *Ssh) HandleV2(writer http.ResponseWriter, req *http.Request) {
	var (
		err     error
		opsOpts *model.OpsOption
		pod     *v1.Pod
	)

	traceId := trace.TraceID()
	log.Logger.Infof("recv ssh-v2 request,traceId=%s", traceId)
	if opsOpts, err = api.ValidateExecRequest(req, traceId); err != nil {
		log.Logger.Warnf("failed to validate the exec param, as=%v, traceId=%s", err, traceId)
		writer.Write([]byte("failed to validate the exec param, as:" + err.Error() + traceId))
		return
	}

	conn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Logger.Warnf("failed to update the connection, err:%v, traceId=%s", err, traceId)
		remote.HandleWSError(conn, err)
		return
	}
	log.Logger.Infof("success to Upgrade webSocket protocol traceId=%s", traceId)

	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	reqIp := ip.GetIp(req)
	pod, err = api.GetPodStructFromRedis(opsOpts.Name, s.RedisClient)
	if err != nil {
		log.Logger.Warnf("failed to get pod struct from redis,podName=%s,err=%s traceId=%s", opsOpts.Name, err, traceId)
		remote.HandleWSError(conn, err)
		return
	}

	hostIp, err := api.GetHostIpByPod(opsOpts, pod, s.Cm, reqIp, traceId)
	if err != nil {
		log.Logger.Warnf("failed to get the host ip for ops=%s, traceId=%s", opsOpts.String(), traceId)
		remote.HandleWSError(conn, err)
		return
	}

	cid, err := api.GetContainerIdByPod(opsOpts.Name, pod)
	if err != nil {
		remote.HandleWSError(conn, errors.Wrapf(err, "get containerId from pod struct by pod=%s failed,err=%s", opsOpts.Name, err.Error()))
		return
	}

	opsOpts.Container = cid
	execParam := remote.OpsOption2ExecParam(opsOpts)
	execParam.HostIP = hostIp

	session, err := remote.CreateExecSession(cancelCtx, execParam, conn, remote.SSHExecMode)
	if err != nil {
		log.Logger.Warnf("failed to create a docker session, err:%v, traceId=%s", err, traceId)
		remote.HandleWSError(conn, err)
		return
	}
	s.welcome(execParam.UserName, execParam.PodName, execParam.Rule, conn)
	session.AddFilter(s.createCommandFilter())
	defer session.Close()

	session.Start(cancelCtx, traceId)
	if err := session.Executor.Shell(cancelCtx, execParam, session.InterStream); err != nil {
		log.Logger.Warnf("run shell err:%v, traceId=%s", err, traceId)
		remote.HandleWSError(conn, err)
		return
	}

	session.IsRemoteClosed = true
	log.Logger.Infof("exit the shell with remote, traceId=%s", traceId)
}

func (s *Ssh) welcome(userName, podName, rule string, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\r\n"))
	if remote.Banner != "" {
		conn.WriteMessage(websocket.TextMessage, []byte(remote.Banner))
	}
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\r\n"))
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("HI %s! welcome to use Dockin ssh terminal.", userName)))
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("It is a tools used to execute command under docker environment, with safe controls")))
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Current login podName: %s, rule: %s.", podName, rule)))
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("If there are any problems in use this tools, please feel free to contact dockin-helper for help.")))
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))

	var cmdList []string
	if config.OpsConfig.CmdFilterType == remote.BlacklistMode {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Current command filter type:%s", remote.BlacklistMode)))
		conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Current forbinden command list as follow:")))
		conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Current command filter type:%s", remote.WhitelistMode)))
		conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Current support command list as follow:")))
		conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
		cmdList = s.GetCommandList()
	}
	tmpBuf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(tmpBuf, 8, 0, 1, ' ', tabwriter.StripEscape)
	for idx, cmd := range cmdList {
		fmt.Fprint(tw, "\t"+cmd)
		if (idx+1)%10 == 0 {
			fmt.Fprint(tw, "\r\n")
		}
	}
	tw.Flush()

	conn.WriteMessage(websocket.TextMessage, tmpBuf.Bytes())
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n"))
}

func (s *Ssh) createCommandFilter() remote.Filter {
	cmdFilterType := config.OpsConfig.CmdFilterType
	if cmdFilterType == remote.BlacklistMode {
		return remote.NewBlacklistFilter(func() []string {
			return nil
		})
	}
	return remote.NewWhitelistFilter(func() []string {
		return s.GetCommandList()
	})
}

func (s *Ssh) GetCommandList() []string {
	log.Logger.Infof("start to GetCommandList")
	raw, err := s.RedisClient.SMembers(keys.GetRawCmdRedisKey())
	if err != nil {
		log.Logger.Warnf(err.Error())
	}

	common, err := s.RedisClient.SMembers(keys.GetCommonCmdRedisKey())
	if err != nil {
		log.Logger.Warnf(err.Error())
	}
	raw = append(raw, common...)
	log.Logger.Infof("end to GetCommandList")
	return raw
}
