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
	"time"

	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
)

type NodeGetter struct {
	ProxyClient *client.ProxyClient
	RedisClient *redis.RedisClient
	Name        string
	PrintType   string
	ClusterId   string
	Rule        string
}

func NewNodeGetter(echo *model.OpsOption, pc *client.ProxyClient, rc *redis.RedisClient, pt string) *NodeGetter {
	pg := &NodeGetter{}
	pg.ProxyClient = pc
	pg.RedisClient = rc
	pg.PrintType = pt
	pg.ClusterId = echo.ClusterId
	pg.Rule = echo.Rule
	pg.Name = echo.Name
	return pg
}

func (p *NodeGetter) GetNode(traceId string) (string, error) {
	if p.Name == "" {
		return "", errors.Errorf("get node must specify a node name")
	}
	if p.PrintType == "wide" {
		return p.GetNodeWide(traceId)
	}
	return p.GetNodeYAML(traceId)
}

func (p *NodeGetter) GetNodeWide(traceId string) (string, error) {
	log.Logger.Infof("start to GetNodeWide,traceId=%s", traceId)
	var (
		content    string
		expiration = time.Duration(config.OpsConfig.RedisConfig.Expiration) * time.Millisecond
		err        error
	)
	key := keys.NodeWideKey(p.Name)
	if givenData, err := p.RedisClient.Get(key); err == nil && givenData != "" {
		content = givenData.(string)
		log.Logger.Infof("get node wide from redis success, key=%s, content=%s,traceId=%s", key, content, traceId)
		return content, nil
	}
	resp, err := p.GetNodeFromApiServer()
	if err != nil {
		log.Logger.Warnf("get node wide from apiserver failed, as %s,traceId=%s", err.Error(), traceId)
		return "", err
	}
	td := &metav1beta1.Table{}
	if err = jsoniter.Unmarshal([]byte(resp), td); err != nil {
		log.Logger.Warnf("failed to unmarshal node response data from apiserver failed, as %s, data=%s,traceId=%s",
			err.Error(), resp, traceId)
		return "", err
	}

	pws := model.V1Table2NodeWide(td, p.ClusterId)
	if content, err = jsoniter.MarshalToString(pws); err != nil {
		log.Logger.Warnf("failed to unmarshal node apiserver data to row data, as err=%s, data=%s,traceId=%s",
			err.Error(), content, traceId)
		return "", err
	}
	if p.RedisClient.Set(key, pws, expiration); err != nil {
		log.Logger.Warnf("failed to set row data to redis, as err=%s, data=%s,traceId=%s",
			err.Error(), pws, traceId)
	}

	log.Logger.Infof("end to GetNodeWide,traceId=%s", traceId)
	return content, nil
}

func (p *NodeGetter) GetNodeFromApiServer() (string, error) {
	req := p.ProxyClient.ApiClient.CoreV1().RESTClient().Get().
		Resource("nodes").
		VersionedParams(&metav1.GetOptions{ResourceVersion: "0"}, scheme.ParameterCodec)
	if p.PrintType == "wide" || p.PrintType == "" {
		req.SetHeader("Accept", tabHeader)
	} else if p.PrintType == "yaml" {
		req.SetHeader("Accept", yamlHeader)
	}
	if p.Name != "" {
		req.Name(p.Name)
	}

	result := req.Do()
	data, err := result.Raw()
	if err != nil {
		log.Logger.Warnf("get node data from request error, %#v, err %s", p, err.Error())
		return "", err
	}

	return string(data), nil
}

func (p *NodeGetter) GetNodeYAML(traceId string) (string, error) {
	log.Logger.Infof("start to GetNodeYAML,traceId=%s", traceId)
	var (
		content    string
		expiration = time.Duration(config.OpsConfig.RedisConfig.Expiration) * time.Millisecond
	)
	key := keys.NodeYamlKey(p.Name)
	if givenData, err := p.RedisClient.Get(key); err == nil && givenData != "" {
		content = givenData.(string)
		log.Logger.Infof("get node YAML from redis success, key=%s, content=%s,traceId=%s", key, content, traceId)
		return content, nil
	}
	resp, err := p.GetNodeFromApiServer()
	if err != nil {
		log.Logger.Warnf("get node YAML from apiserver failed, as %s,traceId=%s", err.Error(), traceId)
		return "", err
	}

	if p.RedisClient.Set(key, resp, expiration); err != nil {
		log.Logger.Warnf("failed to set node row data to redis, as err=%s, data=%s,traceId=%s",
			err.Error(), resp, traceId)
	}
	log.Logger.Infof("end to GetNodeYAML,traceId=%s", traceId)
	return content, nil
}
