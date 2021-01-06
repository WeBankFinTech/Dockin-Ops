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

package ctrl

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/common"

	"github.com/webankfintech/dockin-opserver/internal/api"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/model"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/webankfintech/dockin-opserver/internal/utils/aes"
	"github.com/webankfintech/dockin-opserver/internal/utils/ip"
	"github.com/webankfintech/dockin-opserver/internal/utils/trace"

	v1 "k8s.io/api/core/v1"
)

type Control struct {
	indexTmpl   *template.Template
	redisClient *redis.RedisClient
	cm          *client.Manager
	access      *Access
	allow       *AllowCmd
	account     *Account
	version     *Version
}

func NewControl(cm *client.Manager, r *redis.RedisClient) *Control {
	c := &Control{}
	indexTmpl, err := template.New("manager").Parse(indexTemplate)
	if err != nil {
		log.Logger.Panicf(err.Error())
	}
	c.indexTmpl = indexTmpl
	c.allow = &AllowCmd{redisClient: r}
	c.access = &Access{redisClient: r}
	c.account = &Account{redisClient: r}
	c.version = &Version{redisClient: r}
	c.redisClient = r
	c.cm = cm

	http.HandleFunc("/v1/dockin/opserver/ctrl/manager", c.ManagerIndex)

	http.HandleFunc("/v1/dockin/opserver/ctrl/auth", c.Auth)
	http.HandleFunc("/v1/dockin/opserver/ctrl/login", c.Login)

	http.HandleFunc("/v1/dockin/opserver/ctrl/addRawCmd", c.AddRawCmd)
	http.HandleFunc("/v1/dockin/opserver/ctrl/deleteRawCmd", c.DeleteRawCmd)
	http.HandleFunc("/v1/dockin/opserver/ctrl/addCmd", c.AddCmd)
	http.HandleFunc("/v1/dockin/opserver/ctrl/deleteCmd", c.DeleteCmd)
	http.HandleFunc("/v1/dockin/opserver/ctrl/deleteIP", c.DeleteIP)
	http.HandleFunc("/v1/dockin/opserver/ctrl/addWhitelist", c.AddWhitelist)
	http.HandleFunc("/v1/dockin/opserver/ctrl/getCmdList", c.GetCmdList)
	http.HandleFunc("/v1/dockin/opserver/ctrl/getPodByName", c.GetPodByName)
	http.HandleFunc("/v1/dockin/opserver/ctrl/addAccount", c.AddAccount)
	http.HandleFunc("/v1/dockin/opserver/ctrl/deleteAccount", c.DeleteAccount)
	http.HandleFunc("/v1/dockin/opserver/ctrl/updateVersion", c.UpdateVersion)
	return c
}

type PodInfo struct {
	PodName   string `json:"podName"`
	ClusterId string `json:"clusterId"`
	PodIp     string `json:"podIp"`
	HostIp    string `json:"hostIp"`
}

func (c *Control) ManagerIndex(writer http.ResponseWriter, request *http.Request) {
	data := &model.ControlDTO{}
	rawCmd, commonCmd := c.allow.LoadFromCache()

	for _, c := range commonCmd {
		data.Command = append(data.Command, model.CMD{Cmd: c})
	}
	for _, c := range rawCmd {
		data.RawCommand = append(data.RawCommand, model.CMD{Cmd: c})
	}

	d := c.access.LoadFromCache()
	for rule, ipList := range d {
		data.WhiteList = append(data.WhiteList, model.WhiteList{
			Rule: rule,
			IPList: func() []model.IP {
				var ips []model.IP
				for _, ip := range ipList {
					ips = append(ips, model.IP{Addr: ip})
				}
				return ips
			}(),
		})
	}

	userList := c.account.LoadAccountFromCache()
	for _, user := range userList {
		data.Account = append(data.Account, model.Account{UserName: user})
	}

	data.Version = c.version.LoadVersionFromCache()

	log.Logger.Infof("success get manager index %+v", data)
	c.indexTmpl.Execute(writer, data)
}

