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
	"github.com/webankfintech/dockin-opserver/internal/log"
)

// WinSize represents the width and height of terminal.
type WinSize struct {
	Row uint16
	Col uint16
}

// ConsoleParser is an interface to abstract input layer.
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
	{Key: ControlJ, ASCIICode: []byte{0xa}},
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
	CtrlC      = []rune{0x3}
	// delete chinese word
	DeleteChinese = []rune{0x8, 0x8, 0x1b, 0x5b, 0x4b}

	x = []rune{0x1b, 0x5b, 0x31, 0x40}
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

// remote return left move a character
// just under data is 0x8, without other character
func IsLeftPadding(data []rune) bool {
	if len(data) != 1 {
		return false
	}

	return Compare(data, LeftPaddingSeq)
}

// the chinese world delete
func IsChineseWorkDel(data []rune) bool {
	if len(data) != 5 {
		return false
	}

	return Compare(data, DeleteChinese)
}

func IsControlC(data []rune) bool {
	if len(data) != 1 {
		return false
	}

	return Compare(data, CtrlC)
}

// remote return reach the border, in the end or the front
// just under data is 0x7, without other character
// just ignore it,
func IsBorder(data []rune) bool {
	if len(data) != 1 {
		return false
	}

	return Compare(data, BorderSeq)
}

// remote return right move a character
// just under data is 0x1b, 0x5b, 0x43, without other character
func IsRightPadding(data []rune) bool {
	if len(data) != 3 {
		return false
	}

	return Compare(data, RightPaddingSeq)
}

// remote cursor left move a character
// just under data is 0x1b, 0x5b, 0x43, without other character
func IsLeftCursor(data []rune) bool {
	if len(data) != 3 {
		return false
	}

	return Compare(data, LeftCursorSeq)
}

func WalkRemoteOutput(data []byte, buf *Buffer) {
	runeData := []rune(string(data))
	if IsBorder(runeData) {
		log.Logger.Infof("Border")
		return
	}
	if IsLeftCursor(runeData) {
		log.Logger.Infof("LeftCursor, on term=linux, remote the after")
		buf.Delete(1)
		return
	}
	if IsChineseWorkDel(runeData) {
		log.Logger.Infof("chinese work del, remote the after")
		buf.DeleteBeforeCursor(1)
		return
	}
	if IsRightPadding(runeData) {
		buf.CursorRight(1)
		log.Logger.Infof("RightPadding")
		return
	}
	if IsLeftPadding(runeData) {
		buf.CursorLeft(1)
		log.Logger.Infof("LeftPadding")
		return
	}
	if IsControlC(runeData) {
		log.Logger.Infof("IsControlC")
		buf = NewBuffer()
		return
	}

	del := false
	esc := false
	escape := false
	for _, d := range runeData {
		if d == 0x8 {
			buf.DeleteBeforeCursor(1)
			log.Logger.Infof("remove one character")
			continue
		}
		// 27 91 49 80
		// 27 91 50 80
		// 27 91 49 64(ignore)
		// []rune{0x1b, 0x5b, 0x31, 0x40}
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
				log.Logger.Infof("cursor move, ignore it")
			} else {
				del = true
			}
		} else if del {
			log.Logger.Debugf("80 occurred, ignore it")
			if d == 0x50 || d == 0x40 {
				del = false
			}
		} else {
			if IsBorder([]rune{d}) {
				// for instance: 0x7 0x7 0x7 l s
				continue
			}
			buf.InsertText(string(d), false, true)
			log.Logger.Infof("insert text:%s", string(d))
		}
	}
}
