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

	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/pkg/errors"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

var (
	group      = metav1beta1.GroupName
	version    = metav1beta1.SchemeGroupVersion.Version
	tabHeader  = fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)
	yamlHeader = fmt.Sprintf("application/json;as=Yaml;v=%s;g=%s, application/json", version, group)
)

type GetOps struct {
	ProxyClient *client.ProxyClient
	RedisClient *redis.RedisClient
}

func (g *GetOps) GetResource(echo *model.OpsOption,traceId string) (*model.OpsResult, error) {
	log.Logger.Infof("get k8s resource %#v,traceId=%s", echo,traceId)
	var opsResult *model.OpsResult
	result, err := g.GetK8sResource(echo,traceId)
	if err != nil {
		log.Logger.Warnf("get k8s resource err, %#v, err %s,traceId=%s", echo, err.Error(),traceId)
		return nil, err
	}

	opsResult = model.SuccessOpsResult(result)
	return opsResult, nil
}

func (g *GetOps) GetK8sResource(echo *model.OpsOption,traceId string) (result string, err error) {
	switch echo.Resource {
	case "node", "nodes", "no":
		nodeGetter := NewNodeGetter(echo, g.ProxyClient, g.RedisClient, echo.PrintType)
		return nodeGetter.GetNode(traceId)
	case "pods", "pod", "po":
		podGetter := NewPodGetter(echo, g.ProxyClient, g.RedisClient, echo.PrintType)
		return podGetter.GetPod(traceId)
	default:
		err = errors.Errorf("un support resource %s", echo.Resource)
	}
	return
}