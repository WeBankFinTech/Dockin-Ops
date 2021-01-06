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

package http

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/log"
)

func HttpGet(url string, timeout time.Duration) ([]byte, error){
	log.Logger.Infof("send http request url=%s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cache-Control", "no-cache")
	client := &http.Client{Timeout: timeout}
	resp , err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Warnf("get method return err, url=%s, err=%s", url, err.Error())
		return nil, err
	}

	log.Logger.Infof("send http request finished url=%s, response=%s", url, string(body))
	return body, nil
}

