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

package batch

import (
	"sync"
	"testing"
	"time"
)

type foo struct {
	Name   string
	Status int
}

func Test_Batch(t *testing.T) {
	b := BatchTask{
		Wg:      new(sync.WaitGroup),
		Timeout: time.Duration(time.Second * 3),
	}

	foolist := []*foo{&foo{"first", 1}, &foo{"second", 1}, &foo{"third", 3}}

	var newfoolist []*foo
	setfooname := func(f *foo) {
		f.Name = "zero"
		newfoolist = append(newfoolist, f)
	}

	for _, x := range foolist {
		index := *x
		b.Submit(func() {
			setfooname(&index)
		})
	}

	b.WaitTimeout()
	for _, fo := range newfoolist {
		t.Logf("%#v", fo)
	}
}
