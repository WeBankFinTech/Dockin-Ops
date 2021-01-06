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

package ssh

import (
	"os"

	"github.com/webankfintech/dockin-opsctl/internal/log"
)

var (
	Stdin  = os.Stdin
	Stdout = os.Stdout
	Stderr = os.Stderr
	appSuffix = "]$ "
	rootSuffix = "]# "
)

type ReadWriter struct {
}

func (rw ReadWriter)Write(p []byte) (n int, err error) {
	log.Debugf("receive output:%v, byte=%v", string(p), p)
	n = len(p)
	Stdout.Write(p)
	return
}