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

package utils

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func Test_unTar(t *testing.T)  {
	file := "C:\\MyWork\\securecrt\\rpd_tools.tar"
	fw,_ := os.Open(file)
	UntarFile(fw,"C:\\MyWork\\securecrt")
}

func Test_tarFile(t *testing.T)  {
	 if err := tarFile("C:\\MyWork\\MYpython\\app", "C:\\MyWork\\MYpython\\app.tar");err != nil{
	 	fmt.Println(err.Error())
	 }
}

func Test_tarFileV(t *testing.T)  {
	t.Run("tar", func(t *testing.T) {
		src := filepath.Join("C:\\", "download", "tmp", "hzf")
		f, err := os.Create(filepath.Join("C:\\", "download", "tmp", "test.tar"))
		assert.NoError(t, err)
		bio := bufio.NewWriter(f)
		err = MakeTarV2(src, bio)
		defer bio.Flush()
		assert.NoError(t, err)
	})
}