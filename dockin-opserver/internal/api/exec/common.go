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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/remote"
	"github.com/webankfintech/dockin-opserver/internal/utils/ip"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
)

type Common struct {
	Cm          *client.Manager
	RedisClient *redis.RedisClient
}

func NewCommon(cm *client.Manager, r *redis.RedisClient) *Common {
	com := &Common{}
	com.Cm = cm
	com.RedisClient = r
	http.HandleFunc("/v1/dockin/opserver/common-exec", com.Handle)
	http.HandleFunc("/v1/dockin/opserver/command-exec", com.CommandHandle)
	return com
}

func (c *Common) Handle(writer http.ResponseWriter, req *http.Request) {
	var (
		opsResult *model.OpsResult
		opsOpts   *model.OpsOption
		err       error
		pod       *v1.Pod
	)
	traceId := trace.TraceID()
	log.Logger.Infof("recv common v2 request,traceId=%s", traceId)
	if opsOpts, err = api.ValidateReq(req); err != nil {
		opsResult = model.FailedOpsResult(errors.Errorf("validate common req err=%s,traceId=%s", err.Error(), traceId))
		writer.Write(opsResult.ToByte())
		return
	}
	log.Logger.Infof("data=%s, traceId=%s", opsOpts.String(), traceId)
	reqIp := ip.GetIp(req)
	log.CommandLogger.Info("exec",
		zap.String("operator", opsOpts.Operator),
		zap.String("ip", reqIp),
		zap.String("command", strings.Join(opsOpts.Flags, " ")),
		zap.String("timestamp", fmt.Sprintf("%d", time.Now().Unix())),
		zap.String("podName", opsOpts.Name),
		zap.String("podIp", opsOpts.PodIp))

	if err := api.SetPodOption(opsOpts); err != nil {
		opsResult = model.FailedOpsResult(errors.Errorf("get podInfo from rm failed podName=%s, err=%s,traceId=%s",
			opsOpts.Name, err.Error(), traceId))
		writer.Write(opsResult.ToByte())
		return
	}

	pod, err = api.GetPodStructFromRedis(opsOpts.Name, c.RedisClient)
	if err != nil {
		log.Logger.Warnf("failed to get pod struct from redis,podName=%s,err=%s traceId=%s", opsOpts.Name, err, traceId)
		writer.Write([]byte(err.Error()))
		return
	}

	hostIp, err := api.GetHostIpByPod(opsOpts, pod, c.Cm, reqIp, traceId)
	if err != nil {
		log.Logger.Warnf("failed to get the host ip for ops=%s, traceId=%s", opsOpts.String(), traceId)
		writer.Write([]byte(err.Error()))
		return
	}

	cid, err := api.GetContainerIdByPod(opsOpts.Name, pod)
	if err != nil {
		nerr := errors.Errorf("get containerId from pod struct by pod=%s failed,err=%s", opsOpts.Name, err.Error())
		writer.Write([]byte(nerr.Error()))
		return
	}

	opsOpts.Container = cid
	exec := &ExecCommand{
		OpsOpts: opsOpts,
		Conn:    nil,
		HostIp:  hostIp,
	}
	ioStreams := &remote.IOStreams{
		In:     bytes.NewBuffer([]byte{}),
		Out:    writer,
		ErrOut: bytes.NewBuffer([]byte{}),
	}

	ioStreams, _, _, _ = remote.NewIOStreams()

	if err := exec.RunNoTty(traceId, ioStreams); err != nil {
		opsResult = model.FailedOpsResult(errors.Errorf(ioStreams.ErrOut.(*bytes.Buffer).String()))
		writer.Write(opsResult.ToByte())
		return
	}
	opsResult = model.SuccessOpsResult(ioStreams.Out.(*bytes.Buffer).String())
	writer.Write(opsResult.ToByte())
	log.Logger.Infof("end to common v2")
}

func (c *Common) CommandHandle(writer http.ResponseWriter, req *http.Request) {
	c.Handle(writer, req)
}

func (s *Common) getHostIp(opsOpts *model.OpsOption, reqIp, traceId string) (string, error) {
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
