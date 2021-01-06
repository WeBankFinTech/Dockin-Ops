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

package rm

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/utils/trace"
	"github.com/pkg/errors"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	jsoniter "github.com/json-iterator/go"
	v1 "k8s.io/api/core/v1"
)

type Rm struct {
	Cm          *client.Manager
	RedisClient *redis.RedisClient
}

func NewRM(cm *client.Manager, r *redis.RedisClient) *Rm {
	com := &Rm{}
	com.Cm = cm
	com.RedisClient = r
	http.HandleFunc("/v1/dockin/opserver/rm", com.Handle)
	return com
}

func (rm *Rm) Handle(writer http.ResponseWriter, req *http.Request) {
	var (
		opsResult *model.OpsResult
		opsOpts   *model.OpsOption
		err       error
	)
	traceId := trace.TraceID()
	log.Logger.Infof("recv rm request,traceId=%s", traceId)
	if opsOpts, err = api.ValidateReq(req); err != nil {
		opsResult = model.FailedOpsResult(errors.Errorf("%s,traceId=%s", traceId))
		writer.Write(opsResult.ToByte())
		return
	}

	rmOps := &RmOps{}
	listdata, err := rmOps.GetRmPodResource(opsOpts, traceId)
	if err != nil {
		opsResult = model.FailedOpsResult(err)
		writer.Write(opsResult.ToByte())
		return
	}
	rm.batchSetPodResource(listdata)
	opsResult = model.SuccessOpsResult(listdata)
	log.Logger.Infof("handle rm result %s,traceId=%s", opsResult.ToString(), traceId)
	writer.Write(opsResult.ToByte())
	return
}

func (rm *Rm) batchSetPodResource(listResult []*model.ListResult) {
	for _, data := range listResult {
		key := keys.PodYAMLKey(data.PodName)
		podStr, err := rm.RedisClient.Get(key)
		if err != nil || podStr == "" {
			log.Logger.Warnf("failed to get yaml pod from redis, key=%s", key)
			continue
		}
		pod := &v1.Pod{}
		if err := jsoniter.Unmarshal([]byte(podStr.(string)), pod); err != nil {
			log.Logger.Warnf("failed to unmarshal pod from , data=%s, err=%s", podStr.(string), err.Error())
			continue
		}

		setPodResource(pod, data)
	}
}

func setPodResource(pod *v1.Pod, data *model.ListResult) {
	log.Logger.Infof("start to set pod status, podName=%s", data.PodName)
	scs := pod.Status.ContainerStatuses
	for _, con := range scs {
		if strings.Contains(data.PodName, con.Name) {
			data.SubSysTag = ""
			begin := strings.Index(con.Image, ":")
			end := strings.Index(con.Image, "_img")
			if begin == -1 || end == -1 {
				continue
			}
			data.Version = con.Image[begin+1 : end]
		}
	}
	data.State = getPodStatus(pod)
	data.HostIp = pod.Status.HostIP
	log.Logger.Infof("end query podinfo from apiserver response=%#v", data)
}

func getPodStatus(pod *v1.Pod) string {
	restarts := 0
	readyContainers := 0
	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}
	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	return reason
}
