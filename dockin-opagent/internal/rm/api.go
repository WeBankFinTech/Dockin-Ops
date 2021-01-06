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
	"time"

	"github.com/webankfintech/dockin-opagent/internal/config"
	"github.com/webankfintech/dockin-opagent/internal/log"
	"github.com/webankfintech/dockin-opagent/internal/model"
	"github.com/webankfintech/dockin-opagent/internal/utils/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var (
	EmptyString    = ""
	HostKey        = "app.rm.api"
	rmApiUrl       = ""
	DefaultTimeout = time.Duration(5 * time.Second)
)

func init() {
	rmApiUrl = config.AgentConf.App.Rm.API
}

func GetPodInfoByPodIp(podIp string) (*model.OneRmResultDto, error) {
	if EmptyString == podIp {
		log.Logger.Warnf("pod ip is empty")
		return nil, errors.New("invalid host ip")
	}
	url := fmt.Sprintf("%s/%s?podIp=%s", rmApiUrl, "getPodInfoByPodIp", podIp)
	content, err := http.HttpGet(url, DefaultTimeout)
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
