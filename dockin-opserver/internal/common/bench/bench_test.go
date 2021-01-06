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

package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConf(t *testing.T) {
	content := []byte(`{
  "concurrent": 3,
  "number": 300,
  "timeout": 300,
  "poolSize":3,
  "podList": [
    "1",
    "2",
    "3"
  ]
}`)
	pa := &param{}
	err := json.Unmarshal(content, pa)
	t.Log(pa.String())
	assert.NoError(t, err)
	assert.Equal(t, pa.Concurrent, 3)
	assert.Equal(t, pa.Timeout, int64(300))
}
