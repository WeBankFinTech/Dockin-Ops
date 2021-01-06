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

package echo

import (
	"fmt"

	"github.com/webankfintech/dockin-opserver/internal/common/env"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"

	"net/http"
	"sync"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/base"
	"github.com/webankfintech/dockin-opserver/internal/utils/cmap"
	"github.com/webankfintech/dockin-opserver/internal/utils/ip"

	"github.com/pkg/errors"
)

type Echo struct {
	Cm          *client.Manager
	RedisClient *redis.RedisClient
}

func NewEcho(cm *client.Manager, r *redis.RedisClient) *Echo {
	echo := &Echo{}
	echo.Cm = cm
	echo.RedisClient = r
	http.HandleFunc("/v1/dockin/opserver/echo", echo.Handle)
	return echo
}

func (e *Echo) Handle(writer http.ResponseWriter, req *http.Request) {
	var (
		opsResult *model.OpsResult
		opsOpts   *model.OpsOption
		err       error
	)
	traceId := trace.TraceID()
	if opsOpts, err = api.ValidateReq(req); err != nil {
		log.Logger.Infof("validate request failed, err=%v, traceId=%s", err, traceId)
		opsResult = model.FailedOpsResult(fmt.Errorf("%s,traceId=%s", err.Error(), traceId))
		writer.Write(opsResult.ToByte())
		return
	}

	log.Logger.Infof("recv echo request, option=%#v, traceId=%s", opsOpts, traceId)
	reqIp := ip.GetIp(req)
	if opsOpts.Name != "" {
		opsResult = e.getResource(reqIp, opsOpts, traceId)
	} else {
		opsResult = e.batchGetResource(reqIp, opsOpts, traceId)
	}

	log.Logger.Infof("success handle echo request=%#v, traceId=%s", opsOpts, traceId)
	log.Logger.Debugf("handle echo result %s", opsResult.ToString())
	writer.Write(opsResult.ToByte())
}

func (e *Echo) getK8sPodByUuid(uuid, traceId string) (string, error) {
	key := keys.PodUUIDKey(uuid)
	val, err := e.RedisClient.Get(key)
	if err != nil {
		log.Logger.Warnf("get pod by uuid from redis err=%v, traceId=%s", err, traceId)
		return "", err
	} else if val == nil {
		log.Logger.Infof("get pod by uuid from redis ,value is empty, key=%s, traceId=%s", key, traceId)
		return "", errors.Errorf("key:%s is not exist", key)
	} else {
		log.Logger.Infof("get pod info by uuid from redis success, key=%s, value=%v, traceId=%s", key, val, traceId)
		return val.(string), nil
	}
}

func (e *Echo) getResource(reqIp string, opsOpts *model.OpsOption, traceId string) *model.OpsResult {
	log.Logger.Infof("start to getResource,traceId=%s", traceId)
	switch opsOpts.Resource {
	case "pods", "pod", "po":
		if base.IsUUid(opsOpts.Name) {
			log.Logger.Infof("get pod resource by uuid, traceId=%s", opsOpts.Name, traceId)
			result, err := e.getK8sPodByUuid(opsOpts.Name, traceId)
			if err != nil {
				return model.FailedOpsResult(err)
			}

			return model.SuccessOpsResult(result)
		}
	default:
	}

	if err := api.SetOption(opsOpts); err != nil {
		return model.FailedOpsResult(errors.Errorf("get podInfo from rm failed podName=%s, err=%s,traceId=%s",
			opsOpts.Name, err.Error(), traceId))
	}
	clusterId := opsOpts.ClusterId
	log.Logger.Infof("---- cluster id in opts:", clusterId)
	pc, err := e.Cm.GetProxyClient(reqIp, opsOpts.Rule, clusterId)
	if err != nil {
		log.Logger.Warnf("no proxy config found for ns=%s,traceId=%s", opsOpts.Namespace, traceId)
		return model.FailedOpsResult(err)
	}
	api.SetNamespace(opsOpts, pc.K8sConfig.Contexts[0].Context.Namespace)
	getops := &GetOps{}
	getops.ProxyClient = pc
	getops.RedisClient = e.RedisClient
	res, err := getops.GetResource(opsOpts, traceId)
	if err != nil {
		log.Logger.Warnf("send req to apierver return error=%s,traceId=%s", err.Error(), traceId)
		return model.FailedOpsResult(err)
	}

	md := make(map[string]interface{}, 1)
	md[clusterId] = res.Data

	log.Logger.Infof("end to getResource,traceId=%s", traceId)
	return model.SuccessOpsResult(md)
}

func (e *Echo) batchGetResource(ip string, opsOpts *model.OpsOption, traceId string) *model.OpsResult {
	log.Logger.Infof("start to batchGetResource,traceId=%s", traceId)
	var (
		wg  = new(sync.WaitGroup)
		err error
	)
	list, err := e.Cm.GetProxyByRule(opsOpts.Rule)
	if err != nil || len(list) == 0 {
		return model.FailedOpsResult(errors.Errorf("get podInfo from rm failed ip=%s, rule=%s err=%v, traceId=%s",
			ip, opsOpts.Rule, err, traceId))
	}
	mapdata := cmap.New()
	wg.Add(len(list))
	for _, pc := range list {
		go func(proxyClient *client.ProxyClient) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Logger.Warnf("recover from get resources,traceId=%s", traceId)
					return
				}
			}()
			getops := &GetOps{}
			getops.ProxyClient = proxyClient
			getops.RedisClient = e.RedisClient
			opsOpts.ClusterId = pc.K8sConfig.Dockin.ClusterID
			opsOpts.Namespace = pc.K8sConfig.Contexts[0].Context.Namespace
			res, err := getops.GetResource(opsOpts, traceId)
			if err != nil {
				return
			}
			mapdata.Set(proxyClient.K8sConfig.Dockin.ClusterID, res.Data)
		}(pc)
	}

	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		err = nil
	case <-time.After(env.DefaultTimeout()):
		err = errors.Errorf("canceled as timeout task,traceId=%s", traceId)
		log.Logger.Warnf("batch timeout, err = %s", err.Error())
	}
	if err != nil {
		return model.FailedOpsResult(errors.Errorf("batch get resource timeout, ip=%s, rule=%s,traceId=%s",
			ip, opsOpts.Rule, traceId))
	}
	log.Logger.Infof("end to batchGetResource,traceId=%s", traceId)
	return model.SuccessOpsResult(mapdata)
}
