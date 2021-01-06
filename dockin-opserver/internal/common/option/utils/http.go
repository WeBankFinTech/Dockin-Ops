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

package utils

import (
	"fmt"
	"net/http"

	"github.com/webankfintech/dockin-opserver/internal/log"

	"github.com/valyala/fasthttp"
)

func HttpPost(url, payload string) ([]byte, error) {
	req := &fasthttp.Request{}
	req.SetRequestURI(url)
	req.SetBody([]byte(payload))

		req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")

	resp := &fasthttp.Response{}
	client := &fasthttp.Client{}
	if err := client.Do(req, resp); err != nil {
		fmt.Println("send post failed", err.Error())
		return nil, err
	}

	return resp.Body(), nil
}

func HttpGetFile(url string) (*http.Response, error) {
	log.Logger.Infof("client2server url:%s", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return res, nil
}
