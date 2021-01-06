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

package remote

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/webankfintech/dockin-opserver/internal/log"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

func NewIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	return &IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

func NewTestIOStreamsDiscard() IOStreams {
	in := &bytes.Buffer{}
	return IOStreams{
		In:     in,
		Out:    ioutil.Discard,
		ErrOut: ioutil.Discard,
	}
}

type InteractStream struct {
	outBuffer   chan string
	ctx         context.Context
	inputBuffer chan []byte
}

func (is *InteractStream) Read(p []byte) (n int, err error) {
	select {
	case <-is.ctx.Done():
		n = -1
		err = fmt.Errorf("interactStream close as run finished")
		return
	case data, ok := <-is.inputBuffer:
		if !ok {
			log.Logger.Infof("InteractStream read chan close")
			return
		}
		n = len(data)
		copy(p, data)
	}
	return
}

func (is *InteractStream) Write(p []byte) (n int, err error) {
	n = len(p)
	log.Logger.Debugf("write data to client, in string=%s", string(p))
	is.outBuffer <- string(p)
	return
}

func (is *InteractStream) Close() error {
	close(is.outBuffer)
	close(is.inputBuffer)
	return nil
}
