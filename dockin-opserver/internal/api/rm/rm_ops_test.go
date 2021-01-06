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

package rm

import (
	"go/parser"
	"testing"

	"github.com/webankfintech/dockin-opserver/internal/dockin"
	"github.com/stretchr/testify/assert"
)

func Test_GetPodInfo(t *testing.T) {
	traceId := parser.Trace
	t.Run("get pod info by host", func(t *testing.T) {
		data, err := dockin.GetPodInfoByPodIp("192.168.1.1")
		assert.NoError(t, err)
		t.Log(data)
	})

	t.Run("get pod info by subsystem", func(t *testing.T) {
		data, err := dockin.GetPodInfoBySubsystem("dockin-QQ", "", traceId)
		assert.NoError(t, err)
		t.Log(data)
	})

	t.Run("get pod info by subsystem and dcn", func(t *testing.T) {
		data, err := dockin.GetPodInfoBySubsystem("dockin-QQ", "ND1", traceId)
		assert.NoError(t, err)
		t.Log(data)
	})
}
