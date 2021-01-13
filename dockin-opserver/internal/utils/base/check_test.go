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

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSubSystemId(t *testing.T) {
	t.Run("right system id", func(t *testing.T) {
		input := "0000"
		assert.True(t, IsSubSystemId(input))
	})

	t.Run("right system id 0", func(t *testing.T) {
		input := "0"
		assert.True(t, IsSubSystemId(input))
	})

	t.Run("with system name", func(t *testing.T) {
		input := "525dockin-wx"
		assert.False(t, IsSubSystemId(input))
	})

	t.Run("with mixture name", func(t *testing.T) {
		input := "dockin-wx"
		assert.False(t, IsSubSystemId(input))
	})
}

func TestIsPodSet(t *testing.T) {
	t.Run("given podName", func(t *testing.T) {
		assert.False(t, IsPodSet("xxx"))
	})
	t.Run("test pod set name", func(t *testing.T) {
		assert.True(t, IsPodSet("xxx"))
	})
}

func TestIsPodName(t *testing.T) {
	t.Run("given podName", func(t *testing.T) {
		assert.True(t, IsPodName("xxx"))
	})
	t.Run("test pod set name", func(t *testing.T) {
		assert.False(t, IsPodName("xxx"))
	})
	t.Run("test pod set name", func(t *testing.T) {
		assert.False(t, IsPodName("xxx"))
	})
}
