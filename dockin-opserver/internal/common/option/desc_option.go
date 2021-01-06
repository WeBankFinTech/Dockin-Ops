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
	"fmt"

	"github.com/webankfintech/dockin-opserver/internal/common"
	utils2 "github.com/webankfintech/dockin-opserver/internal/common/option/utils"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type DescOption struct {
	Name              string
	Type              string
	Selector          string
	Namespace         string
	AllNamespaces     bool
	DescriberSettings *common.DescriberSettings
}

func (o *DescOption) Complete(flags *genericclioptions.ConfigFlags, command *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no resource type or name provided")
	} else if len(args) == 1 {
		o.Type = args[0]
	} else {
		o.Type = args[0]
		o.Name = args[1]
	}
	o.Namespace = *flags.Namespace
	//if o.Namespace == "" {
	//	return fmt.Errorf("no namespace provided")
	//}
	return nil
}

func (option *DescOption) getUrl() string {
	return fmt.Sprintf("http://%s?command=%s", utils2.CommonUrl, "describe")
}

func (option *DescOption) Run() error {
	body, err := jsoniter.MarshalToString(option)
	if err != nil {
		fmt.Printf("json encode failed %v", err)
	}
	resp, err := utils2.HttpPost(option.getUrl(), body)
	if err != nil {
		return err
	}
	fmt.Println(string(resp))
	return nil
}
