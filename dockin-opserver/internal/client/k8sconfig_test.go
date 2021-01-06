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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/webankfintech/dockin-opserver/internal/common"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestParseYaml(t *testing.T) {
	confPath := common.GetConfPath()
	kubeConfigPath := filepath.Join(confPath, "dockin.yaml")

	content, err := ioutil.ReadFile(kubeConfigPath)
	assert.NoError(t, err)
	kube := &K8sConfig{}
	err = yaml.Unmarshal(content, kube)
	assert.NoError(t, err)
	t.Log(kube.Contexts[0].Context.Namespace)

}

func TestPrdTCTPYaml(t *testing.T) {
	tctp := []byte(`apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    namespace: test
    user: kubernetes-readonly-user
  name: readonly-user
current-context: readonly-user
kind: Config
preferences: {}
users:
- name: kubernetes-readonly-user
  user:
    password: 
    username: readonly-user
dockin:
  cluster-id: ft01
  rule: test
  whitelist:
    - 127.0.0.1`)
	kube := &K8sConfig{}
	err := yaml.Unmarshal(tctp, kube)
	assert.NoError(t, err)
	t.Log(kube.Contexts[0].Context.Namespace)
}

func TestRunPrdCheck(t *testing.T) {
	var (
		filelist []string
	)

	confPath := common.GetConfPath()
	clusterPath := filepath.Join(confPath, "env", "prd", "cluster")

	appendK8sConfig := func(path string) {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				filelist = append(filelist, path)
			}
			return nil
		})
	}

	appendK8sConfig(clusterPath)
	t.Logf("walk conf path %s, got %#v k8s config file", clusterPath, filelist)

	for _, k8sfile := range filelist {
		yamlbyte, err := ioutil.ReadFile(k8sfile)
		if err != nil {
			t.Logf("read yaml file %s failed, as %s", k8sfile, err.Error())
			return
		}
		ky := &K8sConfig{}
		if err = yaml.Unmarshal(yamlbyte, ky); err != nil {
			t.Logf("unmarshal yaml file %s failed, as %s", k8sfile, err.Error())
			return
		}
		t.Logf("file=%s, namespace=%s", k8sfile, ky.Contexts[0].Context.Namespace)
	}
}
