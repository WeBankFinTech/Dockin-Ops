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

// Key is the type express the key inserted from user.
type Key int

// ASCIICode is the type contains Key and it's ascii byte array.
type ASCIICode struct {
	Key       Key
	ASCIICode []byte
}

const (
	Escape Key = iota
	ControlC
	ControlR
	Right
	Left
	LeftCursor
	Home
	End
	Delete
	Backspace
	Tab
	Enter
	ControlJ
	NotDefined
)
