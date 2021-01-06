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

	"github.com/webankfintech/dockin-opserver/internal/utils/cmap"

	"k8s.io/client-go/tools/clientcmd/api"

	"os"
	"path/filepath"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/informer"
	"github.com/webankfintech/dockin-opserver/internal/log"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ListWatcher struct {
	podInformer   *informer.PodInformer
	nodeInformer  *informer.NodeInformer
	eventInformer *informer.EventInformer
	redisClient   *redis.RedisClient
}

func NewListWatcher(redisClient *redis.RedisClient) *ListWatcher {
	return &ListWatcher{
		podInformer:   &informer.PodInformer{RedisClient: redisClient, HttpMap: cmap.New()},
		nodeInformer:  &informer.NodeInformer{RedisClient: redisClient},
		eventInformer: &informer.EventInformer{},
		redisClient:   redisClient,
	}
}

func (w *ListWatcher) Initialize(stoper chan struct{}) error {
	confPath := common.GetConfPath()
	clusterPath := filepath.Join(confPath, "cluster")
	var filelist []string
	walker := func(path string) {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				filelist = append(filelist, path)
			}
			return nil
		})
	}
	walker(clusterPath)
	for _, k8sfile := range filelist {
		log.Logger.Infof("initialize with k8sfile:%s", k8sfile)
		yamlbyte, err := ioutil.ReadFile(k8sfile)
		if err != nil {
			log.Logger.Warnf("read yaml file %s failed, as %s", k8sfile, err.Error())
			continue
		}
		config, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (config *api.Config, e error) {
			config, err := clientcmd.Load([]byte(yamlbyte))
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

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Logger.Warnf("create kubenetes client set failed, err=%s", err.Error())
			continue
		}

		factory := informers.NewSharedInformerFactory(clientset, time.Minute)
		podInformer := factory.Core().V1().Pods().Informer()
		nodeInformer := factory.Core().V1().Nodes().Informer()
		eventInformer := factory.Core().V1().Events().Informer()
		podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    w.podInformer.AddFunc,
			UpdateFunc: w.podInformer.UpdateFunc,
			DeleteFunc: w.podInformer.DeleteFunc,
		})

		nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    w.nodeInformer.AddFunc,
			UpdateFunc: w.nodeInformer.UpdateFunc,
			DeleteFunc: w.nodeInformer.DeleteFunc,
		})

		eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    w.eventInformer.AddFunc,
			UpdateFunc: w.eventInformer.UpdateFunc,
			DeleteFunc: w.eventInformer.DeleteFunc,
		})

		go func() {
			podInformer.Run(stoper)
			log.Logger.Infof("stop pod listener k8sFile:%s", k8sfile)
		}()

		go func() {
			nodeInformer.Run(stoper)
			log.Logger.Infof("stop node listener k8sFile:%s", k8sfile)
		}()

		go func() {
			eventInformer.Run(stoper)
			log.Logger.Infof("stop event listener k8sFile:%s", k8sfile)
		}()
	}

	return nil
}
