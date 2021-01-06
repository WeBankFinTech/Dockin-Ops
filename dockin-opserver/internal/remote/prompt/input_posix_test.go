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

// +build !windows

package prompt

import (
	"testing"
)

func TestPosixParserGetKey(t *testing.T) {
	pp := &PosixParser{}
	scenarioTable := []struct {
		input    []byte
		expected Key
	}{
		{
			input:    []byte{0x1b},
			expected: Escape,
		},
		{
			input:    []byte{'a'},
			expected: NotDefined,
		},
	}

	for _, s := range scenarioTable {
		key := pp.GetKey(s.input)
		if key != s.expected {
			t.Errorf("Should be %s, but got %s", key, s.expected)
		}
	}
}
