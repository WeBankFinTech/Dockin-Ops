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

package config

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/log"

	"gopkg.in/yaml.v2"
)

type ProxyConfig struct {
	ENV                 string `yaml:"env"`
	RMAddress           string `yaml:"rm-address"`
	BatchTimeout        int64  `yaml:"batch-timeout"`
	HttpPort            int32  `yaml:"http-port"`
	CmdFilterType       string `yaml:"cmd-filter-type"`
	WhileListUpdateTime int64  `yaml:"while-list-update-time"`
	Limits              struct {
		ExecForbidden []string `yaml:"exec-forbidden"`
		ViFileMaxSize int64    `yaml:"vi-file-max-size"`
		K8SQOS        int32    `yaml:"k8s-qos"`
		K8SBurst      int32    `yaml:"k8s-burst"`
	} `yaml:"limits"`
	OpAgentPort int32 `yaml:"opagent-port"`
	RedisConfig struct {
		Expiration int64 `yaml:"expiration"`
	} `yaml:"redis"`
	Accounts []struct {
		Account struct {
			UserName string `yaml:"user-name"`
			Passwd   string `yaml:"passwd"`
		} `yaml:"account"`
	} `yaml:"accounts"`
	Devops struct {
		Address string `yaml:"address"`
		Dir     string `yaml:"dir"`
	} `yaml:"devops"`
	Debug struct {
		Image string `yaml:"image"`
	} `yaml:"debug"`
}

var (
	OpsConfig *ProxyConfig
	once      sync.Once
)

func init() {
	if _, err := NewProxyConfig(); err != nil {
		log.Logger.Panicf(err.Error())
	}
}

func NewProxyConfig() (*ProxyConfig, error) {
	confFile := filepath.Join(common.GetConfPath(), "application.yaml")
	content, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Logger.Errorf("read config file %s, error %s", confFile, err.Error())
		return nil, err
	}
	pc := &ProxyConfig{}
	if err := yaml.Unmarshal(content, pc); err != nil {
		log.Logger.Errorf("unmarshal config file %s, error %s", confFile, err.Error())
		return nil, err
	}
	OpsConfig = pc
	return pc, nil
}
