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
	"github.com/webankfintech/dockin-opsctl/internal/option"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	Long = `
		Display one or many resources

		Prints a table of the most important information about the specified resources.
		You can filter the list using a label selector and the --selector flag. If the
		desired resource type is namespaced you will only see results in your current
		namespace unless you pass --all-namespaces.

		Uninitialized objects are not shown unless --include-uninitialized is passed.

		By specifying the output as 'template' and providing a Go template as the value
		of the --template flag, you can filter the attributes of the fetched resources.`
	Example = `
		# List all pods in ps output format.
		dockin-opsctl get pods

		# List all pods in ps output format with more information (such as node name).
		dockin-opsctl get pods -o wide

		# List a single replication controller with specified NAME in ps output format.
		dockin-opsctl get replicationcontroller web

		# List deployments in JSON output format, in the "v1" version of the "apps" API group:
		dockin-opsctl get deployments.v1.apps -o json

		# List a single pod in JSON output format.
		dockin-opsctl get -o json pod web-pod-13je7

		# List a pod identified by type and name specified in "pod.yaml" in JSON output format.
		dockin-opsctl get -f pod.yaml -o json

		# List resources from a directory with kustomization.yaml - e.g. dir/kustomization.yaml.
		dockin-opsctl get -k dir/

		# Return only the phase value of the specified pod.
		dockin-opsctl get -o template pod/web-pod-13je7 --template={{.status.phase}}

		# List resource information in custom columns.
		dockin-opsctl get pod test-pod -o custom-columns=CONTAINER:.spec.containers[0].name,IMAGE:.spec.containers[0].image

		# List all replication controllers and services together in ps output format.
		dockin-opsctl get rc,services

		# List one or more resources by their type and names.
		dockin-opsctl get rc/web service/frontend pods/web-pod-13je7`
)

func NewGetCmd(configFlags *genericclioptions.ConfigFlags) *cobra.Command {
	opt := &option.GetOption{
		Command:       "get",
		AllNamespaces: false,
	}
	cmd := &cobra.Command{
		Use:                   "get resources name -o wide/yaml",
		DisableFlagsInUseLine: true,
		Short:                 "Display one or many resources",
		Long:                  Long,
		Example:               Example,
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(opt.Complete(configFlags, cmd, args))
			utils.CheckErr(opt.Run())
		},
	}

	cmd.Flags().StringVarP(&opt.OutputFormat, "output", "o", opt.OutputFormat, "Output format. Must be one of yaml|json")
	cmd.Flags().BoolVarP(&opt.AllNamespaces, "all-namespaces", "A", opt.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	return cmd
}
