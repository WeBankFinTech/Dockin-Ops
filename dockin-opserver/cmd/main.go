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
	"context"
	"fmt"
	"net/http"

	"github.com/webankfintech/dockin-opserver/internal/api/ctrl"
	"github.com/webankfintech/dockin-opserver/internal/api/echo"
	"github.com/webankfintech/dockin-opserver/internal/api/exec"
	"github.com/webankfintech/dockin-opserver/internal/api/rm"
	"github.com/webankfintech/dockin-opserver/internal/api/ssh"
	"github.com/webankfintech/dockin-opserver/internal/cache/redis"
	"github.com/webankfintech/dockin-opserver/internal/client"
	"github.com/webankfintech/dockin-opserver/internal/config"
	"github.com/webankfintech/dockin-opserver/internal/controller"
	"github.com/webankfintech/dockin-opserver/internal/log"

	"go.uber.org/fx"

	_ "net/http/pprof"
)

type Server struct {
	Life            fx.Lifecycle
	listenStopper   chan struct{}
	EchoHandler     *echo.Echo
	InteractHandler *exec.Interact
	CommonHandler   *exec.Common
	RmHandler       *rm.Rm
	ControlHandler  *ctrl.Control
	NodeController  *controller.NodeController
	SshHandler      *ssh.Ssh
}

func NewServer(life fx.Lifecycle, cf *config.ProxyConfig) *Server {
	rc, err := redis.NewRedisClient()
	if err != nil {
		log.Logger.Panicf("failed to init redis")
	}

	cm := client.NewManager(rc)
	cm.Initialize()
	return &Server{
		Life:            life,
		listenStopper:   cm.ListenStopper,
		EchoHandler:     echo.NewEcho(cm, rc),
		InteractHandler: exec.NewInteract(cm, rc),
		CommonHandler:   exec.NewCommon(cm, rc),
		RmHandler:       rm.NewRM(cm, rc),
		ControlHandler:  ctrl.NewControl(cm, rc),
		NodeController:  controller.NewNodeController(cm, rc),
		SshHandler:      ssh.NewSsh(cm, rc),
	}

}

func Run(s *Server, cf *config.ProxyConfig) {
	port := cf.HttpPort
	s.Life.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Logger.Infof("starting HTTP server at %v.", port)
			go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
			log.Logger.Infof("started HTTP server at %v. success", port)
			return nil
		},

		OnStop: func(ctx context.Context) error {
			close(s.listenStopper)
			s.EchoHandler.RedisClient.Close()
			log.Logger.Infof("stop HTTP server success")
			return nil
		},
	})
}

func main() {
	go func() {
		http.ListenAndServe(":10000", nil)
	}()

	app := fx.New(
		fx.Provide(func() *config.ProxyConfig {
			return config.OpsConfig
		}),
		fx.Provide(NewServer),
		fx.Invoke(Run))
	if app.Err() != nil {
		log.Logger.Panicf("create proxy failed, %s", app.Err().Error())
	}

	app.Start(context.Background())
	if app.Err() != nil {
		log.Logger.Panicf("start proxy failed, %s", app.Err().Error())
	}

	defer func() {
		app.Stop(context.Background())
		if app.Err() != nil {
			log.Logger.Panicf("stop proxy failed, %s", app.Err().Error())
		}
		log.Logger.Infof("stop proxy success...")
	}()

	log.Logger.Infof("start proxy success...")
	select {
	case <-app.Done():
		break
	}
}
