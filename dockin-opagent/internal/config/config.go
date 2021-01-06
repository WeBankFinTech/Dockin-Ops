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

	"github.com/webankfintech/dockin-opagent/internal/log"
	"gopkg.in/yaml.v2"
)

const fileType = "yaml"
const prefix = "application"

func init() {
	cfg, err := NewProxyConfig()
	if err != nil {
		log.Logger.Panicf(err.Error())
	}
	AgentConf = cfg
}

var (
	AgentConf *AgentConfig
	once      sync.Once
)

type AgentConfig struct {
	App struct {
		Env string `yaml:"env"`
		Rm  struct {
			API string `yaml:"api"`
		} `yaml:"rm"`
		Container struct {
			Ticker int `yaml:"ticker"`
		} `yaml:"container"`
		HTTP struct {
			Port int `yaml:"port"`
		} `yaml:"http"`
		Debug struct {
			Port int `yaml:"port"`
		} `yaml:"debug"`
		Ims struct {
			Logroot string `yaml:"logroot"`
		} `yaml:"ims"`
		Docker struct {
			Sock string `yaml:"sock"`
		} `yaml:"docker"`
		Qos struct {
			Path string `yaml:"path"`
		} `yaml:"qos"`
		Logs struct {
			CmdWhiteList []string `yaml:"cmd-white-list"`
			CmdTimeout   int      `yaml:"cmd-timeout"`
			MaxFileSize  int      `yaml:"max-file-size"`
			MaxLine      int      `yaml:"max-line"`
			Root         string   `yaml:"root"`
		} `yaml:"logs"`
		OOMKmsgPath string `yaml:"oom-kmsg-path"`
		Kubedebug struct{
			User         string   `yaml:"user"`
			Passwd		 string	  `yaml:"passwd"`
		} `yaml:"kubedebug"`
	} `yaml:"app"`
}

func NewProxyConfig() (*AgentConfig, error) {
	agentConfigPath := filepath.Join(GetConfPath(), prefix + "." + fileType)
	log.Logger.Infof("create config from file path=%s", agentConfigPath)

	content, err := ioutil.ReadFile(agentConfigPath)
	if err != nil {
		log.Logger.Warnf("ReadFile file:%s err=%s", agentConfigPath, err.Error())
		return nil, err
	}
	cfg := &AgentConfig{}
	if err = yaml.Unmarshal(content, cfg); err != nil {
		log.Logger.Warnf("yaml.Unmarshal err=%s", err.Error())
		return nil, err
	}

	log.Logger.Infof("success load config, %s", string(content))
	return cfg, nil
}
