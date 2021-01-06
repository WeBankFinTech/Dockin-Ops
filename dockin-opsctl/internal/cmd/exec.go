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

package cmd

import (
	"github.com/webankfintech/dockin-opsctl/internal/common"

	"github.com/webankfintech/dockin-opsctl/internal/option"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	execExample = `
		# Get output from running 'date' command from pod mypod, using the first container by default
		dockin-opsctl exec mypod date

		# Get output from running 'date' command in ruby-container from pod mypod
		dockin-opsctl exec mypod -c ruby-container date

		# Switch to raw terminal mode, sends stdin to 'bash' in ruby-container from pod mypod
		# and sends stdout/stderr from 'bash' back to the client
		dockin-opsctl exec mypod -c ruby-container -i -t -- bash -il

		# List contents of /usr from the first container of pod mypod and sort by modification time.
		# If the command you want to execute in the pod has any flags in common (e.g. -i),
		# you must use two dashes (--) to separate your command's flags/arguments.
		# Also note, do not surround your command and its flags/arguments with quotes
		# unless that is how you would execute it normally (i.e., do ls -t /usr, not "ls -t /usr").
		dockin-opsctl exec mypod -i -t -- ls -t /usr
		`
)

func NewExecCmd(configFlags *genericclioptions.ConfigFlags) *cobra.Command {
	opt := &option.ExecOption{
		Command: "exec",
		Stdin:   true,
	}
	execCmd := &cobra.Command{
		Use:                   "exec pod command args",
		DisableFlagsInUseLine: true,
		Short:                 "exec cmd in pod",
		Long:                  "exec cmd in pod",
		Example:               execExample,
		Run: func(cmd *cobra.Command, args []string) {
			//argsLenAtDash := cmd.ArgsLenAtDash()
			utils.CheckErr(opt.Complete(configFlags, cmd, args))
			utils.CheckErr(opt.Validate())
			utils.CheckErr(opt.Run())
		},
	}
	execCmd.Flags().Duration("pod-running-timeout", common.DefaultTimeout, "The length of time (like 5s, 2m, or 3h, higher than zero) to wait until at least one pod is running")
	execCmd.Flags().StringVarP(&opt.ContainerName, "container", "c", "op_server", "ContainerName name. If omitted, the first container in the pod will be chosen")
	execCmd.Flags().BoolVarP(&opt.Stdin, "stdin", "i", opt.Stdin, "Pass stdin to the container")
	execCmd.Flags().BoolVarP(&opt.TTY, "tty", "t", opt.TTY, "Stdin is a TTY")
	execCmd.Flags().StringVarP(&opt.User, "user", "s", opt.User, "run as user name, default to app")
	execCmd.Flags().StringArrayVarP(&opt.Env, "env", "e", opt.Env, "exec env variable, like: a=1")
	execCmd.Flags().StringVarP(&opt.WorkDir, "work-dir", "w", opt.WorkDir, "exec work directory")
	return execCmd
}
