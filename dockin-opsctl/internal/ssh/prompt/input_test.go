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

package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_walkRemoteOutput(t *testing.T) {
	var buf = NewBuffer()

	buf.NewLine(false)
	t.Run("left padding 1", func(t *testing.T) {
		buf.InsertText("123456", false, true)
		data := []byte{0x8}
		t.Logf("text before:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "123456", buf.Text())
		assert.Equal(t, 6, buf.Document().CursorPositionCol())
		WalkRemoteOutput(data, buf)
		t.Logf("text after:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "123456", buf.Text())
		assert.Equal(t, 5, buf.Document().CursorPositionCol())
	})

	t.Run("delete from the end", func(t *testing.T) {
		buf.InsertText("123456", false, true)
		data := []byte{8, 27, 91, 75}
		t.Logf("text before:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "123456", buf.Text())
		assert.Equal(t, 6, buf.Document().CursorPositionCol())
		WalkRemoteOutput(data, buf)
		t.Logf("text after:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "12345", buf.Text())
		assert.Equal(t, 5, buf.Document().CursorPositionCol())
	})

	t.Run("delete from the middle", func(t *testing.T) {
		buf.InsertText("123456", false, true)
		// 1. move left
		data1 := []byte{8}
		t.Logf("text before move:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "123456", buf.Text())
		assert.Equal(t, 6, buf.Document().CursorPositionCol())
		WalkRemoteOutput(data1, buf)
		t.Logf("text after move:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "123456", buf.Text())
		assert.Equal(t, 5, buf.Document().CursorPositionCol())

		// 2. delete one
		data := []byte{8, 27, 91, 49, 80, 54, 8}
		t.Logf("text delete before:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		WalkRemoteOutput(data, buf)
		t.Logf("text delete after:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "12346", buf.Text())
		assert.Equal(t, 4, buf.Document().CursorPositionCol())
	})

	t.Run("up to exit", func(t *testing.T) {
		buf.InsertText("12346", false, true)
		// 1. move left
		t.Logf("text before exit:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "12346", buf.Text())
		assert.Equal(t, 5, buf.Document().CursorPositionCol())

		// 3. remote to exit
		data2 := []byte{8, 8, 8, 8, 8, 27, 91, 49, 80, 101, 120, 105, 116}
		WalkRemoteOutput(data2, buf)
		t.Logf("text up after:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "exit", buf.Text())
		assert.Equal(t, 4, buf.Document().CursorPositionCol())
	})

	t.Run("down to 12345", func(t *testing.T) {
		buf.InsertText("exit", false, true)
		// 1. move left
		t.Logf("text before exit:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "exit", buf.Text())
		assert.Equal(t, 4, buf.Document().CursorPositionCol())

		// 3. remote to exit
		data2 := []byte{8, 8, 8, 8, 49, 50, 51, 52,53}
		WalkRemoteOutput(data2, buf)
		t.Logf("text up after:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
		assert.Equal(t, "12345", buf.Text())
		assert.Equal(t, 5, buf.Document().CursorPositionCol())
	})

	t.Run("sample", func(t *testing.T) {
		buf.InsertText("ls", false, true)
		// 1. move left
		t.Logf("text before:%s, cusor=%d", buf.Text(), buf.Document().CursorPositionCol())
	})
}
