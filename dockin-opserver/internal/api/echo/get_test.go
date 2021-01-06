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

package echo

import (
	"testing"

	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/url"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestGetOps_GetPod(t *testing.T) {
	m := client.NewManager(RClient)
	m.Initialize()
	assert.Equal(t, 2, len(m.ProxyIpRuleClusterMap), "has tctp and cnc")

	pc, _ := m.GetProxyClient("127.0.0.1", "default", "cls-n20obugi")
	getops := &GetOps{
		ProxyClient: pc,
		RedisClient: RClient,
	}

	t.Run("get pod without without podname", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			PrintType: "yaml",
			Namespace: "dockin",
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get exist pod with default print", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Name:      "dockin-test-20190321-152445825",
			Namespace: "dockin",
			Container: "dockin-test-20190321-152445825",
		})
		assert.NoError(t, err)
		t.Log(resp)
	})
	t.Run("get exist pod with default wide print", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Name:      "dockin-test-20190321-152445825",
			Namespace: "dockin",
			Container: "dockin-test-20190321-152445825",
			PrintType: "wide",
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get exist pod with yaml print", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Name:      "dockin-test-20190321-152445825",
			Namespace: "dockin",
			Container: "dockin-test-20190321-152445825",
			PrintType: "yaml",
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get all pod wide", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Namespace: "dockin",
			PrintType: "wide",
			Params: map[string]interface{}{
				"all": true,
			},
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get all pod yaml", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Namespace: "dockin",
			PrintType: "yaml",
			Params: map[string]interface{}{
				"all": true,
			},
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get not exist pod with yaml print", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Name:      "not-exist",
			Namespace: "dockin",
			Container: "not-exist",
		})
		assert.Error(t, err)
		t.Log(string(resp.ToByte()))
	})

	t.Run("get pods in all namespace", func(t *testing.T) {

		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "pods",
			Name:      "",
			Namespace: "",
			Container: "",
			PrintType: "wide",
		})
		assert.NoError(t, err)
		t.Log(resp.Data.(string))
	})
}

func TestGetOps_GetNode(t *testing.T) {
	m := client.NewManager(RClient)
	m.Initialize()
	assert.Equal(t, 2, len(m.ProxyIpRuleClusterMap), "has tctp and cnc")

	pc, _ := m.GetProxyClient("127.0.0.1", "tctp", "tctp")
	getops := &GetOps{
		ProxyClient: pc,
	}

	t.Run("get node", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource: "nodes",
			Name:     "192-168-1-74",
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get all node yaml", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "nodes",
			Name:      "192-168-1-74",
			Namespace: "dockin",
			PrintType: "yaml",
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get all node", func(t *testing.T) {
		resp, err := getops.GetResource(&model.OpsOption{
			Resource:  "nodes",
			Namespace: "dockin",
			Params: map[string]interface{}{
				"all": true,
			},
		})
		assert.NoError(t, err)
		t.Log(resp)
	})

	t.Run("get url encode", func(t *testing.T) {
		opt := &model.OpsOption{
			Resource:  "nodes",
			Namespace: "dockin",
			Params: map[string]interface{}{
				"all": true,
			},
		}
		d, err := jsoniter.MarshalToString(opt)
		assert.NoError(t, err)
		t.Log(d)
		t.Log(url.UrlEncode(d))
	})
}
