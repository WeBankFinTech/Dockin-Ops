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

// KeyBindFunc receives buffer and processed it.
type KeyBindFunc func(*Buffer)

// KeyBind represents which key should do what operation.
type KeyBind struct {
	Key Key
	Fn  KeyBindFunc
}

// ASCIICodeBind represents which []byte should do what operation
type ASCIICodeBind struct {
	ASCIICode []byte
	Fn        KeyBindFunc
}

// KeyBindMode to switch a key binding flexibly.
type KeyBindMode string

const (
	// CommonKeyBind is a mode without any keyboard shortcut
	CommonKeyBind KeyBindMode = "common"
	// EmacsKeyBind is a mode to use emacs-like keyboard shortcut
	EmacsKeyBind KeyBindMode = "emacs"
)

var CommonKeyBindings = []KeyBind{
	// Go to the End of the line
	{
		Key: End,
		Fn:  GoLineEnd,
	},
	// Go to the beginning of the line
	{
		Key: Home,
		Fn:  GoLineBeginning,
	},
	// Delete character under the cursor
	{
		Key: Delete,
		Fn:  DeleteChar,
	},
	// Backspace
	{
		Key: Backspace,
		Fn:  DeleteBeforeChar,
	},
	// Right allow: Forward one character
	{
		Key: Right,
		Fn:  GoRightChar,
	},
	// Left allow: Backward one character
	{
		Key: Left,
		Fn:  GoLeftChar,
	},
	// delete
	{
		Key: Delete,
		Fn:  DeleteBeforeChar,
	},
	// remote cursor display
	{
		Key: LeftCursor,
		Fn:  Empty,
	},
}
