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

package model

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

var errNotObject = fmt.Errorf("object does not implement the Object interfaces")

type PodWide struct {
	Name              string            `json:"name"`
	Ready             string            `json:"ready"`
	Status            string            `json:"status"`
	Restarts          float64           `json:"restarts""`
	Age               string            `json:"age"`
	IP                string            `json:"ip"`
	Node              string            `json:"node"`
	NominatedNodeName string            `json:"nominatedNodeName"`
	Namespace         string            `json:"namespace"`
	Labels            map[string]string `json:"labels"`
	UID               string            `json:"uid"`
	ClusterID         string            `json:""clusterId`
}

func (p *PodWide) ToJSONString() string {
	d, _ := jsoniter.MarshalToString(p)
	return d
}

func V1Table2PodWide(table *metav1beta1.Table, clusterId string) []*PodWide {
	var wides []*PodWide
	for _, row := range table.Rows {
		pw := &PodWide{
			Name:              row.Cells[0].(string),
			Ready:             row.Cells[1].(string),
			Status:            row.Cells[2].(string),
			Restarts:          row.Cells[3].(float64),
			Age:               row.Cells[4].(string),
			IP:                row.Cells[5].(string),
			Node:              row.Cells[6].(string),
			NominatedNodeName: row.Cells[7].(string),
			ClusterID:         clusterId,
		}
		un := unstructured.Unstructured{}
		if err := un.UnmarshalJSON(row.Object.Raw); err == nil {
			pw.Namespace = un.GetNamespace()
			pw.Labels = un.GetLabels()
			pw.UID = (string)(un.GetUID())
		}
		wides = append(wides, pw)
	}

	return wides
}

func Accessor(obj interface{}) (metav1.Object, error) {
	switch t := obj.(type) {
	case metav1.Object:
		return t, nil
	case metav1.ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotObject
	default:
		return nil, errNotObject
	}
}

type NodeWide struct {
	Name             string `json:"name"`
	Status           string `json:"status"`
	Roles            string `json:"roles"`
	Age              string `json:"age"`
	Version          string `json:"version"`
	InternalIP       string `json:"Internal-IP"`
	ExternalIP       string `json:"External-IP"`
	OSImage          string `json:"osImage"`
	KernelVersion    string `json:"kernelVersion"`
	ContainerRuntime string `json:"containerRuntimeVersion"`
	ClusterID        string `json:"clusterId"`
}

func V1Table2NodeWide(table *metav1.Table, clusterId string) []*NodeWide {
	var wides []*NodeWide
	for _, row := range table.Rows {
		pw := &NodeWide{
			Name:             row.Cells[0].(string),
			Status:           row.Cells[1].(string),
			Roles:            row.Cells[2].(string),
			Age:              row.Cells[3].(string),
			Version:          row.Cells[4].(string),
			InternalIP:       row.Cells[5].(string),
			ExternalIP:       row.Cells[6].(string),
			OSImage:          row.Cells[7].(string),
			KernelVersion:    row.Cells[8].(string),
			ContainerRuntime: row.Cells[9].(string),
			ClusterID:        clusterId,
		}
		wides = append(wides, pw)
	}

	return wides
}

func (n *NodeWide) ToJSONString() string {
	d, _ := jsoniter.MarshalToString(n)
	return d
}
