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

package informer

import (
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/log"

	jsoniter "github.com/json-iterator/go"
	v1 "k8s.io/api/core/v1"
)

// NodeInformer used to watch the node event, send from apiserver
type NodeInformer struct {
	RedisClient *redis.RedisClient
}

// AddFunc watch for node add event
func (n *NodeInformer) AddFunc(obj interface{}) {
	node := obj.(*v1.Node)
	log.Logger.Debugf("add node nodeName:%s,UID:%s", node.Name, node.UID)
	nodeInfo, _ := jsoniter.Marshal(node)
	key := keys.NodeKey(node.Name)

	log.Logger.Debugf("add node Set key:%s,value:%s", key, string(nodeInfo))
	err := n.RedisClient.Set(key, string(nodeInfo), 0)
	if err != nil {
		log.Logger.Warnf("add node Set key:%s err=%s", key, err.Error())
	}

	log.Logger.Debugf("Set node key:%s success", key)
}

// UpdateFunc watch for node update event
func (n *NodeInformer) UpdateFunc(oldObj, newObj interface{}) {
	node := newObj.(*v1.Node)
	oldNode := oldObj.(*v1.Node)
	log.Logger.Debugf("node update,oldNodeName:%s,Uid:%s、newNodeName:%s,Uid:%s",
		oldNode.Name, oldNode.UID, node.Name, node.UID)

	nodeInfo, _ := jsoniter.Marshal(node)
	key := keys.NodeKey(node.Name)

	log.Logger.Debugf("node update Set key:%s,value:%s", key, string(nodeInfo))
	err := n.RedisClient.Set(key, string(nodeInfo), 0)
	if err != nil {
		log.Logger.Warnf("Set key:%s err=%s", key, err.Error())
	}

	log.Logger.Debugf("Set new node key:%s success", key)

}

// DeleteFunc watch for node delete event
func (n *NodeInformer) DeleteFunc(obj interface{}) {
	node := obj.(*v1.Node)
	log.Logger.Debugf("delete node:%s,Uid:%s", node.Name, node.UID)

	key := keys.NodeKey(node.Name)
	err := n.RedisClient.Del(key)
	if err != nil {
		log.Logger.Warnf("Del key:%s err=%s", key, err.Error())
	}

	log.Logger.Debugf("Del node key:%s success", key)
}
