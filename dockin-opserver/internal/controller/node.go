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

package controller

import (
	"net/http"
	"strings"

	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
)

type NodeController struct {
	redisClient *redis.RedisClient
	cm          *client.Manager
}

func NewNodeController(cm *client.Manager, r *redis.RedisClient) *NodeController {
	node := &NodeController{}
	node.cm = cm
	node.redisClient = r
	http.HandleFunc("/v1/dockin/opserver/node/getNodeByName", node.GetNodeByName)
	http.HandleFunc("/v1/dockin/opserver/node/batchGetNode", node.BatchGetNode)

	return node
}

func (n *NodeController) GetNodeByName(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	nodeName := request.FormValue("nodeName")
	taskId := request.FormValue("taskId")
	log.Logger.Infof("get node by name nodeName=%s, taskId=%s", nodeName, taskId)

	key := keys.NodeYamlKey(nodeName)
	data, err := n.redisClient.Get(key)
	if err != nil {
		log.Logger.Warnf("get node by name from redis failed, key=%s, err=%v", key, err)
		resp := &Response{Code: RedisFail, Msg: err.Error(), TaskId: taskId}
		writer.Write(resp.ToJSONBytes())
	}

	resp := &Response{Code: Success, Msg: "success", Data: data.(string), TaskId: taskId}
	writer.Write(resp.ToJSONBytes())

	log.Logger.Infof("success get node by name=%s, taskId=%s", nodeName, taskId)
}

func (n *NodeController) BatchGetNode(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	nodeName := request.FormValue("nodeName")
	taskId := request.FormValue("taskId")
	log.Logger.Infof("batch get node by name nodeName=%s, taskId=%s", nodeName, taskId)
	nodeNameList := strings.Split(nodeName, ",")
	var keyList []string
	for _, name := range nodeNameList {
		keyList = append(keyList, keys.NodeYamlKey(name))
	}

	nodeList, err := n.redisClient.PipelineGet(keyList)
	if err != nil {
		log.Logger.Warnf("batch node by name from redis failed, key=%v, err=%v", keyList, err)
		resp := &Response{Code: RedisFail, Msg: err.Error(), TaskId: taskId}
		writer.Write(resp.ToJSONBytes())
	}
	resp := &Response{Code: Success, Msg: "success", Data: nodeList, TaskId: taskId}
	writer.Write(resp.ToJSONBytes())

	log.Logger.Infof("success batch get node by name=%s, taskId=%s", nodeName, taskId)
}
