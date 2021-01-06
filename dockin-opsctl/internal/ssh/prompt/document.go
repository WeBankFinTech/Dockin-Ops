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
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

type Document struct {
	Text string
	// This represents a index in a rune array of Document.Text.
	// So if Document is "日本(cursor)語", cursorPosition is 2.
	// But DisplayedCursorPosition returns 4 because '日' and '本' are double width characters.
	cursorPosition int
}

func NewDocument() *Document {
	return &Document{
		Text:           "",
		cursorPosition: 0,
	}
}

func (d *Document) DisplayCursorPosition() int {
	var position int
	runes := []rune(d.Text)[:d.cursorPosition]
	for i := range runes {
		position += runewidth.RuneWidth(runes[i])
	}
	return position
}

func (d *Document) GetCharRelativeToCursor(offset int) (r rune) {
	s := d.Text
	cnt := 0

	for len(s) > 0 {
		cnt++
		r, size := utf8.DecodeRuneInString(s)
		if cnt == d.cursorPosition+offset {
			return r
		}
		s = s[size:]
	}
	return 0
}

func (d *Document) TextBeforeCursor() string {
	r := []rune(d.Text)
	return string(r[:d.cursorPosition])
}

func (d *Document) TextAfterCursor() string {
	r := []rune(d.Text)
	return string(r[d.cursorPosition:])
}

func (d *Document) GetWordBeforeCursor() string {
	x := d.TextBeforeCursor()
	return x[d.FindStartOfPreviousWord():]
}

func (d *Document) GetWordAfterCursor() string {
	x := d.TextAfterCursor()
	return x[:d.FindEndOfCurrentWord()]
}

func (d *Document) GetWordBeforeCursorWithSpace() string {
	x := d.TextBeforeCursor()
	return x[d.FindStartOfPreviousWordWithSpace():]
}

func (d *Document) GetWordAfterCursorWithSpace() string {
	x := d.TextAfterCursor()
	return x[:d.FindEndOfCurrentWordWithSpace()]
}

func (d *Document) GetWordBeforeCursorUntilSeparator(sep string) string {
	x := d.TextBeforeCursor()
	return x[d.FindStartOfPreviousWordUntilSeparator(sep):]
}

func (d *Document) GetWordAfterCursorUntilSeparator(sep string) string {
	x := d.TextAfterCursor()
	return x[:d.FindEndOfCurrentWordUntilSeparator(sep)]
}

func (d *Document) GetWordBeforeCursorUntilSeparatorIgnoreNextToCursor(sep string) string {
	x := d.TextBeforeCursor()
	return x[d.FindStartOfPreviousWordUntilSeparatorIgnoreNextToCursor(sep):]
}

func (d *Document) GetWordAfterCursorUntilSeparatorIgnoreNextToCursor(sep string) string {
	x := d.TextAfterCursor()
	return x[:d.FindEndOfCurrentWordUntilSeparatorIgnoreNextToCursor(sep)]
}

func (d *Document) FindStartOfPreviousWord() int {
	x := d.TextBeforeCursor()
	i := strings.LastIndexByte(x, ' ')
	if i != -1 {
		return i + 1
	}
	return 0
}

func (d *Document) FindStartOfPreviousWordWithSpace() int {
	x := d.TextBeforeCursor()
	end := lastIndexByteNot(x, ' ')
	if end == -1 {
		return 0
	}

	start := strings.LastIndexByte(x[:end], ' ')
	if start == -1 {
		return 0
	}
	return start + 1
}

func (d *Document) FindStartOfPreviousWordUntilSeparator(sep string) int {
	if sep == "" {
		return d.FindStartOfPreviousWord()
	}

	x := d.TextBeforeCursor()
	i := strings.LastIndexAny(x, sep)
	if i != -1 {
		return i + 1
	}
	return 0
}

func (d *Document) FindStartOfPreviousWordUntilSeparatorIgnoreNextToCursor(sep string) int {
	if sep == "" {
		return d.FindStartOfPreviousWordWithSpace()
	}

	x := d.TextBeforeCursor()
	end := lastIndexAnyNot(x, sep)
	if end == -1 {
		return 0
	}
	start := strings.LastIndexAny(x[:end], sep)
	if start == -1 {
		return 0
	}
	return start + 1
}

