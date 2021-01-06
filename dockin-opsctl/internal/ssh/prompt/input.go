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
	"github.com/webankfintech/dockin-opsctl/internal/log"
)

type WinSize struct {
	Row uint16
	Col uint16
}

type ConsoleParser interface {
	// Setup should be called before starting input
	Setup() error
	// TearDown should be called after stopping input
	TearDown() error
	// GetKey returns Key correspond to input byte codes.
	GetKey(b []byte) Key
	// GetWinSize returns WinSize object to represent width and height of terminal.
	GetWinSize() *WinSize
	// Read returns byte array.
	Read() ([]byte, error)
}

var asciiSequences = []*ASCIICode{
	{Key: Escape, ASCIICode: []byte{0x1b}},
	{Key: Enter, ASCIICode: []byte{0xd}},
	{Key: Tab, ASCIICode: []byte{0x9}},
	{Key: ControlR, ASCIICode: []byte{0x12}},
	{Key: ControlC, ASCIICode: []byte{0x3}},
}

var (
	// LeftPadding
	LeftPaddingSeq = []rune{0x8}
	// RightPadding
	RightPaddingSeq = []rune{0x1b, 0x5b, 0x43}
	// Border
	BorderSeq = []rune{0x7}
	// LeftCursor
	LeftCursorSeq = []rune{0x1b, 0x5b, 0x4b}
	// Backspace/Delete
	DeleteSeq1 = []rune{0x8, 0x1b, 0x5b, 0x31, 0x50}
	DeleteSeq2 = []rune{0x8, 0x1b, 0x5b, 0x32, 0x50}
	// ControlC
	CtrlC = []rune{0x3}
)

func Compare(lhs, rhs []rune) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	for i, d := range lhs {
		if d != rhs[i] {
			return false
		}
	}
	return true
}

func IsLeftPadding(data []rune) bool {
	if len(data) != 1 {
		return false
	}

	return Compare(data, LeftPaddingSeq)
}

func IsControlC(data []rune) bool {
	if len(data) != 1 {
		return false
	}

	return Compare(data, CtrlC)
}

func IsBorder(data []rune) bool {
	if len(data) != 1 {
		return false
	}

	return Compare(data, BorderSeq)
}

func IsRightPadding(data []rune) bool {
	if len(data) != 3 {
		return false
	}

	return Compare(data, RightPaddingSeq)
}

func IsLeftCursor(data []rune) bool {
	if len(data) != 3 {
		return false
	}

	return Compare(data, LeftCursorSeq)
}

func WalkRemoteOutput(data []byte, buf *Buffer) {
	runeData := []rune(string(data))
	if IsBorder(runeData) {
		log.Debugf("Border")
		return
	}
	if IsLeftCursor(runeData) {
		log.Debugf("LeftCursor")
		return
	}
	if IsRightPadding(runeData) {
		buf.CursorRight(1)
		log.Debugf("RightPadding")
		return
	}
	if IsLeftPadding(runeData) {
		buf.CursorLeft(1)
		log.Debugf("LeftPadding")
		return
	}
	if IsControlC(runeData) {
		buf = NewBuffer()
		log.Debugf("Control C")
		return
	}

	del := false
	esc := false
	escape := false
	for _, d := range runeData {
		if d == 0x8 {
			buf.DeleteBeforeCursor(1)
			log.Debugf("remove one character")
			continue
		}
		// 27 91 49 80
		// 27 91 50 80
		if !esc && d == 0x1b {
			esc = true
			continue
		} else if esc && d == 0x5b {
			esc = false
			escape = true
			continue
		} else if escape {
			escape = false
			if d == 0x41 || d == 0x42 || d == 0x43 || d == 0x44 || d == 0x4b {
				// up/down/left/right
				log.Debugf("cursor move, ignore it")
			} else {
				del = true
			}
		} else if del {
			log.Debugf("80 occurred, ignore it")
			if d == 0x50 {
				del = false
			}
		} else {
			log.Debugf("insert :%s to buf", string(d))
			buf.InsertText(string(d), false, true)
		}
	}
}