func (c *Control) Login(writer http.ResponseWriter, request *http.Request) {
	traceId := trace.TraceID()
	request.ParseForm()
	token := request.Header.Get("access-token")
	if token != "" {
		c.loginByAccessToken(writer, token, traceId)
	} else {
		c.loginByAccount(writer, request, traceId)
	}
}

func (c *Control) loginByAccessToken(writer http.ResponseWriter, token, traceId string) {
	log.Logger.Infof("start to loginByAccessToken token=%s,traceId=%s", token, traceId)
	ud, err := api.ParseAccessToken(token, traceId)
	if err != nil {
		log.Logger.Warnf("ParseAccessToken failed,err=%s,traceId=%s", err.Error(), traceId)
		result := model.FailedOpsResult(errors.Errorf("%s,traceId=%s", err.Error(), traceId))
		writer.Write(result.ToByte())
		return
	}
	if int64(time.Now().Sub(ud.CreateTime).Seconds()) > ud.Expire {
		err := errors.Errorf("access token is expired, userName=%s, create at %v,traceId=%s", ud.UserName, ud.CreateTime, traceId)
		log.Logger.Warnf("access token is expired, create at %v, token=%s,traceId=%s", ud.CreateTime, token, traceId)
		result := model.FailedOpsResult(err)
		writer.Write(result.ToByte())
		return
	}
	ud.AccessToken = token
	udstr, _ := jsoniter.MarshalToString(ud)
	result := model.SuccessOpsResult(udstr)
	log.Logger.Infof("loginByAccessToken success %s, user info=%#v,traceId=%s", ud, traceId)
	writer.Write(result.ToByte())
}

func (c *Control) loginByAccount(writer http.ResponseWriter, request *http.Request, traceId string) {
	log.Logger.Infof("start to loginByAccount,traceId=%s", traceId)
	userName := request.Form.Get("userName")
	password := request.Form.Get("password")
	rule := request.Form.Get("rule")
	if userName == "" {
		log.Logger.Infof("request user name is empty,traceId=%s", traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("user name is empty,traceId=%s", traceId)).ToByte())
		return
	}
	if password == "" {
		log.Logger.Infof("request password is empty,traceId=%s", traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("password is empty,traceId=%s", traceId)).ToByte())
		return
	}
	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		log.Logger.Warnf("NewAes failed err=%s,traceId=%s", err.Error(), traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("aes decode failed,traceId=%s", traceId)).ToByte())
		return
	}

	password, err = aes.AesDecrypt(password)
	if err != nil {
		log.Logger.Warnf("AesDecrypt err=%s,traceId=%s", err.Error(), traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("aes decode failed,traceId=%s", traceId)).ToByte())
		return
	}

	if rule == "" {
		rule = "default"
	}
	if err := c.account.AccountAuth(userName, password, traceId); err != nil {
		log.Logger.Warnf("AccountAuth fail,err=%s,traceId=%s", err.Error(), traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("%s,traceId=%s", err.Error(), traceId)).ToByte())
		return
	}

	ac := model.NewUserIdentity(userName, password, rule)
	acStr, err := ac.ToString()
	if err != nil {
		log.Logger.Warnf("failed to create access token string, ac=%v, err=%v,traceId=%s", ac, err, traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("password is empty,traceId=%s", traceId)).ToByte())
		return
	}
	ac.AccessToken = acStr
	udstr, _ := jsoniter.MarshalToString(ac)
	result := model.SuccessOpsResult(udstr)
	writer.Write(result.ToByte())
	log.Logger.Infof("loginByAccount userName=%s success,traceId=%s", userName, traceId)
}

