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
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/webankfintech/dockin-opsctl/internal/common"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	"github.com/webankfintech/dockin-opsctl/internal/utils/aes"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	authLong    = `auth return a access token according to the user name and password`
	authExample = `dockin-opsctl auth -u admin -p admin -r default`
)

func NewAuthCmd(configFlags *genericclioptions.ConfigFlags) *cobra.Command {
	opt := &AuthOption{}
	sshCmd := &cobra.Command{
		Use:                   "auth user name and password",
		DisableFlagsInUseLine: true,
		Short:                 "auth",
		Long:                  authLong,
		Example:               authExample,
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(opt.Complete(configFlags, cmd, args))
			utils.CheckErr(opt.Validate())
			utils.CheckErr(opt.Run())
		},
	}
	sshCmd.Flags().StringVarP(&opt.UserName, "user name", "u", opt.UserName, "pass user name")
	sshCmd.Flags().StringVarP(&opt.Password, "password", "p", opt.Password, "pass password")
	return sshCmd
}

type AuthOption struct {
	UserName string
	Password string
	Rule     string
}

func (option *AuthOption) Validate() error {
	if option.UserName == "" {
		return errors.Errorf("user name is empty")
	}
	//if option.Password == "" {
	//	return errors.Errorf("password is empty")
	//}
	if option.Rule == "" {
		option.Rule = "default"
	}
	return nil
}

func (option *AuthOption) Complete(configFlags *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	option.Rule, _ = cmd.Flags().GetString("rule")
	return nil
}

func (option *AuthOption) Run() error {
	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		nerr := errors.Errorf("encrypt failed err=%s", err.Error())
		log.Debugf(nerr.Error())
		return nerr
	}

	passAes, err := aes.AesEncrypt(option.Password)
	if err != nil {
		nerr := errors.Errorf("decrypt err=%s", err.Error())
		log.Debugf(nerr.Error())
		return nerr
	}
	authUrl := fmt.Sprintf("%s?userName=%s&password=%s&rule=%s", common.GetCommonUrlByCmd("ctrl/auth"), option.UserName, passAes, option.Rule)
	body, err := utils.HttpGetWithTimeout(authUrl, time.Second*3)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}