func (d *Document) FindEndOfCurrentWord() int {
	x := d.TextAfterCursor()
	i := strings.IndexByte(x, ' ')
	if i != -1 {
		return i
	}
	return len(x)
}

func (d *Document) FindEndOfCurrentWordWithSpace() int {
	x := d.TextAfterCursor()

	start := indexByteNot(x, ' ')
	if start == -1 {
		return len(x)
	}

	end := strings.IndexByte(x[start:], ' ')
	if end == -1 {
		return len(x)
	}

	return start + end
}

func (d *Document) FindEndOfCurrentWordUntilSeparator(sep string) int {
	if sep == "" {
		return d.FindEndOfCurrentWord()
	}

	x := d.TextAfterCursor()
	i := strings.IndexAny(x, sep)
	if i != -1 {
		return i
	}
	return len(x)
}

func (d *Document) FindEndOfCurrentWordUntilSeparatorIgnoreNextToCursor(sep string) int {
	if sep == "" {
		return d.FindEndOfCurrentWordWithSpace()
	}

	x := d.TextAfterCursor()

	start := indexAnyNot(x, sep)
	if start == -1 {
		return len(x)
	}

	end := strings.IndexAny(x[start:], sep)
	if end == -1 {
		return len(x)
	}

	return start + end
}

func (d *Document) CurrentLineBeforeCursor() string {
	s := strings.Split(d.TextBeforeCursor(), "\n")
	return s[len(s)-1]
}

func (d *Document) CurrentLineAfterCursor() string {
	return strings.Split(d.TextAfterCursor(), "\n")[0]
}

func (d *Document) CurrentLine() string {
	return d.CurrentLineBeforeCursor() + d.CurrentLineAfterCursor()
}

func (d *Document) lineStartIndexes() []int {
	// TODO: Cache, because this is often reused.
	// (If it is used, it's often used many times.
	// And this has to be fast for editing big documents!)
	lc := d.LineCount()
	lengths := make([]int, lc)
	for i, l := range d.Lines() {
		lengths[i] = len(l)
	}

	// Calculate cumulative sums.
	indexes := make([]int, lc+1)
	indexes[0] = 0 // https://github.com/jonathanslenders/python-prompt-toolkit/blob/master/prompt_toolkit/document.py#L189
	pos := 0
	for i, l := range lengths {
		pos += l + 1
		indexes[i+1] = pos
	}
	if lc > 1 {
		// Pop the last item. (This is not a new line.)
		indexes = indexes[:lc]
	}
	return indexes
}

func (d *Document) findLineStartIndex(index int) (pos int, lineStartIndex int) {
	indexes := d.lineStartIndexes()
	pos = bisectRight(indexes, index) - 1
	lineStartIndex = indexes[pos]
	return
}

func (d *Document) CursorPositionRow() (row int) {
	row, _ = d.findLineStartIndex(d.cursorPosition)
	return
}

func (d *Document) CursorPositionCol() (col int) {
	// Don't use self.text_before_cursor to calculate this. Creating substrings
	// and splitting is too expensive for getting the cursor position.
	_, index := d.findLineStartIndex(d.cursorPosition)
	col = d.cursorPosition - index
	return
}

func (d *Document) GetCursorLeftPosition(count int) int {
	if count < 0 {
		return d.GetCursorRightPosition(-count)
	}
	if d.CursorPositionCol() > count {
		return -count
	}
	return -d.CursorPositionCol()
}

func (d *Document) GetCursorRightPosition(count int) int {
	if count < 0 {
		return d.GetCursorLeftPosition(-count)
	}
	if len(d.CurrentLineAfterCursor()) > count {
		return count
	}
	return len(d.CurrentLineAfterCursor())
}

func (d *Document) GetCursorUpPosition(count int, preferredColumn int) int {
	var col int
	if preferredColumn == -1 { // -1 means nil
		col = d.CursorPositionCol()
	} else {
		col = preferredColumn
	}

	row := d.CursorPositionRow() - count
	if row < 0 {
		row = 0
	}
	return d.TranslateRowColToIndex(row, col) - d.cursorPosition
}

