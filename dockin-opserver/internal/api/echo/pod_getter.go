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
	"strings"
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

type PodGetter struct {
	ProxyClient  *client.ProxyClient
	RedisClient  *redis.RedisClient
	AllNamespace bool
	Namespace    string
	Name         string
	PrintType    string
	ClusterId    string
	Rule         string
}

func NewPodGetter(echo *model.OpsOption, pc *client.ProxyClient, rc *redis.RedisClient, pt string) *PodGetter {
	pg := &PodGetter{}
	if val, exist := echo.Params["all-namespaces"]; exist {
		pg.AllNamespace = val.(bool)
	}
	if !pg.AllNamespace && echo.Namespace != "" {
		echo.Namespace = pc.K8sConfig.Contexts[0].Context.Namespace
		pg.Namespace = echo.Namespace
	}
	pg.Name = echo.Name
	pg.ProxyClient = pc
	pg.RedisClient = rc
	pg.PrintType = pt
	pg.ClusterId = echo.ClusterId
	pg.Rule = echo.Rule
	return pg
}

func (p *PodGetter) GetPod(traceId string) (string, error) {
	if p.Name == "" {
		return "", errors.Errorf("get pod yaml must specify a pod name")
	}

	if p.PrintType == "wide" {
		return p.GetPodWide(traceId)
	}
	return p.GetPodYAML(traceId)
}

func (p *PodGetter) GetPodWide(traceId string) (string, error) {
	log.Logger.Infof("start to GetPodWide,traceId=%s", traceId)
	var (
		content    string
		expiration = time.Duration(config.OpsConfig.RedisConfig.Expiration) * time.Millisecond
	)
	key := keys.PodWideKey(p.Name)
	if givenData, err := p.RedisClient.Get(key); err == nil && givenData != "" {
		content = givenData.(string)
		log.Logger.Infof("get pod wide from redis success, key=%s, content=%s,traceId=%s", key, content, traceId)
		pw := &model.PodWide{}
		if err = jsoniter.UnmarshalFromString(content, pw); err != nil {
			log.Logger.Warnf("get pod wide from redis not a pod wide struct, key=%s, content=%s, err=%s,traceId=%s",
				key, content, err.Error(), traceId)
		} else {
			tmp, err := jsoniter.MarshalToString([]*model.PodWide{pw})
			if err != nil {
				log.Logger.Warnf("get pod wide from redis not a pod wide struct, key=%s, content=%s, err=%s,traceId=%s",
					key, content, err.Error(), traceId)
			} else {
				return tmp, nil
			}
		}
	}

	log.Logger.Infof("no redis pod wide data exist, get from apiserver, podName=%s,traceId=%s", p.Name, traceId)
	resp, err := p.GetPodFromApiServer()
	if err != nil {
		log.Logger.Warnf("get pod wide from apiserver failed, as %s,traceId=%s", err.Error(), traceId)
		return "", err
	}
	log.Logger.Infof("get pod wide info from apiserver, podName=%s, content=%s,traceId=%s", p.Name, resp, traceId)

	td := &metav1beta1.Table{}
	if err := jsoniter.Unmarshal([]byte(resp), td); err != nil {
		log.Logger.Warnf("failed to unmarshal response data from apiserver failed, as %s, data=%s,traceId=%s",
			err.Error(), resp, traceId)
		return "", err
	}

	pws := model.V1Table2PodWide(td, p.ClusterId)
	if content, err = jsoniter.MarshalToString(pws); err != nil {
		log.Logger.Warnf("failed to unmarshal apiserver data to row data, as err=%s, data=%s,traceId=%s",
			err.Error(), content, traceId)
		return "", err
	}

	if p.RedisClient.Set(key, content, expiration); err != nil {
		log.Logger.Warnf("failed to set row data to redis, as err=%s, data=%s,traceId=%s",
			err.Error(), pws, traceId)
	}
	log.Logger.Infof("end to GetPodWide,traceId=%s", traceId)
	return content, nil
}

func (p *PodGetter) GetPodFromApiServer() (string, error) {
	req := p.ProxyClient.ApiClient.CoreV1().RESTClient().Get().
		Resource("pods").
		VersionedParams(&metav1.GetOptions{ResourceVersion: "0"}, scheme.ParameterCodec)
	if p.PrintType == "wide" || p.PrintType == "" {
		req.SetHeader("Accept", tabHeader)
	} else if p.PrintType == "yaml" {
		req.SetHeader("Accept", yamlHeader)
	}
	if !p.AllNamespace {
		if p.Name != "" {
			req.Name(p.Name)
		}
		req.Namespace(strings.ToLower(p.Namespace))
	}
	result := req.Do()
	data, err := result.Raw()
	if err != nil {
		log.Logger.Warnf("get data from request error, %#v, err %s", p, err.Error())
		return "", err
	}

	return string(data), nil
}

func (p *PodGetter) GetPodYAML(traceId string) (string, error) {
	log.Logger.Infof("start to GetPodYAML,traceId=%s", traceId)
	var (
		content    string
		expiration = time.Duration(config.OpsConfig.RedisConfig.Expiration) * time.Millisecond
	)
	key := keys.PodYAMLKey(p.Name)
	if givenData, err := p.RedisClient.Get(key); err == nil && givenData != "" {
		content = givenData.(string)
		log.Logger.Infof("get pod YAML from redis success, key=%s, content=%s,traceId=%s", key, content, traceId)
		return content, nil
	}
	resp, err := p.GetPodFromApiServer()
	if err != nil {
		log.Logger.Warnf("get pod YAML from apiserver failed, as %s,traceId=%s", err.Error(), traceId)
		return "", err
	}

	if p.RedisClient.Set(key, resp, expiration); err != nil {
		log.Logger.Warnf("failed to set row data to redis, as err=%s, data=%s,traceId=%s",
			err.Error(), resp, traceId)
	}
	log.Logger.Infof("end to GetPodYAML,traceId=%s", traceId)
	return content, nil
}
