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

package remote

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParserShellPrompt(t *testing.T) {
	t.Run("test app", func(t *testing.T) {
		ps, err := ParserShellPrompt("[app@sls-service-20201020-112040385-0 ~]$ ")
		assert.NoError(t, err)
		assert.Equal(t, ps.user, "app")
		assert.Equal(t, ps.instance, "sls-service-20201020-112040385-0")
		assert.Equal(t, ps.dir, "~")
	})

	t.Run("test root", func(t *testing.T) {
		ps, err := ParserShellPrompt("[root@sls-service-20201020-112040385-0 app]# ")
		assert.NoError(t, err)
		assert.Equal(t, ps.user, "root")
		assert.Equal(t, ps.instance, "sls-service-20201020-112040385-0")
		assert.Equal(t, ps.dir, "app")
	})

	t.Run("test for root path", func(t *testing.T) {
		ps, err := ParserShellPrompt("[app@dockin-wx-20191012-180011777-0 /]$ ")
		assert.NoError(t, err)
		assert.Equal(t, ps.user, "app")
		assert.Equal(t, ps.instance, "dockin-wx-20191012-180011777-0")
		assert.Equal(t, ps.dir, "/")
	})

	t.Run("test appsystem", func(t *testing.T) {
		ps, err := ParserShellPrompt("[app@dockin-wx-20191012-180011777-0 appsystems]$ ")
		assert.NoError(t, err)
		assert.Equal(t, ps.user, "app")
		assert.Equal(t, ps.instance, "dockin-wx-20191012-180011777-0")
		assert.Equal(t, ps.dir, "appsystems")
	})

	t.Run("test appsystem extra output", func(t *testing.T) {
		ps, err := ParserShellPrompt("[app@dockin-wx-20191012-180011777-0 appsystems]$ xxxxx")
		assert.NoError(t, err)
		assert.Equal(t, ps.user, "app")
		assert.Equal(t, ps.instance, "dockin-wx-20191012-180011777-0")
		assert.Equal(t, ps.dir, "appsystems")
	})
}
