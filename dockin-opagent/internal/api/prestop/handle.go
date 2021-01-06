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

package prestop

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/webankfintech/dockin-opagent/internal/common"
	"github.com/webankfintech/dockin-opagent/internal/log"
	"github.com/webankfintech/dockin-opagent/internal/model"
	"github.com/webankfintech/dockin-opagent/internal/utils"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const Dir = "/data/dockin_prestop_info"

type PreStopHandler struct {
}

func NewPreStopHandler() *PreStopHandler {
	return &PreStopHandler{}
}

type ReqBody struct {
	PodName string `json:"podName"`
	Command string `json:"command"`
}

func (p *PreStopHandler) Handle(writer http.ResponseWriter, req *http.Request) {
	uid := uuid.New().String()
	log.Logger.Infof("receive prestop request, uid=%s", uid)

	defer log.Logger.Infof("end prestop request, uid=%s", uid)

	if err := common.ValidateRequest(req, uid); err != nil {
		res := model.NewErrorAgentResult(err)
		writer.Write(res.ToJSONByte())
		return
	}

	reqBody, err := p.parseParam(req)
	if err != nil {
		log.Logger.Warnf("parse handle preStop request param err=%s,uid=%s", err.Error(), uid)
		nerr := errors.Errorf("parse handle preStop request param error=%s", err.Error())
		res := model.NewErrorAgentResult(nerr)
		writer.Write(res.ToJSONByte())
		return
	}

	podName := reqBody.PodName
	command := reqBody.Command
	log.Logger.Infof("receive prestop request,podName=%s,command=%s uid=%s", podName, command, uid)
	if !utils.Exists(Dir) {
		os.MkdirAll(Dir, os.ModeDir)
	}

	file := Dir + "/" + podName + "-preStop.sh"
	err = p.cleanAndwrite(podName, file, command)
	if err != nil {
		nerr := errors.Errorf("handle prestop request, cleanAndwrite failed error=%s, uid=%s", err.Error(), uid)
		res := model.NewErrorAgentResult(nerr)
		writer.Write(res.ToJSONByte())
		return
	}

	success := model.NewSuccessAgentResult("").ToJSONByte()
	writer.Write(success)
}

func (p *PreStopHandler) cleanAndwrite(podName, file, command string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, command)
	if err != nil {
		return err
	}
	// offset
	//os.Truncate(filename, 0) //clear
	//n, _ := f.Seek(0, os.SEEK_END)
	//_, err = f.WriteAt([]byte(command), n)
	log.Logger.Infof("cleanAndwrite command=%s to file=%s success", command, file)
	return nil
}

func (p *PreStopHandler) parseParam(req *http.Request) (*ReqBody, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Logger.Warnf("parse req body failed")
		return nil, err
	}

	reqBody := &ReqBody{}
	if err = jsoniter.Unmarshal(body, reqBody); err != nil {
		log.Logger.Warnf("unmarshal failed data=%s", string(body))
		return nil, err
	}

	return reqBody, nil
}
