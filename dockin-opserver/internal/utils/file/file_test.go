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

package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractFileSpec(t *testing.T) {
	path := `tctp/dockin-test:/tmp/foo`
	spec, err := extractFileSpec(path)
	assert.NoError(t, err)
	t.Log(spec)
}

func Test_Untar(t *testing.T) {
	file := "C:\\download\\tmp\\hzfllz\\wcsctl.tar"
	fr, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fr.Close()
	UntarFile(fr, "C:\\download\\tmp\\hzfllz")
}
