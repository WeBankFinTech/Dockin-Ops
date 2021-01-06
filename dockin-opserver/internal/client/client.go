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

	"github.com/webankfintech/dockin-opserver/internal/log"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

var (
	ErrLoadKubeConfig 		= fmt.Errorf("load kube config from KUBECONFIG failed, check env")
	ErrCreateClientSet		= fmt.Errorf("create kubenetes client set failed")
	ErrPropsNotSet			= fmt.Errorf("check application yaml master url and kube config path")

	MasterUrlProps 			= "app.kube.masterurl"
	KubeConfigFileProps 	= "app.kube.KubeConfigFile"
)

type ProxyClient struct {
	ApiClient *kubernetes.Clientset
	RestConfig *restclient.Config
	K8sConfig	*K8sConfig
}

func NewProxyClient(rcfg *restclient.Config, k8s *K8sConfig) *ProxyClient {
	clientset, err := kubernetes.NewForConfig(rcfg)
	if err != nil {
		log.Logger.Panic(ErrCreateClientSet.Error())
	}

	return &ProxyClient{
		ApiClient: clientset,
		RestConfig:rcfg,
		K8sConfig:k8s,
	}
}