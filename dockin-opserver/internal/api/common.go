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

package api

import (
	"fmt"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"

	jsoniter "github.com/json-iterator/go"
	v1 "k8s.io/api/core/v1"
)

func GetPodStructFromRedis(podName string, RedisClient *redis.RedisClient) (*v1.Pod, error) {
	var (
		pod *v1.Pod
		err error
	)
	pod = &v1.Pod{}

	if podstr, err := RedisClient.Get(keys.PodYAMLKey(podName)); err == nil {
		if err := jsoniter.Unmarshal([]byte(podstr.(string)), pod); err == nil {
			log.Logger.Infof("success to get pod struct from redis by podName=%s", podName)
			return pod, nil
		}
		return nil, err
	}
	return nil, err
}

func GetHostIpByPod(opsOpts *model.OpsOption, pod *v1.Pod, cm *client.Manager, reqIp, traceId string) (string, error) {
	hostIp := opsOpts.HostIP
	hostIp = pod.Status.HostIP

	_, err := cm.GetProxyClient(reqIp, opsOpts.Rule, opsOpts.ClusterId)
	if err != nil {
		log.Logger.Warnf("no proxy config found for ip=%s, rule=%s,traceId=%s", reqIp, opsOpts.Rule, traceId)
		return "", fmt.Errorf("no proxy config found for ip=%s, rule=%s", reqIp, opsOpts.Rule)
	}
	return hostIp, nil
}

func GetContainerIdByPod(podName string, pod *v1.Pod) (string, error) {
	var (
		err    error
		substr string
	)

	substr = "docker://"
	containerList := pod.Status.ContainerStatuses
	for _, v := range containerList {
		if !strings.Contains(podName, v.Name) {
			continue
		}
		log.Logger.Infof("get containerId=%s", v.ContainerID)
		realCid := v.ContainerID[strings.Index(v.ContainerID, substr)+len(substr):]
		log.Logger.Infof("get real containerId=%s", realCid)
		return realCid, nil
	}
	return "", err

}