func (d *Document) GetCursorDownPosition(count int, preferredColumn int) int {
	var col int
	if preferredColumn == -1 { // -1 means nil
		col = d.CursorPositionCol()
	} else {
		col = preferredColumn
	}
	row := d.CursorPositionRow() + count
	return d.TranslateRowColToIndex(row, col) - d.cursorPosition
}

func (d *Document) Lines() []string {
	// TODO: Cache, because this one is reused very often.
	return strings.Split(d.Text, "\n")
}

func (d *Document) LineCount() int {
	return len(d.Lines())
}

func (d *Document) TranslateIndexToPosition(index int) (row int, col int) {
	row, rowIndex := d.findLineStartIndex(index)
	col = index - rowIndex
	return
}

func (d *Document) TranslateRowColToIndex(row int, column int) (index int) {
	indexes := d.lineStartIndexes()
	if row < 0 {
		row = 0
	} else if row > len(indexes) {
		row = len(indexes) - 1
	}
	index = indexes[row]
	line := d.Lines()[row]

	// python) result += max(0, min(col, len(line)))
	if column > 0 || len(line) > 0 {
		if column > len(line) {
			index += len(line)
		} else {
			index += column
		}
	}

	// Keep in range. (len(self.text) is included, because the cursor can be
	// right after the end of the text as well.)
	// python) result = max(0, min(result, len(self.text)))
	if index > len(d.Text) {
		index = len(d.Text)
	}
	if index < 0 {
		index = 0
	}
	return index
}

func (d *Document) OnLastLine() bool {
	return d.CursorPositionRow() == (d.LineCount() - 1)
}

func (d *Document) GetEndOfLinePosition() int {
	return len([]rune(d.CurrentLineAfterCursor()))
}

func (d *Document) leadingWhitespaceInCurrentLine() (margin string) {
	trimmed := strings.TrimSpace(d.CurrentLine())
	margin = d.CurrentLine()[:len(d.CurrentLine())-len(trimmed)]
	return
}

func bisectRight(a []int, v int) int {
	return bisectRightRange(a, v, 0, len(a))
}

func bisectRightRange(a []int, v int, lo, hi int) int {
	s := a[lo:hi]
	return sort.Search(len(s), func(i int) bool {
		return s[i] > v
	})
}

func indexByteNot(s string, c byte) int {
	n := len(s)
	for i := 0; i < n; i++ {
		if s[i] != c {
			return i
		}
	}
	return -1
}

func lastIndexByteNot(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != c {
			return i
		}
	}
	return -1
}

type asciiSet [8]uint32

func (as *asciiSet) notContains(c byte) bool {
	return (as[c>>5] & (1 << uint(c&31))) == 0
}

func makeASCIISet(chars string) (as asciiSet, ok bool) {
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c >= utf8.RuneSelf {
			return as, false
		}
		as[c>>5] |= 1 << uint(c&31)
	}
	return as, true
}

func indexAnyNot(s, chars string) int {
	if len(chars) > 0 {
		if len(s) > 8 {
			if as, isASCII := makeASCIISet(chars); isASCII {
				for i := 0; i < len(s); i++ {
					if as.notContains(s[i]) {
						return i
					}
				}
				return -1
			}
		}
		for i := 0; i < len(s); {
			// I don't know why strings.IndexAny doesn't add rune count here.
			r, size := utf8.DecodeRuneInString(s[i:])
			i += size
			for _, c := range chars {
				if r != c {
					return i
				}
			}
		}
	}
	return -1
}

func lastIndexAnyNot(s, chars string) int {
	if len(chars) > 0 {
		if len(s) > 8 {
			if as, isASCII := makeASCIISet(chars); isASCII {
				for i := len(s) - 1; i >= 0; i-- {
					if as.notContains(s[i]) {
						return i
					}
				}
				return -1
			}
		}
		for i := len(s); i > 0; {
			r, size := utf8.DecodeLastRuneInString(s[:i])
			i -= size
			for _, c := range chars {
				if r != c {
					return i
				}
			}
		}
	}
	return -1
}
