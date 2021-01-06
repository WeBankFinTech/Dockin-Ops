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
	"github.com/webankfintech/dockin-opserver/internal/cache/keys"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/log"
)

type Access struct {
	redisClient *redis.RedisClient
}

func (a *Access) LoadFromCache() map[string][]string {
	mem := make(map[string][]string)
	ruleList, err := a.redisClient.SMembers(keys.GetRuleRedisKey())
	if err != nil {
		log.Logger.Warnf(err.Error())
		return mem
	}

	for _, rule := range ruleList {
		ruleKey := keys.GetRedisWhiteKeyByRule(rule)
		tmp, err := a.redisClient.SMembers(ruleKey)
		if err != nil {
			log.Logger.Warnf(err.Error())
			continue
		}
		mem[rule] = tmp
	}
	return mem
}

func (a *Access) AddAccess(rule string, ips []string) error {
	if err := a.redisClient.SAdd(keys.GetRuleRedisKey(), []string{rule}); err != nil {
		log.Logger.Warnf("add access err=%s", err.Error())
		return err
	}

	ruleIpKey := keys.GetRedisWhiteKeyByRule(rule)
	if err := a.redisClient.SAdd(ruleIpKey, ips); err != nil {
		log.Logger.Warnf("failed to add ips=%v, rule=%s", ips, rule)
		return err
	}

	return nil
}

func (a *Access) RemoveAccess(rule string, ips []string) error {
	ruleIpKey := keys.GetRedisWhiteKeyByRule(rule)
	if err := a.redisClient.SRem(ruleIpKey, ips); err != nil {
		log.Logger.Warnf("failed to remove ips=%v, rule=%s", ips, rule)
		return err
	}

	return nil
}
