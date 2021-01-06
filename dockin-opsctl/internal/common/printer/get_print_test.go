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

package printer

import "testing"

func TestPrettyPrint(t *testing.T) {
	content := []byte(`{
  "kind": "Table",
  "apiVersion": "meta.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/api/v1/namespaces/dockin/pods/dockin-test-test",
    "resourceVersion": "61650911"
  },
  "columnDefinitions": [
    {
      "name": "Name",
      "type": "string",
      "format": "name",
      "description": "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
      "priority": 0
    },
    {
      "name": "Ready",
      "type": "string",
      "format": "",
      "description": "The aggregate readiness state of this pod for accepting traffic.",
      "priority": 0
    },
    {
      "name": "Status",
      "type": "string",
      "format": "",
      "description": "The aggregate status of the containers in this pod.",
      "priority": 0
    },
    {
      "name": "Restarts",
      "type": "integer",
      "format": "",
      "description": "The number of times the containers in this pod have been restarted.",
      "priority": 0
    },
    {
      "name": "Age",
      "type": "string",
      "format": "",
      "description": "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
      "priority": 0
    },
    {
      "name": "IP",
      "type": "string",
      "format": "",
      "description": "IP address allocated to the pod. Routable at least within the cluster. Empty if not yet allocated.",
      "priority": 1
    },
    {
      "name": "Node",
      "type": "string",
      "format": "",
      "description": "NodeName is a request to schedule this pod onto a specific node. If it is non-empty, the scheduler simply schedules this pod onto that node, assuming that it fits resource requirements.",
      "priority": 1
    },
    {
      "name": "Nominated Node",
      "type": "string",
      "format": "",
      "description": "nominatedNodeName is set only when this pod preempts other pods on the node, but it cannot be scheduled right away as preemption victims receive their graceful termination periods. This field does not guarantee that the pod will be scheduled on this node. Scheduler may decide to place the pod elsewhere if other nodes become available sooner. Scheduler may also decide to give the resources on this node to a higher priority pod that is created after preemption. As a result, this field may be different than PodSpec.nodeName when the pod is scheduled.",
      "priority": 1
    }
  ],
  "rows": [
    {
      "cells": [
        "dockin-test-test",
        "1/1",
        "Running",
        0,
        "45h",
        "192.168.1.43",
        "192-168-1-195",
        "<none>"
      ],
      "object": {
        "kind": "PartialObjectMetadata",
        "apiVersion": "meta.k8s.io/v1beta1",
        "metadata": {
          "name": "dockin-test-test",
          "namespace": "dockin",
          "selfLink": "/api/v1/namespaces/dockin/pods/dockin-test-test",
          "uid": "0722e2d1-b1fa-11e9-8b47-525400cdf4e8",
          "resourceVersion": "61650911",
          "creationTimestamp": "2019-07-29T12:11:43Z",
          "labels": {
            "dcn": "AB0",
            "deployFailureKeyWord": "failure",
            "deployResultPath": "path.data.appsystems.dockin-test.bin.deploy_result",
            "deploySuccessKeyWord": "success",
            "subsystem": "dockin-test"
          },
          "annotations": {
            "network-gateway": "192.168.1.1",
            "network-ip": "192.168.1.2",
            "network-mask": "255.255.255.192"
          }
        }
      }
    }
  ]
}`)
	PrettyPrint(content)
}

func TestPrintTableNamespace(t *testing.T) {
	content := []byte(`{
  "kind": "Table",
  "apiVersion": "meta.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/api/v1/namespaces/dockin",
    "resourceVersion": "139866"
  },
  "columnDefinitions": [
    {
      "name": "Name",
      "type": "string",
      "format": "name",
      "description": "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
      "priority": 0
    },
    {
      "name": "Status",
      "type": "string",
      "format": "",
      "description": "The status of the namespace",
      "priority": 0
    },
    {
      "name": "Age",
      "type": "string",
      "format": "",
      "description": "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.\n\nPopulated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
      "priority": 0
    }
  ],
  "rows": [
    {
      "cells": [
        "dockin",
        "Active",
        "244d"
      ],
      "object": {
        "kind": "PartialObjectMetadata",
        "apiVersion": "meta.k8s.io/v1beta1",
        "metadata": {
          "name": "dockin",
          "selfLink": "/api/v1/namespaces/dockin",
          "uid": "d6f659ec-f308-11e8-9b4c-525400f6e58d",
          "resourceVersion": "139866",
          "creationTimestamp": "2018-11-28T12:26:32Z"
        }
      }
    }
  ]
}`)
	PrettyPrint(content)
}
