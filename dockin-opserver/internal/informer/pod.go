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

package informer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/cmap"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/log"

	jsoniter "github.com/json-iterator/go"
	v1 "k8s.io/api/core/v1"
)

type PodInformer struct {
	RedisClient *redis.RedisClient
	HttpMap     cmap.ConcurrentMap
}

const uuidPodInfoExpirationTime = 7 * 24 * 60 * time.Minute

type PreStopReq struct {
	PodName string `json:"podName"`
	Command string `json:"command"`
}

func (p *PodInformer) AddFunc(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Logger.Infof("AddFunc obj is not pod type")
		return
	}
	log.Logger.Debugf("add pod podName:%s, UID:%s", pod.Name, string(pod.UID))
	podInfo, _ := jsoniter.Marshal(pod)
	key := keys.PodUUIDKey(string(pod.UID))

	p.getPodPreStopInfoAndSend(obj)

	err := p.RedisClient.Set(key, string(podInfo), uuidPodInfoExpirationTime)
	if err != nil {
		log.Logger.Warnf("add pod uid Set key:%s err=%s", key, err.Error())
	}

	key = keys.PodYAMLKey(pod.Name)
	log.Logger.Debugf("add pod Set key:%s,value:%s", key, string(podInfo))
	err = p.RedisClient.Set(key, string(podInfo), 0)
	if err != nil {
		log.Logger.Warnf("add pod Set key:%s err=%s", key, err.Error())
	}

	log.Logger.Debugf("Set podName key:%s success", key)
}

func (p *PodInformer) UpdateFunc(oldObj, newObj interface{}) {
	pod, ok := newObj.(*v1.Pod)
	if !ok {
		log.Logger.Infof("UpdateFunc obj is not pod type")
		return
	}
	oldPod := oldObj.(*v1.Pod)
	log.Logger.Debugf("update pod,oldPodName:%s,Uid:%s、newPodName:%s,Uid:%s",
		oldPod.Name, string(oldPod.UID), pod.Name, string(pod.UID))

	podInfo, _ := jsoniter.Marshal(pod)
	key := keys.PodUUIDKey(string(pod.UID))
	err := p.RedisClient.Set(key, string(podInfo), uuidPodInfoExpirationTime)
	if err != nil {
		log.Logger.Warnf("Set key:%s err=%s", key, err.Error())
	}

	key = keys.PodYAMLKey(string(pod.Name))
	log.Logger.Debugf("Set key:%s,value:%s", key, string(podInfo))
	err = p.RedisClient.Set(key, string(podInfo), 0)
	if err != nil {
		log.Logger.Warnf("Set key:%s err=%s", key, err.Error())
	}

	log.Logger.Debugf("Set new podName key:%s success", key)

}

func (p *PodInformer) getPodPreStopInfoAndSend(obj interface{}) {
	var (
		pod    *v1.Pod
		hostIp string
	)
	uid := uuid.New().String()
	pod = obj.(*v1.Pod)
	log.Logger.Debugf("start getPodPreStopInfoAndSend podName=%s,uid=%s", pod.Name, uid)

	podInfo, _ := jsoniter.Marshal(pod)

	hostIp = pod.Status.HostIP
	if hostIp == "" {
		//pending 状态（需UpdateFunc 获取running状态pod hostip）
		log.Logger.Debugf("getPodPreStopInfoAndSend pod hostIp is null,pod yaml info=%s", string(podInfo))
		return
	}

	lp := pod.Spec.Containers[0].Lifecycle
	if lp == nil {
		log.Logger.Debugf("getPodPreStopInfoAndSend pod Lifecycle is nil,pod yaml info=%s", string(podInfo))
		return
	}
	pp := lp.PreStop
	if pp == nil {
		log.Logger.Debugf("getPodPreStopInfoAndSend pod PreStop is nil,pod yaml info=%s", string(podInfo))
		return
	}

	ep := pp.Exec
	if ep == nil {
		log.Logger.Debugf("getPodPreStopInfoAndSend pod Exec is nil,pod yaml info=%s", string(podInfo))
		return
	}

	if len(ep.Command) < 3 {
		log.Logger.Debugf("getPodPreStopInfoAndSend pod command len is valid,command=%s", ep.Command)
		return
	}

	hc := p.getHttpClient(hostIp)
	sendPreStopRequest(hostIp, pod.Name, ep.Command[2], uid, hc)

	log.Logger.Debugf("end to getPodPreStopInfoAndSend podName=%s,uid=%s", pod.Name, uid)
}

func (p *PodInformer) DeleteFunc(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Logger.Infof("DeleteFunc obj is not pod type")
		return
	}
	log.Logger.Debugf("delete pod,podName: %s, UID: %s", pod.Name, string(pod.UID))

	key := keys.PodYAMLKey(string(pod.Name))
	err := p.RedisClient.Del(key)
	if err != nil {
		log.Logger.Warnf("Del key:%s err=%s", key, err.Error())
	}

	log.Logger.Debugf("Del podName key:%s success", key)
}

func (p *PodInformer) getHttpClient(nodeIp string) *http.Client {
	client, ok := p.HttpMap.Get(nodeIp)
	if !ok {
		log.Logger.Debugf("no http client found for nodeIp=%s, create new", nodeIp)
		nc := &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5000,
			},
			Timeout: time.Second * 5,
		}
		p.HttpMap.Set(nodeIp, nc)
		return nc
	}
	return client.(*http.Client)
}

func sendPreStopRequest(nodeIp, podName, command, uid string, nodeHttp *http.Client) {
	urlStr := fmt.Sprintf("http://%s:%d/dockin/opagent/prestop", nodeIp, config.OpsConfig.OpAgentPort)

	req := &PreStopReq{
		PodName: podName,
		Command: command,
	}
	payload, _ := jsoniter.MarshalToString(req)

	log.Logger.Debugf("sent preStop request url:%s,payload=%s,uid=%s", urlStr, payload, uid)
	httpReq, _ := http.NewRequest("POST", urlStr, strings.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("access-token", model.OpagentAccessToken())
	resp, err := nodeHttp.Do(httpReq)

	if err != nil {
		log.Logger.Warnf("send http preStop req to opAgent failed, uid=%s, err=%s", uid, err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrapf(err, "read opAgent body failed err=%s, url=%s, uid=%s",
			err.Error(), urlStr, uid)
		log.Logger.Warnf(err.Error())
		return
	}

	log.Logger.Debugf("get preStop result from opagent result=%s,url=%s, uid=%s", string(body), urlStr, uid)
	return
}
