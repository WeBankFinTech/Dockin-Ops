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

import jsoniter "github.com/json-iterator/go"

type K8sConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify"`
			Server                string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster   string `yaml:"cluster"`
			Namespace string `yaml:"namespace"`
			User      string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			Password string `yaml:"password"`
			Username string `yaml:"username"`
		} `yaml:"user"`
	} `yaml:"users"`
	Dockin struct {
		ClusterID string `yaml:"cluster-id"`
		Rule      string `yaml:"rule"`
	} `yaml:"dockin"`
}

func (k *K8sConfig) ToString() string {
	str, _ := jsoniter.MarshalToString(k)
	return str
}
