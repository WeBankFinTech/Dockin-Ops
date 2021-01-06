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

package exec

import (
	"fmt"
	"testing"

	"github.com/webankfintech/dockin-opserver/internal/model"
	"github.com/webankfintech/dockin-opserver/internal/remote"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand_RunNoTty(t *testing.T) {
	t.Run("bach -c", func(t *testing.T) {
		exec := &ExecCommand{
			OpsOpts: &model.OpsOption{
				Container: "dbdcss-csloypconf-20200423-150120359",
				Name:      "dbdcss-csloypconf-20200423-150120359-0",
				Flags:     []string{"ls"},
			},
			//Conn:    conn,
			HostIp: "192.168.1.1",
		}
		ios, _, out, _ := remote.NewIOStreams()
		err := exec.RunNoTty("traceId", ios)
		assert.NoError(t, err)
		fmt.Println(out.String())
	})
}

func Test_CpFile(t *testing.T) {
	fileInPod := "/data/a.txt"
	fileInNode := "/logs"
	t.Run("bach -c", func(t *testing.T) {
		exec := &ExecCommand{
			OpsOpts: &model.OpsOption{
				Container: "dbdcss-csloypconf-20200423-150120359",
				Name:      "dbdcss-csloypconf-20200423-150120359-0",
				Flags:     []string{"cp", "-rf", fileInPod, fileInNode},
			},
			//Conn:    conn,
			HostIp: "192.168.1.1",
		}
		ios, _, out, _ := remote.NewIOStreams()
		err := exec.RunNoTty("traceId", ios)
		fmt.Println(ios.ErrOut)
		assert.NoError(t, err)
		fmt.Println(out.String())
	})
}

func Test_FileExist(t *testing.T) {
	t.Run("bach -c", func(t *testing.T) {
		exec := &ExecCommand{
			OpsOpts: &model.OpsOption{
				Container: "dbdcss-csloypconf-20200423-150120359",
				Name:      "dbdcss-csloypconf-20200423-150120359-0",
				Flags:     []string{"/bin/bash", "-c", "ls"},
				//Flags: []string{"du","-shb","/tmp/devops"},
			},
			//Conn:    conn,
			HostIp: "192.168.1.1",
		}
		ios, _, out, _ := remote.NewIOStreams()
		err := exec.RunNoTty("traceId", ios)
		fmt.Println(ios.ErrOut)
		assert.NoError(t, err)
		fmt.Println(out.String())
	})
}
