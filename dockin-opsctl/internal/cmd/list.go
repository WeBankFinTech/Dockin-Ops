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
	"github.com/spf13/cobra"
	"github.com/webankfintech/dockin-opsctl/internal/option"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	listLong    = templates.LongDesc(`get resource info from rm interface`)
	listExample = templates.Examples(`
		# get all container info from ip
		dockin-opsctl list ip "0.0.0.0" -d dcn
`)
)

func NewListCmd(configFlags *genericclioptions.ConfigFlags) *cobra.Command {
	o := &option.ListOption{
		Command: "list",
	}
	cmd := &cobra.Command{
		Use:                   "list resources [subsysName|subsysId|ip] -d [dcn]",
		DisableFlagsInUseLine: true,
		Short:                 "get resource info from rm interface",
		Long:                  listLong,
		Example:               listExample,
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(o.Complete(configFlags, cmd, args))
			utils.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.Condition, "condition", "d", o.Condition, "According dcn to screening")
	return cmd
}
