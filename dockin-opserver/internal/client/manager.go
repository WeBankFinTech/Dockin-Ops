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

package client

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd/api"

	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/utils/cmap"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
)

type Manager struct {
	ProxyIpRuleClusterMap cmap.ConcurrentMap

	ProxyClusterNSMap cmap.ConcurrentMap

	ProxyClusterUATMap cmap.ConcurrentMap
	ListenStopper      chan struct{}

	listWatcher *ListWatcher
	redisClient *redis.RedisClient

	whitelist *whitelist
}

func NewManager(rc *redis.RedisClient) *Manager {
	return &Manager{
		ProxyIpRuleClusterMap: cmap.New(),
		ProxyClusterNSMap:     cmap.New(),
		ProxyClusterUATMap:    cmap.New(),
		ListenStopper:         make(chan struct{}),
		redisClient:           rc,
		whitelist:             newWhitelist(rc),
	}
}

func (m *Manager) printCurrentProxy() {
	for k, v := range m.ProxyIpRuleClusterMap.Items() {
		log.Logger.Debugf("ProxyIpRuleClusterMap: key = %s, value=%v", k, v)
	}

	for k, v := range m.ProxyClusterNSMap.Items() {
		log.Logger.Debugf("ProxyClusterNSMap: key = %s, value=%v", k, v)
	}

	for k, v := range m.ProxyClusterUATMap.Items() {
		log.Logger.Debugf("ProxyClusterUATMap: key = %s, value=%v", k, v)
	}
}

func (m *Manager) Initialize() {
	var (
		filelist []string
	)

	confPath := common.GetConfPath()
	clusterPath := filepath.Join(confPath, "cluster")

	appendK8sConfig := func(path string) {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				filelist = append(filelist, path)
			}
			return nil
		})
	}

	if err := m.whitelist.initialize(m.ListenStopper); err != nil {
		log.Logger.Panicf("initialize whitelist failed as %s", err.Error())
	}
	appendK8sConfig(clusterPath)
	log.Logger.Infof("walk conf path %s, got %#v k8s config file", clusterPath, filelist)

	for _, k8sfile := range filelist {
		yamlbyte, err := ioutil.ReadFile(k8sfile)
		if err != nil {
			log.Logger.Warnf("read yaml file %s failed, as %s", k8sfile, err.Error())
			continue
		}
		ky := &K8sConfig{}
		if err = yaml.Unmarshal([]byte(yamlbyte), ky); err != nil {
			log.Logger.Warnf("unmarshal yaml file %s failed, as %s", k8sfile, err.Error())
			continue
		}
		if ky.Contexts[0].Context.Namespace == "" {
			log.Logger.Warnf("no namespace in context provided in %s, ignore it", k8sfile)
			continue
		}
		log.Logger.Infof("load file=%s success", k8sfile)
		m.setK8sConfigMap(ky, string(yamlbyte))
	}

	m.InitListener()
	m.printCurrentProxy()
}

func (m *Manager) InitListener() {
	m.listWatcher = NewListWatcher(m.redisClient)
	m.listWatcher.Initialize(m.ListenStopper)
}

func (m *Manager) GetProxyClient(ip, rule, clusterId string) (*ProxyClient, error) {
	if err := m.Allow(rule, ip); err != nil {
		log.Logger.Warnf(err.Error())
		return nil, err
	}

	list, err := m.GetProxyByRule(rule)
	if err != nil {
		return nil, err
	}

	var pc *ProxyClient
	for _, vv := range list {
		log.Logger.Infof("---- cluster in proxy:", vv.K8sConfig.Dockin.ClusterID)
		if strings.EqualFold(vv.K8sConfig.Dockin.ClusterID, clusterId) {
			log.Logger.Infof("get proxy found clusterId = %s proxy, namespace=%s",
				clusterId, vv.K8sConfig.Contexts[0].Context.Namespace)
			pc = vv
			break
		}
	}

	if pc == nil {
		return nil, errors.Errorf("no proxy found for ip=%s, rule=%s, clusterId=%s", ip, rule, clusterId)
	}

	return pc, nil
}

