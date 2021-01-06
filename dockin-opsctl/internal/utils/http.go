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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/webankfintech/dockin-opsctl/internal/log"
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

func HttpClient() *http.Client {
	nc := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5000,
		},
		Timeout: time.Second * 5,
	}

	return nc
}

func HttpGet(getUrl string, params map[string]string, headers map[string]string, body io.Reader) (*http.Response, error) {
	//new request
	req, err := http.NewRequest("GET", getUrl, body)
	if err != nil {
		return nil, errors.New("new request is fail ")
	}

	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	getUrl, _ = url.QueryUnescape(req.URL.String())
	log.Debugf("client2server url:%s", getUrl)
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{Timeout: time.Second * 60 * 5}
	return client.Do(req)

}

func HttpPostFile(postUrl, paramName, filePath string) (*http.Response, error) {
	log.Debugf("client2server url:%s", postUrl)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", postUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: time.Second * 60 * 5}
	res, err := client.Do(request)

	return res, err
}

func HttpGetFile(url string) (*http.Response, error) {
	log.Debugf("client2server url:%s", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func HttpGetWithTimeout(url string, timeout time.Duration) ([]byte, error) {
	log.Debugf("send http request url=%s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cache-Control", "no-cache")
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("get method return err, url=%s, err=%s", url, err.Error())
		return nil, err
	}

	log.Debugf("send http request finished url=%s, response=%s", url, string(body))
	return body, nil
}

func HttpGetWithHeader(url string, timeout time.Duration, header http.Header) ([]byte, error) {
	log.Debugf("send http request url=%s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	header.Add("Cache-Control", "no-cache")
	req.Header = header
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("get method return err, url=%s, err=%s", url, err.Error())
		return nil, err
	}

	log.Debugf("send http request finished url=%s, response=%s", url, string(body))
	return body, nil
}
