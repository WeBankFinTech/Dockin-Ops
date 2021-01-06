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
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ExplainOption struct {
	APIVersion string
	Recursive  bool
}

func (option *ExplainOption) Complete(flags *genericclioptions.ConfigFlags, command *cobra.Command) error {
	return nil
}

func (option *ExplainOption) Validate(strings []string) error {
	return nil
}

func (option *ExplainOption) Run(strings []string) error {
	return nil
}

func NewExplainOption() *ExplainOption {
	return &ExplainOption{}
}