func (m *Manager) GetSshProxyClient(ip, rule, clusterId string) (*ProxyClient, error) {
	list, err := m.GetProxyByRule(rule)
	if err != nil {
		return nil, err
	}

	var pc *ProxyClient
	for _, vv := range list {
		if strings.EqualFold(vv.K8sConfig.Dockin.ClusterID, clusterId) {
			log.Logger.Infof("get proxy found clusterId = %s proxy, namespace=%s",
				clusterId, vv.K8sConfig.Contexts[0].Context.Namespace)
			pc = vv
			break
		}
	}

	if pc == nil {
		return nil, errors.Errorf("no proxy found for ip=%s, rule=%s, clusterId=%s", ip, rule, clusterId)
	}

	return pc, nil
}

func (m *Manager) GetProxyByRule(rule string) ([]*ProxyClient, error) {
	value, ok := m.ProxyIpRuleClusterMap.Get(rule)
	if !ok {
		return nil, errors.Errorf("no proxy found for rule=%s", rule)
	}
	list := value.([]*ProxyClient)

	return list, nil
}

func (m *Manager) setK8sConfigMap(k8s *K8sConfig, content string) {
	clusterId := k8s.Dockin.ClusterID
	rule := k8s.Dockin.Rule
	log.Logger.Infof("set k8s config for clusterId=%s, rule=%s", clusterId, rule)

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (config *api.Config, e error) {
		config, err := clientcmd.Load([]byte(content))
		if err != nil {
			return nil, err
		}

		for key, obj := range config.AuthInfos {
			//obj.LocationOfOrigin = filename
			config.AuthInfos[key] = obj
		}
		for key, obj := range config.Clusters {
			//obj.LocationOfOrigin = filename
			config.Clusters[key] = obj
		}
		for key, obj := range config.Contexts {
			//obj.LocationOfOrigin = filename
			config.Contexts[key] = obj
		}

		if config.AuthInfos == nil {
			config.AuthInfos = map[string]*api.AuthInfo{}
		}
		if config.Clusters == nil {
			config.Clusters = map[string]*api.Cluster{}
		}
		if config.Contexts == nil {
			config.Contexts = map[string]*api.Context{}
		}

		return config, nil
	})
	if err != nil {
		log.Logger.Panicf("build config from kubeconfig err=%v", err)
	}

	if m.ProxyIpRuleClusterMap.Has(rule) {
		value, _ := m.ProxyIpRuleClusterMap.Get(rule)
		proxylist := value.([]*ProxyClient)
		proxylist = append(proxylist, NewProxyClient(config, k8s))
		m.ProxyIpRuleClusterMap.Set(rule, proxylist)
	} else {
		proxylist := []*ProxyClient{NewProxyClient(config, k8s)}
		m.ProxyIpRuleClusterMap.Set(rule, proxylist)
	}

	namespace := k8s.Contexts[0].Context.Namespace
	if namespace == "" {
		log.Logger.Warnf("set proxy for cluster and namespace, but namespace is empty")
		return
	}
	clusterIdNSKey := strings.ToLower(fmt.Sprintf("%s:%s", clusterId, namespace))
	if m.ProxyClusterNSMap.Has(clusterIdNSKey) {
		log.Logger.Warnf("proxy for cs and ns key=%s has existed", clusterIdNSKey)
		return
	}
	m.ProxyClusterNSMap.Set(clusterIdNSKey, NewProxyClient(config, k8s))
	m.ProxyClusterUATMap.Set(strings.ToLower(clusterId), NewProxyClient(config, k8s))
	log.Logger.Infof("add cluster-namespace proxy, key=%s", clusterIdNSKey)
}

func (m *Manager) GetProxyByClusterAndNS(clusterId, namespace string) (*ProxyClient, error) {
	clusterIdNSKey := strings.ToLower(fmt.Sprintf("%s:%s", clusterId, namespace))
	if !m.ProxyClusterNSMap.Has(clusterIdNSKey) {
		return nil, errors.Errorf("no proxy found for clusterId=%s, namespace=%s",
			clusterId, namespace)
	}

	vv, _ := m.ProxyClusterNSMap.Get(clusterIdNSKey)
	return vv.(*ProxyClient), nil
}

func (m *Manager) Allow(rule, ip string) error {
	return m.whitelist.allow(rule, ip)
}

func (m *Manager) SShAllow(rule, ip, userName string) error {
	if strings.Compare(userName, common.UserName) == 0 {
		log.Logger.Infof("user is %s", common.UserName)
		return nil
	}
	return m.whitelist.allow(rule, ip)
}