func (c *Control) Auth(writer http.ResponseWriter, request *http.Request) {
	traceId := trace.TraceID()
	request.ParseForm()
	userName := request.Form.Get("userName")
	password := request.Form.Get("password")
	rule := request.Form.Get("rule")
	log.Logger.Infof("recv auth request,userName=%s ,password=%s,rule=%s traceId=%s", userName, password, rule, traceId)
	if userName == "" {
		log.Logger.Infof("request user name is empty,traceId=%s", traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("user name is empty,traceId=%s", traceId)).ToByte())
		return
	}

	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		log.Logger.Warnf("New aes failed err=%s,traceId=%s", err.Error(), traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("aes decode failed,traceId=%s", traceId)).ToByte())
		return
	}

	password, err = aes.AesDecrypt(password)
	if err != nil {
		log.Logger.Warnf("aes decrypt err=%s,traceId=%s", err.Error(), traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("aes decode failed,traceId=%s", traceId)).ToByte())
		return
	}

	if rule == "" {
		rule = "default"
	}
	if password == "" {
		log.Logger.Infof("request password is empty,traceId=%s", traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("password is empty,traceId=%s", traceId)).ToByte())
		return
	}
	if err := c.account.AccountAuth(userName, password, traceId); err != nil {
		log.Logger.Warnf("account auth failed,err=%s,traceId=%s", err.Error(), traceId)
		writer.Write(model.FailedOpsResult(err).ToByte())
		return
	}

	ac := model.NewUserIdentity(userName, password, rule)
	acStr, err := ac.ToString()
	if err != nil {
		log.Logger.Warnf("failed to create access token string, ac=%v, err=%s,traceId", ac, err.Error(), traceId)
		writer.Write(model.FailedOpsResult(fmt.Errorf("password is empty,traceId=%s", traceId)).ToByte())
		return
	}

	result := model.SuccessOpsResult(acStr)
	log.Logger.Infof("success create access token for user name=%s, access token=%s,traceId=%s", userName, acStr, traceId)
	writer.Write(result.ToByte())
}

func (c *Control) AddRawCmd(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	cmdName := request.Form.Get("cmdName")
	if cmdName == "" {
		log.Logger.Infof("cmdName is empty")
		writer.Write([]byte("cmdName is empty"))
		return
	}

	log.Logger.Infof("add raw command receive %s", cmdName)

	cmdList := strings.Split(cmdName, ",")
	if len(cmdList) == 0 {
		log.Logger.Warnf("cmd list is emtpy %s", cmdName)
		writer.Write([]byte("cmd list is empty"))
		return
	}
	if err := c.allow.AddRaw(cmdList); err != nil {
		log.Logger.Warnf("failed to add raw cmd=%s, err=%s", cmdName, err.Error())
		writer.Write([]byte("failed to add raw cmd"))
		return
	}

	writer.Write([]byte("success"))
	log.Logger.Infof("add raw cmd success %s\n", cmdName)
}

func (c *Control) DeleteRawCmd(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	cmdName := request.Form.Get("cmdName")
	if cmdName == "" {
		log.Logger.Infof("cmdName is empty")
		writer.Write([]byte("cmdName is empty"))
		return
	}

	log.Logger.Infof("delete raw command receive %s", cmdName)
	cmdList := strings.Split(cmdName, ",")

	if err := c.allow.RemoveRaw(cmdList); err != nil {
		log.Logger.Warnf("failed remove raw command %s", cmdName)
		writer.Write([]byte("failed remove raw command"))
		return
	}

	writer.Write([]byte("success"))
	log.Logger.Infof("delete raw cmd %s success", cmdName)
}

