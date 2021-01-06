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

package common

import (
	"errors"
	"net/http"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/log"
	"github.com/webankfintech/dockin-opagent/internal/utils/aes"

	jsoniter "github.com/json-iterator/go"
)

var (
	tokenHeaderKey = "access-token"
	aesKey         = "doc-opserver8085"
	user           = "dockin-opagent"

	ErrNoAccessToken = errors.New("no access token found")
	ErrCreateSig     = errors.New("create encrypt sig err")
	ErrUnmarshal     = errors.New("err in unmarshal the token")
	ErrTokenExpired  = errors.New("access token is expired")
	ErrInvalidUser   = errors.New("invalid user identity")
)

type Identity struct {
	UserName   string    `json:"userName"`
	Expire     int64     `json:"expire"`
	CreateTime time.Time `json:"createTime"`
}

func ValidateRequest(req *http.Request, uid string) error {
	token := req.Header.Get(tokenHeaderKey)
	if token == "" {
		return ErrNoAccessToken
	}
	aes, err := aes.NewAes(aesKey)
	if err != nil {
		return ErrCreateSig
	}

	out, err := aes.AesDecrypt(token)
	id := &Identity{}
	if err := jsoniter.Unmarshal([]byte(out), id); err != nil {
		log.Logger.Warnf("failed to unmarshal token, err=%v, uid=%s", err, uid)
		return ErrUnmarshal
	}
	if id.UserName != user {
		log.Logger.Warnf("invalid user name=%s, uid=%s", id.UserName, uid)
		return ErrInvalidUser
	}
	if int64(time.Now().Sub(id.CreateTime).Seconds()) > id.Expire {
		log.Logger.Warnf("access token is expired, create at %v, uid=%s", id.CreateTime, token)
		return ErrTokenExpired
	}

	return nil
}

func ValidateRequestV2(req *http.Request, uid string) error {
	req.ParseForm()
	token := req.Form.Get(tokenHeaderKey)
	if token == "" {
		return ErrNoAccessToken
	}
	aes, err := aes.NewAes(aesKey)
	if err != nil {
		return ErrCreateSig
	}

	out, err := aes.AesDecrypt(token)
	id := &Identity{}
	if err := jsoniter.Unmarshal([]byte(out), id); err != nil {
		log.Logger.Warnf("failed to unmarshal token, err=%v, uid=%s", err, uid)
		return ErrUnmarshal
	}
	if id.UserName != user {
		log.Logger.Warnf("invalid user name=%s, uid=%s", id.UserName, uid)
		return ErrInvalidUser
	}
	if int64(time.Now().Sub(id.CreateTime).Seconds()) > id.Expire {
		log.Logger.Warnf("access token is expired, create at %v, uid=%s", id.CreateTime, token)
		return ErrTokenExpired
	}

	return nil
}
