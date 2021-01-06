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

package redis

import (
	"github.com/webankfintech/dockin-opserver/internal/log"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_InitClusterClient(t *testing.T) {
	client, err := NewRedisClient()
	assert.NoError(t, err)
	client.Del("goRedis")
	client.Get("goRedis")
	client.Set("goRedis", "hello go-redis", 0)
	client.Get("goRedis")
	client.Close()

}

func Test_Hash(t *testing.T) {
	rc, err := NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}

	key := fmt.Sprintf("%s_account", "v_wbzfhe")
	field := "name"
	value := "v_wbzfhe"
	if err := rc.HSet(key, field, value); err != nil {
		log.Logger.Warnf("redis HSet failed,err=%s", err.Error())
	}

	field = "age"
	value = "30"
	if err := rc.HSet(key, field, value); err != nil {
		log.Logger.Warnf("redis HSet failed,err=%s", err.Error())
	}

	field = "addr"
	value = "zhongguo"
	if err := rc.HSet(key, field, value); err != nil {
		log.Logger.Warnf("redis HSet failed,err=%s", err.Error())
	}

	if ise, err := rc.HExist(key, field); err != nil {
		log.Logger.Warnf("redis HExist failed,err=%s", err.Error())
	} else if ise {
		fmt.Println(rc.HGet(key, field))
	}

	fmt.Println(rc.HGetAll(key))
}

func Test_DeleteHash(t *testing.T) {
	rc, err := NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}

	key := fmt.Sprintf("%s_account", "v_wbzfhe")

	fmt.Println(rc.HGetAll(key))
	field := "age"
	rc.HDel(key, field)
	fmt.Println(rc.HGetAll(key))
	field = "name"
	rc.HDel(key, field)
	fmt.Println(rc.HGetAll(key))
	field = "addr"
	rc.HDel(key, field)
	fmt.Println(rc.HGetAll(key))
}

func Test_DeleteAllHash(t *testing.T) {
	rc, err := NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}

	key := fmt.Sprintf("%s_account", "v_wbzfhe")

	fmt.Println(rc.HGetAll(key))
	//rc.Del(key)
}

func Test_Hexsit(t *testing.T) {
	rc, err := NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}
	key := fmt.Sprintf("%s_account", "v_wbzfhe")

	fmt.Println(rc.HGetAll(key))
	field := "ages"

	if ise, err := rc.HExist(key, field); err != nil {
		log.Logger.Warnf("redis HExist failed,err=%s", err.Error())
	} else if ise {
		fmt.Println(rc.HGet(key, field))
	}
}
