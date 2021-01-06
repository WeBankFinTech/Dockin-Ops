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

package option

import (
	"github.com/webankfintech/dockin-opsctl/internal/common/protocol"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ExecOption struct {
	Command string
	Type    string
	Name    string

	CommandList []string

	ContainerName string
	Rule          string
	Namespace     string

	Stdin bool
	TTY   bool
	User          string
	Env           []string
	WorkDir       string
}

func (option *ExecOption) Complete(configFlags *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.Errorf("%s\n%s", NoTypeOrNameErr, ExecCommandSuggest)
	} else {
		option.Name = args[0]
	}

	log.Debugf("cmdLine params:%s", args)
	option.CommandList = args[1:]
	option.Namespace, _ = cmd.Flags().GetString("namespace")
	option.Rule, _ = cmd.Flags().GetString("rule")

	return nil
}

func (option *ExecOption) Validate() error {
	if len(option.CommandList) == 0 {
		return errors.Errorf("%s\n%s", NoCommandErr, ExecCommandSuggest)
	}
	return nil
}

func (option *ExecOption) Run() error {
	log.Debugf("exec option = %#v", option)
	proto := protocol.NewProto()
	proto.Command = option.Command
	proto.Resource = option.Type
	proto.Name = option.Name
	proto.Flags = option.CommandList
	proto.Env =	option.Env
	proto.WorkDir = option.WorkDir

	if option.Namespace != "" {
		proto.Params["namespace"] = option.Namespace
	}

	if option.Rule != "" {
		proto.Params["rule"] = option.Rule
	}
	if option.TTY{
		proto.Params["tty"] = true
	}

	return  ssh.RunInteractive(proto)
}
