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

	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"

	"github.com/stretchr/testify/assert"
)

var (
	RClient *redis.RedisClient
	err     error
)

func init() {
	RClient, err = redis.NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}
}

func TestGetPods(t *testing.T) {
	m := client.NewManager(RClient)
	m.Initialize()
	echo := &Echo{Cm: m}

	t.Run("func get pod without pod", func(t *testing.T) {
		res := echo.batchGetResource("127.0.0.1", &model.OpsOption{
			Resource:  "pods",
			Rule:      "tctp",
			PrintType: "wide",
		})

		assert.Equal(t, 0, res.Code)
		t.Log(res.ToString())
	})

	t.Run("func get pod with pod", func(t *testing.T) {
		res := echo.getResource("127.0.0.1", &model.OpsOption{
			Resource:  "pods",
			Name:      "bcces-cls-20190828-160144420-0",
			Rule:      "tctp",
			PrintType: "wide",
		}, trace.TraceID())

		assert.Equal(t, 0, res.Code)
		t.Log(res.ToString())
	})

	t.Run("func get pod with pod, with namespace", func(t *testing.T) {
		res := echo.getResource("127.0.0.1", &model.OpsOption{
			Resource:  "pods",
			Name:      "bcces-cls-20190828-160144420-0",
			Rule:      "tctp",
			PrintType: "wide",
			Namespace: "xxx",
		}, trace.TraceID())

		assert.Equal(t, -1, res.Code)
		t.Log(res.ToString())
	})

}
