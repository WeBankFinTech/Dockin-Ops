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

package aes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAes(t *testing.T) {
	key := "doc-opserver8085"
	t.Run("encrypt", func(t *testing.T) {
		aes, err := NewAes(key)
		assert.NoError(t, err)

		out, err := aes.AesEncrypt(`{"code":0,"msg":"OK", "page":{"totalPage":10,"prePage":0,"nextPage":2,"pageNum":1,"pageSize":10}}`)
		assert.NoError(t, err)
		t.Log(out)
		next, err := aes.AesDecrypt(out)
		assert.NoError(t, err)
		t.Log(next)
	})
}
