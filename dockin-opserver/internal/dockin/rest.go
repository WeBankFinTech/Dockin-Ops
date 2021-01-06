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

package dockin

import (
	"fmt"
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/rest"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var (
	DefaultTimeout = time.Duration(10 * time.Second)
	HostKey        = "app.rm.api"
	EmptyString    = ""
	rmApiUrl       string
)

func init() {
	rmApiUrl = config.OpsConfig.RMAddress
}

func GetPodListInfoByHostIp(hostIp string) (*model.RmResultDto, error) {
	if EmptyString == hostIp {
		log.Logger.Warnf("host ip is empty")
		return nil, errors.New("invalid host ip")
	}

	url := fmt.Sprintf("%s/%s?hostIp=%s", rmApiUrl, "getPodInfoByHostIp", hostIp)

	content, err := rest.HttpGet(url, DefaultTimeout)
	if err != nil {
		log.Logger.Warnf("http get GetPodListInfoByHostIp, hostIp=%s err %s",
			hostIp, err.Error())
		return nil, err
	}

	rmresult := &model.RmResultDto{}
	if err = jsoniter.Unmarshal(content, rmresult); err != nil {
		log.Logger.Warnf("GetPodListInfoByHostIp, unmarshal err, ip=%s,content=%s",
			hostIp, string(content))
		return nil, err
	}
	if rmresult.Code != 0 {
		err = errors.Errorf("code = %d,get pod list info by hostIP=%s", rmresult.Code, hostIp)
		log.Logger.Warnf(err.Error())
		return nil, err
	}
	return rmresult, nil
}

func GetPodInfoByPodIp(podIp string) (*model.OneRmResultDto, error) {
	if EmptyString == podIp {
		log.Logger.Warnf("pod ip is empty")
		return nil, errors.New("invalid host ip")
	}
	url := fmt.Sprintf("%s/%s?podIp=%s", rmApiUrl, "getPodInfoByPodIp", podIp)
	content, err := rest.HttpGet(url, DefaultTimeout)
	if err != nil {
		log.Logger.Warnf("http get GetPodInfoByPodIp error, url=%s", url)
		return nil, err
	}

	oneresult := &model.OneRmResultDto{}
	err = jsoniter.Unmarshal(content, oneresult)
	if err != nil {
		log.Logger.Warnf("unmarshal error,get pod info by podIp from rm error=%s", err.Error())
		return nil, err
	}
	if oneresult.Code != 0 {
		err = errors.Errorf("code = %d,get pod info by podIp=%s", oneresult.Code, podIp)
		log.Logger.Warnf(err.Error())
		return nil, err
	}
	return oneresult, nil
}

func GetPodInfoByPodName(podName string) (*model.OneRmResultDto, error) {
	if EmptyString == podName {
		log.Logger.Warnf("pod name is empty")
		return nil, errors.New("pod name is empty")
	}
	url := fmt.Sprintf("%s/%s?podName=%s", rmApiUrl, "getPodInfoByPodName", podName)
	content, err := rest.HttpGet(url, DefaultTimeout)
	if err != nil {
		log.Logger.Warnf("http get PodInfoByHostIp err, podName=%s, err=%s", podName, err.Error())
		return nil, err
	}
	rmresult := &model.OneRmResultDto{}
	if err = jsoniter.Unmarshal(content, rmresult); err != nil {
		log.Logger.Warnf("unmarshal PodInfoByHostIp err, podName=%s, err=%s", podName, err.Error())
		return nil, err
	}
	if rmresult.Code != 0 {
		err = errors.Errorf("code = %d,get pod info by podName=%s failed from rm interface", rmresult.Code, podName)
		log.Logger.Warnf(err.Error())
		return nil, err
	}
	return rmresult, nil
}