func (c *Control) AddCmd(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	cmdName := request.Form.Get("cmdName")
	log.Logger.Infof("add common command receive %s", cmdName)
	if cmdName == "" {
		log.Logger.Infof("cmdName is empty")
		writer.Write([]byte("cmdName is empty"))
		return
	}

	cmdList := strings.Split(cmdName, ",")
	if len(cmdList) == 0 {
		log.Logger.Warnf("cmd list is emtpy %s", cmdName)
		writer.Write([]byte("cmd list is empty"))
		return
	}
	if err := c.allow.AddCommon(cmdList); err != nil {
		log.Logger.Warnf("failed to add common cmd=%s, err=%s", cmdName, err.Error())
		writer.Write([]byte("failed to add common cmd"))
		return
	}

	writer.Write([]byte("success"))
	log.Logger.Infof("add common cmd success %s\n", cmdName)
}

func (c *Control) DeleteCmd(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	cmdName := request.Form.Get("cmdName")
	if cmdName == "" {
		log.Logger.Infof("cmd name is empty")
		writer.Write([]byte("cmd name is empty"))
		return
	}
	log.Logger.Infof("delete common command receive %s", cmdName)
	cmdList := strings.Split(cmdName, ",")

	if err := c.allow.RemoveCommon(cmdList); err != nil {
		log.Logger.Warnf("failed remove common command %s", cmdName)
		writer.Write([]byte("failed remove common command"))
		return
	}

	writer.Write([]byte("success"))
	log.Logger.Infof("delete common cmd %s success", cmdName)
}

func (c *Control) AddWhitelist(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	whiteIp := request.Form.Get("ip")
	rule := request.Form.Get("rule")
	if rule == "" || whiteIp == "" {
		log.Logger.Infof("rule or ip is empty")
		writer.Write([]byte("rule or ip is empty"))
		return
	}

	log.Logger.Infof("add whitelist receive, rule=%s, ip=%s", rule, whiteIp)

	ips := strings.Split(strings.TrimSpace(whiteIp), ",")
	if err := c.access.AddAccess(rule, ips); err != nil {
		log.Logger.Warnf("failed to add whitelist, rule=%s, ip=%s, err=%s", rule, whiteIp, err.Error())
		writer.Write([]byte("failed to add whitelist"))
		return
	}

	writer.Write([]byte("success"))
	log.Logger.Infof("add whitelist success, rule=%s, ip=%s", rule, whiteIp)
}

func (c *Control) DeleteIP(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	whiteIp := request.Form.Get("ip")
	rule := request.Form.Get("rule")
	if rule == "" || whiteIp == "" {
		log.Logger.Infof("rule or ip is empty")
		writer.Write([]byte("rule or ip is empty"))
		return
	}

	log.Logger.Infof("delete whitelist receive, rule=%s, ip=%s", rule, whiteIp)

	ips := strings.Split(strings.TrimSpace(whiteIp), ",")
	if err := c.access.RemoveAccess(rule, ips); err != nil {
		log.Logger.Warnf("failed to remove whitelist, rule=%s, ip=%s, err=%s", rule, whiteIp, err.Error())
		writer.Write([]byte("failed to remove whitelist"))
		return
	}

	writer.Write([]byte("success"))
	log.Logger.Infof("delete whitelist success, rule=%s, ip=%s", rule, whiteIp)
}

func (c *Control) GetCmdList(writer http.ResponseWriter, request *http.Request) {
	mm := map[string][]string{}
	rawCmd, commonCmd := c.allow.LoadFromCache()
	mm["raw"] = rawCmd
	mm["common"] = commonCmd
	data, _ := jsoniter.Marshal(mm)
	log.Logger.Infof("command:%s", commonCmd)
	writer.Write(data)
}

