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
	"time"

	"github.com/pkg/errors"
)

type Future func()

type BatchTask struct {
	task    []Future
	Wg      *sync.WaitGroup
	Timeout time.Duration
}

func (b *BatchTask) Submit(fn Future) {
	b.task = append(b.task, fn)
}

func (b *BatchTask) WaitTimeout() error {
	b.Wg.Add(len(b.task))
	b.run()

	c := make(chan struct{})
	go func() {
		defer close(c)
		b.Wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case <-time.After(b.Timeout):
		return errors.Errorf("canceled as timeout task")
	}
}

func (b *BatchTask) run() {
	for _, fn := range b.task {
		go func() {
			defer b.Wg.Done()
			fn()
		}()
	}
}
