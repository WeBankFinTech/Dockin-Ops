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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Whitelist(t *testing.T) {
	wlf := NewWhitelistFilter(func() []string {
		return []string{"ls", "pwd"}
	})

	t.Run("in the list", func(t *testing.T) {
		cu, err := ParseCmdlineToCmdUnitList("ls -al")
		assert.NoError(t, err)
		err = wlf.Do(cu)
		assert.NoError(t, err)
	})

	t.Run("not in the list", func(t *testing.T) {
		cu, err := ParseCmdlineToCmdUnitList("top -p 1")
		assert.NoError(t, err)
		err = wlf.Do(cu)
		assert.Error(t, err)
	})
}

func Test_Blacklist(t *testing.T) {
	blf := NewBlacklistFilter(func() []string {
		return []string{"ls", "pwd"}
	})

	t.Run("in the list", func(t *testing.T) {
		cu, err := ParseCmdlineToCmdUnitList("ls -al")
		assert.NoError(t, err)
		err = blf.Do(cu)
		assert.Error(t, err)
	})

	t.Run("not in the list", func(t *testing.T) {
		cu, err := ParseCmdlineToCmdUnitList("top -p 1")
		assert.NoError(t, err)
		err = blf.Do(cu)
		assert.NoError(t, err)
	})
}
