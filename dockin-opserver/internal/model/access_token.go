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

package model

import (
	"fmt"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/utils/aes"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type UserIdentity struct {
	UserName    string    `json:"userName"`
	Password    string    `json:"password"`
	Rule        string    `json:"rule"`
	Expire      int64     `json:"expire"`
	CreateTime  time.Time `json:"createTime"`
	AccessToken string    `json:"accessToken"`
}

var (
	expireAt int64 = 3600
)

func NewUserIdentity(userName, password, rule string) *UserIdentity {
	return &UserIdentity{
		UserName:   userName,
		Password:   password,
		Rule:       rule,
		Expire:     expireAt,
		CreateTime: time.Now(),
	}
}

func (at *UserIdentity) ToString() (string, error) {
	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		err = errors.Errorf("create encrypt failed, err=%s", err.Error())
		log.Logger.Warnf(err.Error())
		return "", err
	}

	str, err := jsoniter.MarshalToString(at)
	if err != nil {
		log.Logger.Warnf("failed to unmarshal access token to string, %v, %v", at, err)
		return "", fmt.Errorf("failed to unmarshal access token to string")
	}
	return aes.AesEncrypt(str)
}

func OpagentAccessToken() string {
	expire := 3600
	ui := &UserIdentity{
		UserName:   "dockin-opagent",
		Expire:     int64(expire),
		CreateTime: time.Now(),
	}
	token, _ := ui.ToString()
	return token
}
