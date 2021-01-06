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

package dockershim

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
)

type IOStreams struct {
	// In think, os.Stdin
	In io.Reader
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

func NewStdIOStreams() *IOStreams {
	return &IOStreams{
		In: os.Stdin,
		Out: os.Stdout,
		ErrOut: os.Stderr,
	}
}

func NewBufferIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	return &IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

func NewBufferIOStreamsDiscard() *IOStreams {
	in := &bytes.Buffer{}
	return &IOStreams{
		In:     in,
		Out:    ioutil.Discard,
		ErrOut: ioutil.Discard,
	}
}
