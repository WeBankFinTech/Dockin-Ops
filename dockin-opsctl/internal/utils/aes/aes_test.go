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
	"fmt"
	"testing"
)

func Test_Aes(t *testing.T) {
	biz := "doc-opserver8085"
	fmt.Println(biz)
	aes, _ := NewAes(biz)
	str1, _ := aes.AesEncrypt("{\"2\":\"efg\",\"1\":\"abc\"}")
	fmt.Println("str1:", str1)
	str2, _ := aes.AesDecrypt(str1)
	fmt.Println("str2", str2)

	aes1, _ := NewAes(biz)
	str3, _ := aes1.AesDecrypt(str1)
	fmt.Println("str3", str3)
}
