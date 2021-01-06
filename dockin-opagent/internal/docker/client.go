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

package docker

import (
	"context"
	"github.com/webankfintech/dockin-opagent/internal/server/streaming"
	"k8s.io/apimachinery/pkg/util/runtime"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/config"
	"github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/docker/shim/libdocker"
	"github.com/webankfintech/dockin-opagent/internal/log"
	dockertypes "github.com/docker/docker/api/types"
)

type Client struct {
	dockerInterface libdocker.Interface
}

var _ streaming.Runtime = &Client{}

func NewClient() (*Client, error) {
	endPoint := config.AgentConf.App.Docker.Sock
	dockerInterface, err := libdocker.NewDockerClientFromConfig(endPoint, time.Minute * 15)
	if err != nil {
		return nil, err
	}
	return &Client{
		dockerInterface: dockerInterface,
	}, nil
}

func (c *Client)PullImage(image string, auth dockertypes.AuthConfig, opts dockertypes.ImagePullOptions) error {
	return c.dockerInterface.PullImage(image,auth,opts)
}



func (c *Client)RunContainer(opt dockertypes.ContainerCreateConfig) (string,error) {
	createdBody, err := c.dockerInterface.CreateContainer(opt)
	if err != nil {
		return "", err
	}
	if err := c.dockerInterface.StartContainer(createdBody.ID); err != nil {
		return "", err
	}
	return createdBody.ID, nil
}

func (c *Client) Exec(ctx context.Context, traceId, containerId, podName string, execCfg dockertypes.ExecConfig, iostream *dockershim.IOStreams, resize <-chan dockershim.TerminalSize) (*dockertypes.ContainerExecInspect, error) {
	native := dockershim.NativeExecHandler{}
	insp, err := native.ExecInContainer(c.dockerInterface, containerId, traceId, podName, execCfg, iostream, resize)
	if err != nil {
		log.Logger.Warnf("exec command err=%v, uid=%s, podName=%s, execConfig=%#v", err, traceId, podName, execCfg)
		return nil,err
	}

	log.Logger.Infof("succ to Exec,traceId=%s",traceId)
	return insp, nil
}

func (c *Client)Attach(containerID string, iostream *dockershim.IOStreams, tty bool,resize <-chan dockershim.TerminalSize,uid string) error {
	log.Logger.Infof("AttachToContainer debug container=%s,traceId=%s", containerID, uid)
	HandleResizing(resize, func(size dockershim.TerminalSize) {
		c.dockerInterface.ResizeContainerTTY(containerID, uint(size.Height), uint(size.Width))
	})

	opts := dockertypes.ContainerAttachOptions{
		Stream: true,
		Stdin:  iostream.In != nil,
		Stdout: iostream.Out != nil,
		Stderr: iostream.ErrOut != nil,
	}

	sopts := libdocker.StreamOptions{
		InputStream:  iostream.In,
		OutputStream: iostream.Out,
		ErrorStream:  iostream.ErrOut,
		RawTerminal:  tty,
	}

	return c.dockerInterface.AttachToContainer(containerID,opts, sopts)
	return nil
}

func (c *Client) ContainerList() ([]dockertypes.Container, error) {
	containers, err := c.dockerInterface.ListContainers(dockertypes.ContainerListOptions{All:true})
	if err != nil {
		log.Logger.Warnf("failed to list containers, err=%s", err.Error())
		return nil, err
	}
	return containers, nil
}

func (c *Client)CleanContainer(id string)  {
	log.Logger.Infof("CleanContainer id=%s",id)
	c.dockerInterface.CleanContainer(id)
}

func HandleResizing(resize <-chan dockershim.TerminalSize, resizeFunc func(size dockershim.TerminalSize)) {
	if resize == nil {
		return
	}

	go func() {
		defer runtime.HandleCrash()

		for size := range resize {
			if size.Height < 1 || size.Width < 1 {
				continue
			}
			resizeFunc(size)
		}
	}()
}