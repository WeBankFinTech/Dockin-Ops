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
	"github.com/webankfintech/dockin-opsctl/internal/utils/aes"
	"io/ioutil"

	"github.com/webankfintech/dockin-opsctl/internal/common"
	"github.com/webankfintech/dockin-opsctl/internal/common/protocol"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ListOption struct {
	Command string
	Type    string
	Name    string

	Rule      string
	Namespace string

	Condition string
}

func (o *ListOption) Complete(flags *genericclioptions.ConfigFlags, command *cobra.Command, args []string) error {
	if len(args) <= 1 {
		return errors.Errorf("%s\n%s", NoTypeOrNameErr, ListCommandSuggest)
	} else {
		o.Type = args[0]
		o.Name = args[1]
	}
	log.Debugf("cmdLine params:%s", args)
	o.Namespace, _ = command.Flags().GetString("namespace")
	o.Rule, _ = command.Flags().GetString("rule")

	return nil
}

func (option *ListOption) Run() error {
	proto := protocol.NewProto()
	proto.Command = option.Command
	proto.Resource = option.Type
	proto.Name = option.Name

	if option.Condition != "" {
		proto.Params["dcn"] = option.Condition
	}

	if option.Namespace != "" {
		proto.Params["namespace"] = option.Namespace
	}

	if option.Rule != "" {
		proto.Params["rule"] = option.Rule
	}

	data, err := jsoniter.MarshalToString(proto)
	if err != nil {
		log.Debugf("json marshal %s err:%s", data, err.Error())
		return err
	}

	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		nerr := errors.Errorf("encrypt failed err=%s", err.Error())
		log.Debugf(nerr.Error())
		return nerr
	}

	encode, err := aes.AesEncrypt(data)
	if err != nil {
		nerr := errors.Errorf("decrypt err=%s", err.Error())
		log.Debugf(nerr.Error())
		return nerr
	}

	params := make(map[string]string)
	params["params"] = encode

	resp, err := utils.HttpGet(common.GetCommonUrlByCmd("rm"), params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//提取响应头数据
	b, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(b))
	return nil
}
