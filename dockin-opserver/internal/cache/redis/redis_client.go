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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/common"
	"github.com/webankfintech/dockin-opserver/internal/log"

	"github.com/go-redis/redis/v7"
	"gopkg.in/yaml.v2"
)

type RedisClient struct {
	Client *redis.Client
}

type RdsOptions struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

func NewRedisClient() (*RedisClient, error) {
	confFile := filepath.Join(common.GetConfPath(), "redis_origin.yaml")
	content, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Logger.Warnf("failed to read %s", confFile)
		return nil, err
	}

	opts := &RdsOptions{}
	if err = yaml.Unmarshal(content, opts); err != nil {
		log.Logger.Warnf("failed to parse redis config file %s", string(content))
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", opts.Server, opts.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: opts.Password,
		DB:       opts.Db,
	})

	log.Logger.Infof("create weredis proxy pool success")
	return &RedisClient{client}, nil
}

func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	_, err := r.Client.Set(key, value, expiration).Result()
	return err
}

func (r *RedisClient) Get(key string) (interface{}, error) {

	val, err := r.Client.Get(key).Result()
	return val, err
}

func (r *RedisClient) HSet(key, field string, value interface{}) error {
	_, err := r.Client.HSet(key, field, value).Result()
	return err
}

func (r *RedisClient) HGet(key, field string) (interface{}, error) {
	val, err := r.Client.HGet(key, field).Result()
	return val, err
}

func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	val, err := r.Client.HGetAll(key).Result()
	return val, err
}

func (r *RedisClient) HExist(key, field string) (bool, error) {
	val, err := r.Client.HExists(key, field).Result()
	return val, err
}

func (r *RedisClient) HDel(key string, fields ...string) error {
	_, err := r.Client.HDel(key, fields...).Result()
	return err
}

func (r *RedisClient) PipelineSet(data map[string]string, expiration time.Duration) error {
	return nil
}

func (r *RedisClient) PipelineGet(keys []string) ([]string, error) {
	var (
		data []string
	)

	pip := r.Client.Pipeline()
	pipeRun := func(pipKey []string) error {
		for _, key := range keys {
			if _, err := pip.Get(key).Result(); err != nil && err != redis.Nil {
				log.Logger.Warnf("failed to pipeline get key=%s, err=%s", key, err.Error())
				continue
			}
		}
		result, err := pip.Exec()
		if err != nil && err != redis.Nil {
			log.Logger.Warnf("failed to exec pipe, as err=%s", err.Error())
			return err
		}

		for _, r := range result {
			if s, ok := r.(*redis.StringCmd); ok {
				ss, _ := s.Result()
				data = append(data, ss)
			}
		}
		return nil
	}
	max := 50
	num := len(keys) / max
	i := 0
	for i = 0; i < num; i++ {
		pipeRun(keys[i : i*max])
	}
	if i*max < len(keys) {
		pipeRun(keys[i:])
	}

	return data, nil
}

func (r *RedisClient) Del(key string) error {
	_, err := r.Client.Del([]string{key}...).Result()
	return err
}

func (r *RedisClient) SplitRPush(key string, values []string, expiration time.Duration) error {
	sz := len(values)
	if sz < 100 {
		r := r.Client.RPush(key, values)
		if _, err := r.Result(); err != nil {
			log.Logger.Warnf("failed to rpush data to redis, key=%s", key)
			return err
		}
		return nil
	}

	slot := sz / 100
	log.Logger.Infof("values len %d is large than 100, split into %d", sz, slot)
	for i := 0; i < slot; i++ {
		if i+1 == slot {
			r := r.Client.RPush(key, values[i*100:])
			if _, err := r.Result(); err != nil {
				log.Logger.Warnf("failed to rpush data last data to redis, key=%s", key)
			}
		} else {
			r := r.Client.RPush(key, values[i*100:])
			if _, err := r.Result(); err != nil {
				log.Logger.Warnf("failed to rpush data to redis, key=%s", key)
				continue
			}
		}
	}
	if _, err := r.Client.Expire(key, expiration).Result(); err != nil {
		log.Logger.Warnf("set timeout to key = %s failed, err=%s", key, err.Error())
		r.Client.Del(key)
	}
	log.Logger.Infof("success rpush slot value, key=%s", key)
	return nil
}

func (r *RedisClient) SplitLRange(key string) ([]string, error) {
	log.Logger.Infof("split lrange data, key=%s", key)

	var (
		sz   int64
		data []string
		err  error
	)
	if sz, err = r.Client.LLen(key).Result(); err != nil {
		log.Logger.Warnf("failed to get len about list, key=%s", key)
		return nil, err
	}
	log.Logger.Infof("the %s got %d data", key, sz)

	if sz < 100 {
		if data, err = r.Client.LRange(key, 0, -1).Result(); err != nil {
			log.Logger.Warnf("failed to lrange key=%s", key)
			return nil, err
		}
	} else {
		slot := sz / 100
		var i int64 = 0
		for i = 0; i < slot; i++ {
			if i+1 == slot {
				if temp, err := r.Client.LRange(key, i*100, -1).Result(); err == nil {
					data = append(data, temp...)
				}
			} else {
				if temp, err := r.Client.LRange(key, i*100, (i+1)*100).Result(); err == nil {
					data = append(data, temp...)
				}
			}
		}
	}

	log.Logger.Infof("success lrange key=%s", key)
	return data, nil
}

func (r *RedisClient) SAdd(key string, members []string) error {
	if _, err := r.Client.SAdd(key, members).Result(); err != nil {
		log.Logger.Warnf("failed to sadd, key=%s, member=%v", key, members)
		return err
	}
	return nil
}

func (r *RedisClient) SRem(key string, members []string) error {

	if _, err := r.Client.SRem(key, members).Result(); err != nil {
		log.Logger.Warnf("failed to SRem, key=%s, member=%v", key, members)
		return err
	}

	return nil
}

func (r *RedisClient) SMembers(key string) ([]string, error) {

	data, err := r.Client.SMembers(key).Result()
	if err != nil {
		log.Logger.Warnf("failed to SMembers, key=%s", key)
		return nil, err
	}

	return data, nil
}

func (r *RedisClient) Close() {
	r.Client.Close()
}