func GetPodInfoBySubsystem(subsystem, dcn, traceId string) (*model.RmResultDto, error) {
	log.Logger.Infof("start to GetPodInfoBySubsystem,traceId=%s", traceId)
	if EmptyString == subsystem && EmptyString == dcn {
		log.Logger.Warnf("subsystem and dcn at least provide one parameter,traceId=%s", traceId)
		return nil, errors.New("subsystem and dcn at least provide one parameter")
	}
	url := fmt.Sprintf("%s/%s", rmApiUrl, "getPodInfoBySubsystem")
	if EmptyString != subsystem {
		url = fmt.Sprintf("%s?subsystem=%s", url, subsystem)
	}
	if EmptyString != dcn {
		url = fmt.Sprintf("%s&dcn=%s", url, dcn)
	}

	content, err := rest.HttpGet(url, DefaultTimeout)
	if err != nil {
		log.Logger.Warnf("http get PodInfoByHostIp err, subsystem=%s,dcn=%s err %s,traceId=%s",
			subsystem, dcn, err.Error(), traceId)
		return nil, err
	}

	rmresult := &model.RmResultDto{}
	if err = jsoniter.Unmarshal(content, rmresult); err != nil {
		log.Logger.Warnf("unmarshal err, subsystem=%s,dcn=%s err %s,traceId=%s",
			subsystem, dcn, err.Error(), traceId)
		return nil, err
	}
	if rmresult.Code != 0 {
		err = errors.Errorf("code = %d,get pod info by subsystem=%s, dns=%s,traceId=%s",
			rmresult.Code, subsystem, dcn, traceId)
		log.Logger.Warnf(err.Error())
		return nil, err
	}

	log.Logger.Infof("end to GetPodInfoBySubsystem,traceId=%s", traceId)
	return rmresult, nil
}

func GetClusterIdByHostIp(hostIp string) (string, error) {
	if EmptyString == hostIp {
		log.Logger.Warnf("host ip is empty")
		return "", errors.New("invalid host ip")
	}
	tempIp := hostIp
	tempIp = strings.ReplaceAll(tempIp, "-", ".")

	url := fmt.Sprintf("%s/%s?hostIp=%s", rmApiUrl, "getClusterId", tempIp)
	content, err := rest.HttpGet(url, DefaultTimeout)
	if err != nil {
		log.Logger.Warnf("get clusterId by hostIp =%s, error=%s", tempIp, err.Error())
		return "", err
	}

	cluseterInfo := &model.RmClusterData{}
	if err = jsoniter.Unmarshal(content, cluseterInfo); err != nil {
		log.Logger.Warnf("unmarshal err, hostIp=%s， err %s", tempIp, err.Error())
		return "", err
	}
	if cluseterInfo.Code != 0 {
		err = errors.Errorf("code = %d,get cluster info by hostIp=%s",
			cluseterInfo.Code, tempIp)
		log.Logger.Warnf(err.Error())
		return "", err
	}
	return cluseterInfo.Data.ClusterID, nil
}

func BatchGetPodInfoByPodName(podNameList []string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", rmApiUrl, "getPodInfosByPodNameList")
	payload, _ := jsoniter.Marshal(podNameList)
	return rest.HttpPost(url, payload)
}

func GetPodInfoByPodSetId(podSetId string) (*model.RmResultData, error) {
	url := fmt.Sprintf("%s/%s?podSetId=%s", rmApiUrl, "getPodInfoByPodSetId", podSetId)
	content, err := rest.HttpGet(url, DefaultTimeout)
	if err != nil {
		log.Logger.Warnf("get pod info by podSetId =%s, error=%s", podSetId, err.Error())
		return nil, err
	}

	dto := &model.RmResultDto{}
	if err = jsoniter.Unmarshal(content, dto); err != nil {
		log.Logger.Warnf("unmarshal err, hostIp=%s， err %s", podSetId, err.Error())
		return nil, err
	}
	if dto.Code != 0 {
		err = errors.Errorf("code = %d,get pod info by podSetId=%s",
			dto.Code, podSetId)
		log.Logger.Warnf(err.Error())
		return nil, err
	}
	for _, rrd := range dto.Data {
		if rrd.Status == "ALLOCATED" {
			return rrd, nil
		}
	}
	return nil, fmt.Errorf("no running pod exist, for pod set id=%s", podSetId)
}
