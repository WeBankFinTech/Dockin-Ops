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
	_ "net/http/pprof"

	"github.com/webankfintech/dockin-opagent/internal/api/exec"
	"github.com/webankfintech/dockin-opagent/internal/api/prestop"
	"github.com/webankfintech/dockin-opagent/internal/config"
	"github.com/webankfintech/dockin-opagent/internal/docker"
	"github.com/webankfintech/dockin-opagent/internal/log"

	"go.uber.org/fx"
)

type Server struct {
	PreStopHandler *prestop.PreStopHandler
	DockerHandler  *exec.DockerHandler
}

func NewServer() *Server {
	dc := docker.NewDockerService()
	return &Server{
		PreStopHandler: prestop.NewPreStopHandler(),
		DockerHandler:  exec.NewDockerHandler(dc),
	}
}

func Run(s *Server, life fx.Lifecycle) {
	httpPort := config.AgentConf.App.HTTP.Port
	life.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Logger.Infof("starting HTTP server at %v.", httpPort)
			http.HandleFunc("/dockin/opagent/prestop", s.PreStopHandler.Handle)
			http.HandleFunc("/dockin/opagent/exec/interactive", s.DockerHandler.InteractiveHandle)
			http.HandleFunc("/dockin/opagent/exec/ssh", s.DockerHandler.SSHHandle)
			http.HandleFunc("/dockin/opagent/exec/common", s.DockerHandler.CommonHandle)
			http.HandleFunc("/dockin/opagent/exec/serverexec", s.DockerHandler.ServerExec)
			go http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
			log.Logger.Infof("started HTTP server at %v. success", httpPort)

			return nil
		},

		OnStop: func(ctx context.Context) error {
			log.Logger.Infof("stop HTTP server")
			log.Logger.Infof("stop HTTP server success")
			return nil
		},
	})
}

func main() {
	var (
		debugPort = config.AgentConf.App.Debug.Port
	)
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", debugPort), nil)
	}()

	app := fx.New(fx.Provide(NewServer), fx.Invoke(Run))
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
		// log.Clean()
	}()

	log.Logger.Infof("start proxy success...")
	select {
	case <-app.Done():
		break
	}
}
