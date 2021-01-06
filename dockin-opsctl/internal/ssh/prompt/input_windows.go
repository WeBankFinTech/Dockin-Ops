
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
	"bytes"
	"errors"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/mattn/go-tty"
)

const maxReadBytes = 1024

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var procGetNumberOfConsoleInputEvents = kernel32.NewProc("GetNumberOfConsoleInputEvents")

type WindowsParser struct {
	tty *tty.TTY
}

func (p *WindowsParser) Setup() error {
	t, err := tty.Open()
	if err != nil {
		return err
	}
	p.tty = t
	return nil
}

func (p *WindowsParser) TearDown() error {
	return p.tty.Close()
}

func (p *WindowsParser) GetKey(b []byte) Key {
	for _, k := range asciiSequences {
		if bytes.Compare(k.ASCIICode, b) == 0 {
			return k.Key
		}
	}
	return NotDefined
}

func (p *WindowsParser) Read() ([]byte, error) {
	var ev uint32
	r0, _, err := procGetNumberOfConsoleInputEvents.Call(p.tty.Input().Fd(), uintptr(unsafe.Pointer(&ev)))
	if r0 == 0 {
		return nil, err
	}
	if ev == 0 {
		return nil, errors.New("EAGAIN")
	}

	r, err := p.tty.ReadRune()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, maxReadBytes)
	n := utf8.EncodeRune(buf[:], r)
	for p.tty.Buffered() && n < maxReadBytes {
		r, err := p.tty.ReadRune()
		if err != nil {
			break
		}
		n += utf8.EncodeRune(buf[n:], r)
	}
	return buf[:n], nil
}

func (p *WindowsParser) GetWinSize() *WinSize {
	w, h, err := p.tty.Size()
	if err != nil {
		panic(err)
	}
	return &WinSize{
		Row: uint16(h),
		Col: uint16(w),
	}
}

func NewStandardInputParser() *WindowsParser {
	return &WindowsParser{}
}
