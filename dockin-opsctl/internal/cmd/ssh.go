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
	"net/http"
	"os/user"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/webankfintech/dockin-opsctl/internal/common"
	"github.com/webankfintech/dockin-opsctl/internal/common/protocol"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/utils/aes"
	"github.com/webankfintech/dockin-opsctl/internal/utils/base"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/webankfintech/dockin-opsctl/internal/ssh"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/webankfintech/dockin-opsctl/internal/utils"
)

var (
	sshLong = `
		# ssh to pod, according the the podName, and password, userName
		# podName is need, userName or password will alter to input if not provided
		# only allowed to ssh pods which you owns the access authority, 
		# for instance, developers A belongs to department tctp, he can not
		# ssh to pods belongs to department tdtp
		# 
		# for instance
		# ssh according to podName
		# 	dockin-opsctl ssh dockin-test-20191012-182050448-0 -u admin -p admin -r default
		# ssh according to pod ip
		# dockin-opsctl ssh 192.168.1.1 -u admin -p admin -r default
		# ssh according to access token
		# dockin-opsctl ssh 192.168.1.1 --access-token foiudepjfpghuqwipr1028390eu8fihyedpqrhfuwospkal
`
	sshExample = `dockin-opsctl ssh dockin-test-20191012-182050448-0 -u admin -p admin -r default`

	accessTokenExpire int64 = 3600
)

func NewSSHCmd(configFlags *genericclioptions.ConfigFlags) *cobra.Command {
	opt := &SSHOption{}
	sshCmd := &cobra.Command{
		Use:                   "ssh to pod",
		DisableFlagsInUseLine: true,
		Short:                 "ssh to pod",
		Long:                  sshLong,
		Example:               sshExample,
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(opt.Complete(configFlags, cmd, args))
			utils.CheckErr(opt.Validate())
			utils.CheckErr(opt.Run())
		},
	}
	sshCmd.Flags().StringVarP(&opt.UserName, "user name", "u", opt.UserName, "pass user name")
	sshCmd.Flags().StringVarP(&opt.Password, "password", "p", opt.Password, "pass password")
	sshCmd.Flags().StringVarP(&opt.ContainerName, "container", "c", "op_server", "ContainerName name. If omitted, the first container in the pod will be chosen")
	sshCmd.Flags().StringVarP(&opt.AccessToken, "access-token", "a", opt.UserName, "access token to access the pod by ssh command, generate from auth command")
	sshCmd.Flags().StringVarP(&opt.User, "user", "s", opt.User, "run as user name, default to app")
	sshCmd.Flags().StringArrayVarP(&opt.Env, "env", "e", opt.Env, "exec env variable, like: a=1")
	sshCmd.Flags().StringVarP(&opt.WorkDir, "work-dir", "w", opt.WorkDir, "exec work directory")
	return sshCmd
}

type SSHOption struct {
	UserName      string
	Password      string
	PodName       string
	ContainerName string
	AccessToken   string
	Rule          string
	Namespace     string
	User          string
	Env           []string
	WorkDir       string
}

func (option *SSHOption) Validate() error {
	if option.PodName == "" {
		return errors.Errorf("pod name or ip is empty")
	}
	if option.AccessToken == "" && option.UserName == "" {
		return errors.Errorf("user name or access token must be assigned")
	}
	return nil
}

func (option *SSHOption) Complete(configFlags *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.Errorf("%s\n%s",
			"no resource type or name provided", "See 'dockin-opsctl ssh -h' for help and examples.")
	} else {
		option.PodName = args[0]
	}
	if option.UserName == "" {
		if u, err := user.Current(); err != nil {
			option.UserName = u.Username
		}
	}
	option.Rule, _ = cmd.Flags().GetString("rule")
	option.Namespace, _ = cmd.Flags().GetString("namespace")

	return nil
}

func (option *SSHOption) Run() error {
	var (
		podInfo *ssh.PodInfo
		err     error
	)

	if err := option.Login(); err != nil {
		log.Debugf("validate user failed for userName=%s, err=%s",
			option.UserName, err.Error())
		log.Output(err.Error())
		return err
	}

	if podInfo, err = ssh.ValidatePod(option.Rule, option.PodName, option.UserName); err != nil {
		log.Debugf("validate pod failed, err=%s, podName=%s, rule=%s", err.Error(), option.PodName, option.Rule)
		return fmt.Errorf("validata pod failed podName = %s, rule = %s，err=%s", option.PodName, option.Rule, err.Error())
	}

	if base.IsIp(option.PodName) {
		option.PodName = podInfo.PodName
	}

	//t.ClusterId = podInfo.ClusterId
	//t.HostIp = podInfo.HostIp
	//t.PodIp = podInfo.PodIp

	proto := protocol.NewProto()
	proto.UserName = option.UserName
	proto.Password = option.Password
	proto.AccessToken = option.AccessToken
	proto.Env = option.Env
	proto.WorkDir = option.WorkDir
	proto.Name = option.PodName
	proto.Container = option.ContainerName

	return ssh.RunBash(proto)
}

func (option *SSHOption) Login() error {
	if option.UserName == "" && option.AccessToken == "" {
		return errors.New("user name or access token is empty")
	}

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

	baseUrl := common.GetCommonUrlByCmd("ctrl/login")
	reqUrl := fmt.Sprintf("%s?userName=%s&password=%s&rule=%s", baseUrl, option.UserName, passAes, option.Rule)

	hder := http.Header{}
	hder.Set(ssh.DockinAesHeader, option.AccessToken)
	body, err := utils.HttpGetWithHeader(reqUrl, time.Second*3, hder)
	if err != nil {
		log.Debugf("timeout in validate account, err=%s, podName=%s, rule=%s", err.Error(), option.UserName, option.Password)
		return fmt.Errorf("timeout in login, try again")
	}
	opserver := &ssh.OPServerResult{}
	log.Debugf("get auth data from opserver success, %s", string(body))
	if err := jsoniter.Unmarshal(body, opserver); err != nil {
		log.Debugf("failed to auth account %s with server, as unmarshal err %v", option.UserName, err)
		return fmt.Errorf("failed to login account %s with server, as http data invalid", option.UserName)
	}

	if opserver.Code != 0 {
		log.Debugf("failed to auth account %s with server, errMsg=%s", option.UserName, opserver.Message)
		return fmt.Errorf("failed to login %s with server, errMsg=%s", option.UserName, opserver.Message)
	}
	log.Debugf("success auth userName=%s, access token=%v", option.UserName, opserver.Data)
	ud := &ssh.UserIdentity{}
	if err := jsoniter.UnmarshalFromString(opserver.Data.(string), ud); err != nil {
		log.Debugf("failed to unmarshal login response, errMsg=%s", option.UserName, opserver.Message)
		return fmt.Errorf("failed to unmarshal login response, %s", option.UserName)
	}

	option.UserName = ud.UserName
	option.AccessToken = ud.AccessToken
	option.Rule = ud.Rule
	option.Password = ud.Password
	//t.IsAdmin = t.UserName == "admin"
	//t.LoginTime = time.Now()
	//t.IsLogin = true

	log.Debugf("login success current context = %#v", option)
	return nil
}
