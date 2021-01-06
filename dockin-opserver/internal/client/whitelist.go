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
	"strings"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/config"

	"github.com/webankfintech/dockin-opserver/internal/log"
	"github.com/webankfintech/dockin-opserver/internal/utils/cmap"

	"github.com/pkg/errors"
)

type whitelist struct {
	ruleMap        cmap.ConcurrentMap
	redisClient    *redis.RedisClient
	lastModifyTime string
	updateTicker   *time.Ticker
}

func newWhitelist(rc *redis.RedisClient) *whitelist {
	w := &whitelist{}
	w.ruleMap = cmap.New()
	w.redisClient = rc
	return w
}

func (w *whitelist) initialize(ch chan struct{}) error {
	data := w.LoadFromCache()
	log.Logger.Infof("load whitelist from redis data %s", data)

	for rule, wl := range data {
		w.ruleMap.Set(rule, wl)
		log.Logger.Infof("success add whitelist rule=%s, ipList=%+v", rule, wl)
	}
	w.lastModifyTime = time.Now().Format("2006:01:02 03:04:05")
	w.addFRedisListener(ch)
	return nil
}

func (w *whitelist) LoadFromCache() map[string][]string {
	mem := make(map[string][]string)
	ruleList, err := w.redisClient.SMembers(keys.GetRuleRedisKey())
	if err != nil {
		log.Logger.Warnf(err.Error())
		return mem
	}

	for _, rule := range ruleList {
		ruleKey := keys.GetRedisWhiteKeyByRule(rule)
		tmp, err := w.redisClient.SMembers(ruleKey)
		if err != nil {
			log.Logger.Warnf(err.Error())
			continue
		}
		mem[rule] = tmp
	}
	return mem
}

func (w *whitelist) allow(rule string, ip string) error {
	data, ok := w.ruleMap.Get(rule)
	if !ok {
		err := errors.Errorf("no rule whitelist exist for rule=%s", rule)
		log.Logger.Infof(err.Error())
		return err
	}

	wl := data.([]string)
	for _, v := range wl {
		if strings.EqualFold(v, ip) {
			return nil
		}
	}

	err := errors.Errorf("ip = %s is not in rule=%s white list", ip, rule)
	log.Logger.Infof(err.Error())
	return err
}

func (w *whitelist) addFRedisListener(ch chan struct{}) error {
	w.updateTicker = time.NewTicker(time.Millisecond * time.Duration(config.OpsConfig.WhileListUpdateTime))
	go func() {
		for {
			select {
			case <-ch:
				w.updateTicker.Stop()
				log.Logger.Infof("exit update whileList cache ticker")
				return
			case <-w.updateTicker.C:
				log.Logger.Debugf("ticker to update whileList cache")
				w.updateWhitelist()
			}
		}
	}()
	return nil
}

func (w *whitelist) updateWhitelist() {
	data := w.LoadFromCache()
	log.Logger.Infof("load whitelist from redis data %s", data)

	for _, cid := range w.ruleMap.Keys() {
		_, ok := data[cid]
		if !ok {
			w.ruleMap.Remove(cid)
			log.Logger.Infof("remove rule=%s whitelist", cid)
		}
	}

	for cid, data := range data {
		w.ruleMap.Set(cid, data)
		log.Logger.Infof("update or add whitelist rule=%s, ip list=%+v", cid, data)
	}
	w.lastModifyTime = time.Now().Format("2006:01:02 03:04:05")
	log.Logger.Infof("last update whileList time:%s", w.lastModifyTime)
}

func (w *whitelist) printWhitelist() {
	for k, v := range w.ruleMap.Items() {
		log.Logger.Infof("rule=%s, ipList=%+v", k, v)
	}
}
