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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/atomic"
	"k8s.io/apimachinery/pkg/util/rand"
)

var (
	totalCnt = atomic.NewInt32(0)
	succCnt  = atomic.NewInt32(0)
	_1Cnt    = atomic.NewInt32(0)
	_2Cnt    = atomic.NewInt32(0)
	_3Cnt    = atomic.NewInt32(0)
	_4Cnt    = atomic.NewInt32(0)
)

type param struct {
	Concurrent int      `json:"concurrent"`
	Number     int      `json:"number"`
	PodList    []string `json:"podList"`
	Timeout    int64    `json:"timeout"`
	PoolSize   int      `json:"poolSize"`
}

func (p *param) String() string {
	return fmt.Sprintf("Concurrent=%d\tNumber=%d\tTimeout=%v\tPoolSize=%d\n", p.Concurrent, p.Number, p.Timeout, p.PoolSize)
}

type Result struct {
	Code    int
	Message string
	Data    interface{}
}

func run(podIp string, timeout time.Duration) int {
	url := fmt.Sprintf("http://127.0.0.1:8083/v1/dockin/opserver/ps?podIp=%s&userName=root&rule=dockin-opagent", podIp)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1
	}

	req.Header.Set("Cache-Control", "no-cache")
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return -2
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -3
	}
	r := &Result{}
	err = json.Unmarshal(body, r)
	if err != nil {
		return -4
	}
	return 0
}

func main() {
	pa := &param{}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	content, err := ioutil.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, pa)
	if err != nil {
		panic(err)
	}
	fmt.Printf(pa.String())
	fmt.Printf("start at %v\n", time.Now().String())

	size := len(pa.PodList)
	wg := new(sync.WaitGroup)
	wg.Add(pa.Concurrent)
	for i := 0; i < pa.Concurrent; i++ {
		go func() {
			defer wg.Done()
			p, _ := ants.NewPool(pa.PoolSize)
			defer p.Release()

			w := new(sync.WaitGroup)
			for k := 0; k < pa.Number; k++ {
				w.Add(1)
				p.Submit(func() {
					totalCnt.Add(1)
					podIp := pa.PodList[rand.Intn(size)]
					result := run(podIp, time.Duration(pa.Timeout)*time.Millisecond)
					if result == 0 {
						succCnt.Add(1)
					} else if result == -1 {
						_1Cnt.Add(1)
					} else if result == -2 {
						_2Cnt.Add(1)
					} else if result == -3 {
						_3Cnt.Add(1)
					} else if result == -4 {
						_4Cnt.Add(1)
					}

					w.Done()
				})
			}
			w.Wait()
		}()
	}
	ch := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				fmt.Printf("total=%d\tsuccess=%d\tNewReqErr=%d\tDoReqErr=%d\tReadBodyErr=%d\tUnMarshalBodyErr=%d\n",
					totalCnt.Load(), succCnt.Load(), _1Cnt.Load(), _2Cnt.Load(), _3Cnt.Load(), _4Cnt.Load())

				totalCnt.Store(0)
				succCnt.Store(0)
				_1Cnt.Store(0)
				_2Cnt.Store(0)
				_3Cnt.Store(0)
				_4Cnt.Store(0)
			case <-ch:
				fmt.Println("timer exit")
			}
		}
	}()
	wg.Wait()
	time.Sleep(time.Second * 3)
	ch <- true
	fmt.Println("benchmark end")
}
