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
	"time"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/dockin"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/base"
	"github.com/pkg/errors"
)

var (
	DefaultTimeout = time.Duration(10 * time.Second)
	HostKey        = "app.rm.host"
	InfoPath       = "dockin-rm"
	EmptyString    = ""
)

type RmOps struct {
	ProxyClient *client.ProxyClient
}

func (g *RmOps) GetRmPodResource(echo *model.OpsOption, traceId string) (listResult []*model.ListResult, err error) {
	log.Logger.Infof("start to GetRmPodResource,traceId=%s", traceId)
	var (
		rrd        []*model.RmResultData
		resultList []*model.ListResult
	)
	if base.IsIp(echo.Name) {
		rrd, err = g.getPodsInfoByIp(echo.Name, traceId)
		if err != nil {
			log.Logger.Warnf("get pod info by podIp err, %#v, err %s,traceId=%s", echo, err.Error(), traceId)
			return nil, err
		}
	} else if base.IsSubsystem(echo.Name) || base.IsSubSystemId(echo.Name) {
		dcn, exist := echo.Params["dcn"]
		if !exist {
			dcn = ""
		}
		rmresult, err := dockin.GetPodInfoBySubsystem(echo.Name, dcn.(string), traceId)
		if err != nil {
			log.Logger.Warnf("get pod info by subsystem err, %#v, err %s,traceId=%s", echo, err.Error(), traceId)
			return nil, err
		}
		rrd = rmresult.Data
	} else if base.IsPodName(echo.Name) {
		podName := api.GetPodName(echo.Name)
		rmresult, err := dockin.GetPodInfoByPodName(podName)
		if err != nil {
			log.Logger.Warnf("get pod by podname from rm failed, %#v, err %s,traceId=%s", echo, err.Error(), traceId)
			return nil, err
		}

		if rmresult.Data != nil {
			rrd = append(rrd, rmresult.Data)
		}
	} else if base.IsPodSet(echo.Name) {
		rmresult, err := dockin.GetPodInfoByPodSetId(echo.Name)
		if err != nil {
			log.Logger.Warnf("get pod by podname from rm failed, %#v, err %s,traceId=%s", echo, err.Error(), traceId)
			return nil, err
		}
		rrd = append(rrd, rmresult)
	} else {
		err = errors.Errorf("un support type %s, must be one of ip/subSysName/podName,traceId=%s", echo.Name, traceId)
		log.Logger.Warnf(err.Error())
		return
	}
	for _, rr := range rrd {
		getresult := &model.ListResult{
			SubSysName: rr.SubSystem,
			SubSysId:   rr.SubSystemId,
			Dcn:        rr.Dcn,
			PodName:    rr.PodName,
			Namespace:  rr.Namespace,
			ClusterId:  rr.ClusterID,
			HostIp:     rr.HostIP,
			PodIp:      rr.PodIP,
			LimitMem:   rr.Mem,
			LimitCpu:   rr.CPU,
		}
		resultList = append(resultList, getresult)
	}
	log.Logger.Infof("end to GetRmPodResource,traceId=%s", traceId)
	return resultList, nil
}

func (g *RmOps) getPodsInfoByIp(ip, traceId string) ([]*model.RmResultData, error) {
	log.Logger.Infof("start to getPodsInfoByIp,traceId=%s", traceId)
	var resultData []*model.RmResultData
	oneresult, err := dockin.GetPodInfoByPodIp(ip)
	if err != nil {
		log.Logger.Warnf("try to get pod info by ip err, %#v, err %s,traceId=%s", ip, err.Error(), traceId)

		mulResult, err := dockin.GetPodListInfoByHostIp(ip)
		if err != nil {
			log.Logger.Warnf("try to get node info by ip err, %s, err=%s,traceId=%s", ip, err.Error(), traceId)
			return nil, err
		}
		resultData = mulResult.Data
	} else {
		resultData = append(resultData, oneresult.Data)
	}

	log.Logger.Infof("end to getPodsInfoByIp,traceId=%s", traceId)
	return resultData, nil
}
