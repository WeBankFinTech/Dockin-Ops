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

package utils

import (
	"bytes"
	"context"
	"github.com/webankfintech/dockin-opagent/internal/log"
	"github.com/pkg/errors"
	"os/exec"
	"time"
)

func ExecCmd(name string, command string, timeOut int64) (string, error) {
	var (
		stdout  bytes.Buffer
		stderr  bytes.Buffer
		Timeout = time.Duration(timeOut) * time.Millisecond
	)
	ctxt, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctxt, name, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Logger.Warnf("start cmd:%s err:%s", command, err.Error())
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		log.Logger.Warnf("cmd:{%s} %s,%s", command, err.Error(), stderr.String())
		return "", errors.Errorf("%s,%s", err.Error(), stderr.String())
	}

	return stdout.String(), nil
}
