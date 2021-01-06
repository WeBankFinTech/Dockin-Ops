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
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/webankfintech/dockin-opsctl/internal/common"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)


func ValidatePod(rule, podName, userName string) (podInfo *PodInfo, err error) {
	baseUrl := common.GetCommonUrlByCmd("ctrl/getPodByName")
	reqUrl := fmt.Sprintf("%s?rule=%s&podName=%s&userName=%s", baseUrl, rule, podName, userName)
	body, err := utils.HttpGetWithTimeout(reqUrl, time.Second*3)
	if err != nil {
		log.Debugf("timeout to get pod failed podName=%s, rule=%s", err.Error(), podName, rule)
		return nil, errors.Errorf("timeout to get pod info failed,podName=%s,rule=%s", podName, rule)
	}

	result := make(map[string]interface{})
	if err := jsoniter.Unmarshal(body, &result); err != nil {
		return nil, errors.New("validate pod or account failed")
	}

	code := result["Code"].(float64)
	if code != 0 {
		msg := result["Message"].(string)
		return nil, errors.Errorf("get pod from server failed, message=%s", msg)
	}

	//将 map 转换为指定的结构体
	if err := mapstructure.Decode(result["Data"].(map[string]interface{}), &podInfo); err != nil {
		log.Debugf("mapstructure Decode Data map=%v to PodInfo failed", result["Data"])
		return nil, errors.Errorf("parse login response data faild")
	}

	return podInfo, nil

}

