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

package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/log"

	"github.com/valyala/fasthttp"
)

func HttpGet(url string, timeout time.Duration) ([]byte, error) {
	log.Logger.Infof("send http request url=%s", url)

	code, body, err := fasthttp.GetTimeout(nil, url, timeout)
	if err != nil {
		log.Logger.Warnf("get method return err, url=%s, err=%s", url, err.Error())
		return nil, err
	}
	if code != http.StatusOK {
		reqErr := fmt.Errorf("get method return not 200/OK, url=%s, status Code=%s", url, code)
		log.Logger.Warnf(reqErr.Error())
		return nil, err
	}

	log.Logger.Infof("send http request finished url=%s, response=%s", url, string(body))
	return body, nil
}

func HttpPost(url string, payload []byte) ([]byte, error) {
	req := &fasthttp.Request{}
	req.SetRequestURI(url)
	req.SetBody(payload)
	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	resp := &fasthttp.Response{}
	client := &fasthttp.Client{}
	if err := client.Do(req, resp); err != nil {
		return nil, err
	}

	return resp.Body(), nil
}
