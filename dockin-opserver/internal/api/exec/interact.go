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
	"fmt"
	"net/http"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/utils/cmd"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/remote"
	"github.com/webankfintech/dockin-opserver/internal/utils/ip"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"
	"github.com/gorilla/websocket"
)

type Interact struct {
	Cm          *client.Manager
	RedisClient *redis.RedisClient
}

func NewInteract(cm *client.Manager, r *redis.RedisClient) *Interact {
	inter := &Interact{}
	inter.Cm = cm
	inter.RedisClient = r
	http.HandleFunc("/v1/dockin/opserver/interact-exec", inter.Handle)
	return inter
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (i *Interact) Handle(writer http.ResponseWriter, req *http.Request) {
	var (
		err     error
		opsOpts *model.OpsOption
		pod     *v1.Pod
	)
	traceId := trace.TraceID()
	log.Logger.Infof("receive interactive v2 request,traceId=%s", traceId)
	if opsOpts, err = api.ValidateExecRequest(req, traceId); err != nil {
		log.Logger.Warnf("failed to validate the exec param, as=%v, traceId=%s", err, traceId)
		writer.Write([]byte("failed to validate the exec param, as:" + err.Error() + traceId))
		return
	}

	log.Logger.Infof("ops=%s, traceId=%s", opsOpts.String(), traceId)
	conn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Logger.Warnf("failed to update the connection, err:%v, traceId=%s", err, traceId)
		remote.HandleWSError(conn, err)
		return
	}
	log.Logger.Infof("success to Upgrade webSocket protocol traceId=%s", traceId)

	if err := i.validateCmd(opsOpts); err != nil {
		log.Logger.Warnf("validateCmd err:%v, traceId=%s", err, traceId)
		remote.HandleWSError(conn, err)
		return
	}

	reqIp := ip.GetIp(req)
	pod, err = api.GetPodStructFromRedis(opsOpts.Name, i.RedisClient)
	if err != nil {
		log.Logger.Warnf("failed to get pod struct from redis,podName=%s,err=%s traceId=%s", opsOpts.Name, err, traceId)
		remote.HandleWSError(conn, err)
		return
	}

	hostIp, err := api.GetHostIpByPod(opsOpts, pod, i.Cm, reqIp, traceId)
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

	exec := &ExecCommand{
		OpsOpts: opsOpts,
		Conn:    conn,
		HostIp:  hostIp,
	}
	if err := exec.RunWithTty(traceId); err != nil {
		log.Logger.Warnf("run interactive exec err:%v, traceId=%s", err, traceId)
		//remote.HandleWSError(conn, err)
		return
	}

	log.Logger.Infof("finish interactive v2 request, traceId=%s", traceId)
}

func (i *Interact) validateCmd(opt *model.OpsOption) error {
	cmdInStr := strings.Join(opt.Flags, " ")
	cmdList, err := cmd.GetCommandList(cmdInStr)
	if err != nil {
		ne := fmt.Errorf("failed to parse cmd %s, err=%v", cmdInStr, err)
		log.Logger.Infof(ne.Error())
		return ne
	}
	sz := len(cmdList)
	for i := 0; i < sz; i++ {
		in := cmdList[i]
		for _, cmd := range forbiddenExecCmd {
			if strings.EqualFold(in, cmd) {
				if strings.EqualFold("/bin/bash", in) || strings.EqualFold("/bin/sh", in) {
					if i < sz-1 {
						nextC := cmdList[i+1]
						if nextC == "-c" {
							continue
						}
					}
				}
				log.Logger.Warnf("not allowed command, cmd=%s", in)
				return errors.Errorf("command [%s] is not allowed to execute", in)
			}
		}
	}
	return nil
}

func (s *Interact) getHostIp(opsOpts *model.OpsOption, reqIp, traceId string) (string, error) {
	hostIp := opsOpts.HostIP
	pod := v1.Pod{}
	if podstr, err := s.RedisClient.Get(keys.PodYAMLKey(opsOpts.Name)); err == nil {
		if err := jsoniter.Unmarshal([]byte(podstr.(string)), &pod); err == nil {
			hostIp = pod.Status.HostIP
			log.Logger.Infof("use host ip from redis cache, stsName=%s, HostIp=%s,traceId=%s", opsOpts.Name, hostIp, traceId)
			return hostIp, nil
		}
	}

	_, err := s.Cm.GetProxyClient(reqIp, opsOpts.Rule, opsOpts.ClusterId)
	if err != nil {
		log.Logger.Warnf("no proxy config found for ip=%s, rule=%s,traceId=%s", reqIp, opsOpts.Rule, traceId)
		return "", fmt.Errorf("no proxy config found for ip=%s, rule=%s", reqIp, opsOpts.Rule)
	}
	return hostIp, nil
}
