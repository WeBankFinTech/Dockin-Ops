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
	"net/http"
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/dockin"
	"github.com/webankfintech/dockin-opserver/internal/utils/aes"

	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/base"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var (
	errNewAes       = fmt.Errorf("internal exception decrypt, check key")
	errDecrypt      = fmt.Errorf("invalid request, access token is illegal as decrypt failed")
	errUnmarshal    = fmt.Errorf("invalid request, access token is illegal as unmarshal failed")
	errTokenExpired = fmt.Errorf("access token is expired, try to relogin")
)

func ValidateAccessToken(token, traceId string) error {
	ac, err := ParseAccessToken(token, traceId)
	if err != nil {
		log.Logger.Warnf(err.Error())
		return err
	}

	if int64(time.Now().Sub(ac.CreateTime).Seconds()) > ac.Expire {
		log.Logger.Warnf("access token is expired, create at %v, token=%s,traceId=%s", ac.CreateTime, token, traceId)
		return errTokenExpired
	}
	return nil
}

func ParseAccessToken(token, traceId string) (*model.UserIdentity, error) {
	log.Logger.Infof("start to ParseAccessToken,token=%s,traceId=%s", token, traceId)
	if token == "" {
		log.Logger.Infof("token is empty,traceId=%s", traceId)
		return nil, fmt.Errorf("token is empty, please check login status")
	}
	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		log.Logger.Warnf("new aes %s, err %v, token=%s,traceId=%s", errNewAes.Error(), err, token, traceId)
		return nil, errNewAes
	}

	accD, err := aes.AesDecrypt(token)
	if err != nil {
		log.Logger.Warnf("invalid request, access token is illegal as decrypt failed, token=%s, err=%v,traceId=%s", token, err, traceId)
		return nil, errDecrypt
	}

	log.Logger.Infof("get access token result str %s,traceId=%s", accD, traceId)
	ac := &model.UserIdentity{}
	if err = jsoniter.UnmarshalFromString(accD, ac); err != nil {
		log.Logger.Warnf("invalid request, access token is illegal as unmarshal failed, err=%v, token=%s,traceId=%s", err, token, traceId)
		return nil, errUnmarshal
	}

	log.Logger.Infof("end to ParseAccessToken,token=%s,traceId=%s", token, traceId)
	return ac, nil
}

func ValidateReq(req *http.Request) (opts *model.OpsOption, err error) {
	err = req.ParseForm()
	if err != nil {
		return
	}
	params := req.Form.Get("params")
	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		err = errors.Errorf("encrypt failed err=%s", err.Error())
		log.Logger.Warnf(err.Error())
		return
	}

	decoded, err := aes.AesDecrypt(params)
	if err != nil {
		err = errors.Errorf("decrypt err=%s", err.Error())
		log.Logger.Warnf(err.Error())
		return
	}

	log.Logger.Infof("handle request param = %s", decoded)
	opts = &model.OpsOption{}
	err = jsoniter.Unmarshal([]byte(decoded), opts)
	if err != nil {
		log.Logger.Warnf("unmarshal read data error, check the input data, err=%s", err.Error())
		err = errors.Errorf("unmarshal read data error, check the input data, %s", err.Error())
		return
	}
	rule, exist := opts.Params["rule"]
	if !exist {
		rule = "default"
	}
	opts.Rule = rule.(string)
	namespace, exist := opts.Params["namespace"]
	if exist {
		opts.Namespace = namespace.(string)
	}

	operator, exist := opts.Params["operator"]
	if !exist {
		operator = opts.Rule
	}
	opts.Operator = operator.(string)

	opts.Image = config.OpsConfig.Debug.Image
	image, exist := opts.Params["image"]
	if exist {
		opts.Image = image.(string)
	}
	return
}

func SetOption(o *model.OpsOption) error {
	if o.Name == "" {
		return nil
	}

	switch o.Resource {
	case "pods", "pod", "po":
		return SetPodOption(o)
	case "node", "nodes", "no":
		return SetNodeOption(o)
	}
	return nil
}

func SetPodOption(o *model.OpsOption) error {
	var (
		podName       string
		podIp         string
		clusterId     string
		containerName string
		hostIP        string
	)
	input := o.Name
	if base.IsIp(input) {
		log.Logger.Infof("IsIp get podInfo input=%s", input)
		rmData, err := dockin.GetPodInfoByPodIp(input)
		if err != nil {
			log.Logger.Warnf(err.Error())
			return err
		}
		podName = rmData.Data.PodName
		podIp = o.Name
		clusterId = rmData.Data.ClusterID
		containerName = rmData.Data.PodName
		hostIP = rmData.Data.HostIP
	} else if base.IsPodSet(input) {
		log.Logger.Infof("IsPodSet get podInfo input=%s", input)
		rmData, err := dockin.GetPodInfoByPodSetId(input)
		if err != nil {
			log.Logger.Warnf(err.Error())
			return err
		}
		podName = rmData.PodName
		podIp = o.Name
		clusterId = rmData.ClusterID
		containerName = rmData.PodName
		hostIP = rmData.HostIP
	} else {
		log.Logger.Infof("else get podInfo input=%s", input)
		rmPodName := input
		if IsStsName(input) {
			rmPodName = input[:len(input)-2]
		}
		rmData, err := dockin.GetPodInfoByPodName(rmPodName)
		if err != nil {
			log.Logger.Warnf(err.Error())
			return err
		}
		clusterId = rmData.Data.ClusterID
		containerName = rmData.Data.PodName
		hostIP = rmData.Data.HostIP
		podName = o.Name
		podIp = rmData.Data.PodIP
	}

	o.ClusterId = clusterId
	o.Container = containerName
	o.Name = podName
	o.HostIP = hostIP
	o.PodIp = podIp

	log.Logger.Infof("validate pod option result, podName=%s, containerName=%s, clusterId=%s, hostIP = %s",
		podName, containerName, clusterId, hostIP)
	return nil
}

func SetNodeOption(o *model.OpsOption) error {
	hostIp := o.Name
	hostIp = strings.ReplaceAll(hostIp, "_", ".")

	cid, err := dockin.GetClusterIdByHostIp(o.Name)
	if err != nil {
		log.Logger.Warnf(err.Error())
		return err
	}
	o.ClusterId = cid
	return nil
}

func SetNamespace(o *model.OpsOption, inputNS string) {
	if o.Namespace == "" {
		log.Logger.Infof("use default namespace = %s, from yaml context", inputNS)
		o.Namespace = inputNS
	}
}

func GetPodName(name string) string {
	if IsStsName(name) {
		return name[:len(name)-2]
	}
	return name
}

func IsStsName(podName string) bool {
	return false //env.IsTestEnv() && strings.HasSuffix(podName, "-0")
}

func ValidateExecRequest(req *http.Request, traceId string) (opts *model.OpsOption, err error) {
	log.Logger.Infof("start to exec Validate Req traceId=%s", traceId)
	opts, err = ValidateReq(req)
	if err != nil {
		return nil, err
	}

	if err = SetPodOption(opts); err != nil {
		return nil, err
	}

	return
}
