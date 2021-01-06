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

package ssh

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/webankfintech/dockin-opsctl/internal/common/protocol"
	"github.com/webankfintech/dockin-opsctl/internal/utils"

	"github.com/chzyer/readline"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/webankfintech/dockin-opsctl/internal/common"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/utils/aes"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

func RunBash(proto *protocol.Proto) error {
	proto.Command = "/bin/bash"
	proto.Params["tty"] = true
	width := 116
	height := 22

	fd := int(os.Stdin.Fd())
	width, height, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}
	log.Debugf("width=%d, height=%d", width, height)
	proto.Params["height"] = height
	proto.Params["width"] = width

	uri := url.URL{
		Scheme:   "ws",
		Host:     common.RemoteHost,
		Path:     "/v1/dockin/opserver/ssh-v2",
		RawQuery: createRequestQuery(proto),
	}

	state, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Debugf("terminal.MakeRaw err=%s", err.Error())
		return err
	}
	defer terminal.Restore(fd, state)
	return Exec(proto, uri)
}

func RunInteractive(proto *protocol.Proto) error {
	width := 116
	height := 22

	fd := int(os.Stdin.Fd())
	width, height, err := readline.GetSize(fd)
	if err != nil {
		panic(err)
	}
	log.Debugf("width=%d, height=%d", width, height)
	proto.Params["height"] = height
	proto.Params["width"] = width
	uri := url.URL{
		Scheme:   "ws",
		Host:     common.RemoteHost,
		Path:     "/v1/dockin/opserver/interact-exec",
		RawQuery: createRequestQuery(proto),
	}
	state, err := readline.MakeRaw(fd)
	if err != nil {
		log.Debugf("terminal.MakeRaw err=%s", err.Error())
		return err
	}
	defer readline.Restore(fd, state)
	return Exec(proto, uri)
}

func RunCommon(proto *protocol.Proto) ([]byte, error) {
	data := createRequestQuery(proto)
	reqUrl := common.GetCommonUrlByCmd("common-exec") + "?" + data
	hder := http.Header{}
	return utils.HttpGetWithHeader(reqUrl, time.Second*60*5, hder)
}

func AttachDebug(proto *protocol.Proto) error {
	width := 116
	height := 22

	fd := int(os.Stdin.Fd())
	width, height, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}
	log.Debugf("width=%d, height=%d", width, height)
	proto.Params["height"] = height
	proto.Params["width"] = width
	proto.Params["tty"] = true

	uri := url.URL{
		Scheme:   "ws",
		Host:     common.RemoteHost,
		Path:     "/v1/dockin/opserver/debug",
		RawQuery: createRequestQuery(proto),
	}
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Debugf("terminal.MakeRaw err=%s", err.Error())
		return err
	}

	defer terminal.Restore(fd, state)
	return Exec(proto, uri)
}

func DevOps(proto *protocol.Proto) error {
	width := 116
	height := 22

	fd := int(os.Stdin.Fd())
	width, height, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}

	log.Debugf("width=%d, height=%d", width, height)
	proto.Params["height"] = height
	proto.Params["width"] = width

	uri := url.URL{
		Scheme:   "ws",
		Host:     common.RemoteHost,
		Path:     "/v1/dockin/opserver/devops-exec",
		RawQuery: createRequestQuery(proto),
	}

	state, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Debugf("terminal.MakeRaw err=%s", err.Error())
		return err
	}
	defer terminal.Restore(fd, state)

	return Exec(proto, uri)
}

func Exec(proto *protocol.Proto, uri url.URL) error {
	log.Debugf("connecting to %s", uri.String())

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
		WriteBufferSize:  8192,
	}
	wsconn, _, err := dialer.Dial(uri.String(), nil)
	if err != nil {
		err := fmt.Errorf("make connection remote proxy err:%v", err)
		log.Debugf(err.Error())
		return err
	}
	defer wsconn.Close()

	client := NewClient(wsconn)

	ctx, cancel := context.WithCancel(context.Background())
	go client.WritePump(ctx)
	go client.ReadPump(ctx)
	go func() { UpdateTerminalSize(client.WinSize, client.CloseChan) }()
	go func() {
		buf := bufio.NewReader(os.Stdin)
		for {
			r, n, err := buf.ReadRune()
			if err != nil {
				err := fmt.Errorf("exit read stdin loop as err:%v", err)
				log.Debugf(err.Error())
				return
			}
			log.Debugf("receive input:%v, string:%s", r, string(r))
			if n > 0 {
				client.Send <- r
			}
		}
	}()

	<-client.CloseChan
	cancel()
	log.Debugf("dockin ssh client exit......")

	return nil
}

func createRequestQuery(proto *protocol.Proto) string {
	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		nerr := errors.Errorf("encrypt failed err=%s", err.Error())
		panic(nerr)
	}

	data, _ := jsoniter.MarshalToString(proto)
	ad, err := aes.AesEncrypt(data)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("params=%s", ad)
}
