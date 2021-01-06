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

func GoLineEnd(buf *Buffer) {
	x := []rune(buf.Document().TextAfterCursor())
	buf.CursorRight(len(x))
}

func GoLineBeginning(buf *Buffer) {
	x := []rune(buf.Document().TextBeforeCursor())
	buf.CursorLeft(len(x))
}

func DeleteChar(buf *Buffer) {
	buf.Delete(1)
}

func DeleteWord(buf *Buffer) {
	buf.DeleteBeforeCursor(len([]rune(buf.Document().TextBeforeCursor())) - buf.Document().FindStartOfPreviousWordWithSpace())
}

func DeleteBeforeChar(buf *Buffer) {
	buf.DeleteBeforeCursor(1)
}

func GoRightChar(buf *Buffer) {
	buf.CursorRight(1)
}

func GoLeftChar(buf *Buffer) {
	buf.CursorLeft(1)
}

func GoRightWord(buf *Buffer) {
	buf.CursorRight(buf.Document().FindEndOfCurrentWordWithSpace())
}

func GoLeftWord(buf *Buffer) {
	buf.CursorLeft(len([]rune(buf.Document().TextBeforeCursor())) - buf.Document().FindStartOfPreviousWordWithSpace())
}

func Empty(buf *Buffer) {

}