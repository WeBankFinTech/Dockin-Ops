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

package client

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/log"

	"go.uber.org/zap"

	"github.com/webankfintech/dockin-opserver/internal/common"

	"github.com/stretchr/testify/assert"
)

func TestWhitelist(t *testing.T) {
	rc, err := redis.NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}

	t.Run("test init", func(t *testing.T) {
		w := newWhitelist(rc)
		ch := make(chan struct{})
		err := w.initialize(ch)
		assert.NoError(t, err)
		w.printWhitelist()
	})

	log.CommandLogger.Info("ssh",
		zap.String("operator", "bruceliu"),
		zap.String("ip", "jumper ip"),
		zap.String("command", "ls -al"),
		zap.String("timestamp", "456"),
		zap.String("podName", "podName"),
		zap.String("podIp", "123"))

	t.Run("test update", func(t *testing.T) {
		w := newWhitelist(rc)
		ch := make(chan struct{})
		err := w.initialize(ch)
		assert.NoError(t, err)
		w.printWhitelist()

		content := []byte(`{
  "cls-1": [
    "192.168.1.1"
  ],
  "cls-2": [
    "192.168.1.2"
  ]
}`)
		confPath := common.GetConfPath()
		whitelistfile := filepath.Join(confPath, "whitelist.json")
		assert.NoError(t, ioutil.WriteFile(whitelistfile, content, 0777))
		time.Sleep(time.Second)
		w.printWhitelist()
	})

	t.Run("test Allow", func(t *testing.T) {
		w := newWhitelist(rc)
		ch := make(chan struct{})
		err := w.initialize(ch)
		assert.NoError(t, err)
		w.printWhitelist()
		assert.Error(t, w.allow("cls-1", "error"))
		assert.NoError(t, w.allow("cls-2", "192.168.1.2"))
	})
}
