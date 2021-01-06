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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/log"
)

var nameField = "name"

type Account struct {
	redisClient *redis.RedisClient
}

func (a *Account) LoadAccountFromCache() []string {
	userNameList, err := a.redisClient.SMembers(keys.GetAccountKey())
	if err != nil {
		log.Logger.Warnf("failed to get all account, accountKey=%s, err=%s",
			keys.GetAccountKey(), err.Error())
	}
	return userNameList

}

func (a *Account) AddAccount(userName, nameFile string) error {
	err := a.redisClient.SAdd(keys.GetAccountKey(), []string{userName})
	if err != nil {
		log.Logger.Warnf("failed to SAdd, accountKey=%s, userName=%s, nameFile=%s, err=%s",
			keys.GetAccountKey(), userName, nameFile, err.Error())
		return err
	}

	userKey := keys.GetUserNameKey(userName)
	if err := a.redisClient.HSet(userKey, nameField, userName); err != nil {
		log.Logger.Warnf("failed to hSet userKey=%s,nameField=%s userName=%s, err=%s",
			userKey, nameField, userName, err.Error())
		return err
	}

	return nil
}

func (a *Account) DeleteAccount(userName string) error {
	err := a.redisClient.SRem(keys.GetAccountKey(), []string{userName})
	if err != nil {
		log.Logger.Warnf("failed to delete key=%s, userName=%s", keys.GetAccountKey(), userName)
		return err
	}

	userKey := keys.GetUserNameKey(userName)
	if err := a.redisClient.Del(userKey); err != nil {
		log.Logger.Warnf("failed to del userKey=%s userName=%s, err=%s",
			userKey, userName, err.Error())
		return err
	}
	return nil
}

func getUserNameKey(userName string) string {
	return fmt.Sprintf("%s_account", userName)
}

func (a *Account) ValidateAccountMD5(userName, password string) error {
	ps, err := a.redisClient.Get(getUserNameKey(userName))
	if err != nil {
		log.Logger.Warnf("failed to get key=%s, userName=%s", "account", userName)
		return err
	}
	h := md5.New()
	h.Write([]byte("dockin" + ps.(string)))
	passRedisMd5 := hex.EncodeToString(h.Sum(nil))
	if passRedisMd5 != password {
		err = fmt.Errorf("password is invalid")
		log.Logger.Infof("account validate failed, userName=%s, provide password=%s, cache value=%s",
			userName, password, ps)
		return err
	}

	return nil
}

func (a *Account) AccountAuth(userName, password, traceId string) error {
	log.Logger.Infof("start to AccountAuth userName=%s,password=%s,traceId", userName, password, traceId)

	accountList := config.OpsConfig.Accounts
	log.Logger.Infof("account info list=%v", accountList)
	for _, account := range accountList {
		if strings.EqualFold(account.Account.UserName, userName) {
			if strings.EqualFold(account.Account.Passwd, password) {
				return nil
			} else {
				log.Logger.Warnf("validate fixAccount input %s:%s failed,config account %s:%s,traceId=%s",
					userName, password, account.Account.UserName, account.Account.Passwd, traceId)
				return errors.Errorf("validate fix account input,userName=%s", userName)
			}
		}
	}

	return errors.Errorf("validate fix account input,userName=%s", userName)

}