func (c *Control) GetPodByName(writer http.ResponseWriter, request *http.Request) {
	var (
		podName   string
		podIp     string
		clusterId string
		hostIP    string
		opsOpts   *model.OpsOption
	)
	request.ParseForm()
	rule := request.Form.Get("rule")
	podName = request.Form.Get("podName")
	userName := request.Form.Get("userName")
	if rule == "" || podName == "" {
		log.Logger.Infof("rule or podName is empty")
		writer.Write([]byte("rule or podName is empty"))
	}

	reqIp := ip.GetIp(request)
	log.Logger.Infof("get pod by name, podName=%s, rule=%s, ip=%s", podName, rule, reqIp)

	if err := c.cm.SShAllow(rule, reqIp, userName); err != nil {
		log.Logger.Infof(err.Error())
		writer.Write([]byte(fmt.Sprintf("rule = %s, ip = %s access to pod = %s is forbidded", rule, reqIp, podName)))
		return
	}

	opsOpts = &model.OpsOption{}
	opsOpts.Name = podName
	if err := api.SetPodOption(opsOpts); err != nil {
		res := model.FailedOpsResult(errors.Errorf("get podInfo from rm failed podName=%s, err=%s",
			opsOpts.Name, err.Error()))
		writer.Write(res.ToByte())
		return
	}

	podName = opsOpts.Name

	podYAMLKey := keys.PodYAMLKey(podName)
	podStr, err := c.redisClient.Get(podYAMLKey)
	if err != nil {
		err := errors.Errorf("get pod from redis failed, podName=%s, err=%s", podName, err.Error())
		log.Logger.Warnf(err.Error())
		res := model.FailedOpsResult(err)
		writer.Write(res.ToByte())
		return
	}
	pod := v1.Pod{}
	if err := jsoniter.Unmarshal([]byte(podStr.(string)), &pod); err != nil {
		log.Logger.Warnf("unmarshal pod str failed, podName=%s, content=%s", podName, podStr)
		err := errors.Errorf("pod content invalid")
		res := model.FailedOpsResult(err)
		writer.Write(res.ToByte())
		return
	}

	hostIP = opsOpts.HostIP
	podIp = opsOpts.PodIp
	clusterId = opsOpts.ClusterId
	podInfo := &PodInfo{
		PodName:   podName,
		ClusterId: clusterId,
		PodIp:     podIp,
		HostIp:    hostIP}
	res := model.SuccessOpsResult(podInfo)
	writer.Write(res.ToByte())
	log.Logger.Infof("succ to response get pod by name request,res=%s", res.ToByte())

}

func (c *Control) AddAccount(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	userName := request.Form.Get("userName")
	if userName == "" {
		log.Logger.Infof("user name  is empty")
		writer.Write([]byte("user name is empty"))
		return
	}

	log.Logger.Infof("add account, userName=%s", userName)

	if err := c.account.AddAccount(userName, nameField); err != nil {
		log.Logger.Warnf("failed to add acount, as %s userName=%s, nameField=%s",
			err.Error(), userName, nameField)
		writer.Write([]byte("add account failed"))
		return
	}

	log.Logger.Infof("success add account, userName=%s, nameField=%s", userName, nameField)
	writer.Write([]byte("success"))
}

func (c *Control) DeleteAccount(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	userName := request.Form.Get("userName")
	log.Logger.Infof("delete account, userName=%s", userName)
	if userName == "" {
		log.Logger.Infof("user name is empty")
		writer.Write([]byte("user name is empty"))
		return
	}

	if err := c.account.DeleteAccount(userName); err != nil {
		log.Logger.Warnf("failed to delete account, as %s, userName=%s", err.Error(), userName)
		writer.Write([]byte("delete account failed"))
		return
	}

	log.Logger.Infof("success delete account, userName=%s", userName)
	writer.Write([]byte("success"))
}

func (c *Control) UpdateVersion(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	version := request.Form.Get("version")
	log.Logger.Infof("update version =%s", version)
	if version == "" {
		log.Logger.Infof("version is empty")
		writer.Write([]byte("version is empty"))
		return
	}

	if err := c.version.UpdateVersion(version); err != nil {
		log.Logger.Warnf("failed to update version=%s,err=%s", version, err.Error())
		writer.Write([]byte("update version failed"))
		return
	}

	log.Logger.Infof("success update version=%s", version)
	writer.Write([]byte("success"))
}
