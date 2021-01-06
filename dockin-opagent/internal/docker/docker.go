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
	"math/rand"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/webankfintech/dockin-opagent/internal/config"
	dockershim "github.com/webankfintech/dockin-opagent/internal/docker/shim"
	"github.com/webankfintech/dockin-opagent/internal/log"
	"github.com/webankfintech/dockin-opagent/internal/utils/cmap"
)

var (
	PodNameLabel       = "io.kubernetes.pod.name"
	PodContainerLabel  = "io.kubernetes.container.name"
	PodUidLabel        = "io.kubernetes.pod.uid"
	EmptyString        = ""
	tickerTimes  int64 = 15
)

func init() {
	tickerTimes = int64(config.AgentConf.App.Container.Ticker)
}

type DockerService struct {
	clientList    []*Client
	dockerTicker  *time.Ticker
	close         chan error
	Container2Pod cmap.ConcurrentMap
	Uid2Pod       cmap.ConcurrentMap
}

func NewDockerService() *DockerService {
	var clientList []*Client
	for i := 0; i < 10; i++ {
		cli, err := NewClient()
		if err != nil {
			log.Logger.Warnf("create docker client err=%v", err)
			continue
		}
		clientList = append(clientList, cli)
	}

	ds := &DockerService{
		clientList:    clientList,
		dockerTicker:  time.NewTicker(time.Duration(tickerTimes) * time.Second),
		close:         make(chan error),
		Container2Pod: cmap.New(),
		Uid2Pod:       cmap.New(),
	}
	ds.initContainerList()
	go ds.updateContainerTask()
	return ds
}

func (d *DockerService) IsBizContainer(containerId string) bool {
	return d.Container2Pod.Has(containerId)
}

func (d *DockerService) GetClient() *Client {
	cnt := len(d.clientList)
	if cnt == 0 {
		return nil
	}
	return d.clientList[rand.Intn(cnt)]
}

func (d *DockerService) GetDockerIdByPodName(podName string) (containerId string, err error) {
	cid, ok := d.Container2Pod.Get(podName)
	if !ok {
		return EmptyString, errors.Errorf("no containerId exist for podName=%s", podName)
	}
	return cid.(string), nil
}

func (d *DockerService) GetPodNameByPodUid(podUid string) (podName string, err error) {
	name, ok := d.Uid2Pod.Get(podUid)
	if !ok {
		return EmptyString, errors.Errorf("no podName exist for podUid=%s", podUid)
	}
	return name.(string), nil
}

func (d *DockerService) Exec(ctx context.Context, uid, podName string, execCfg dockertypes.ExecConfig, iostream *dockershim.IOStreams, resize chan dockershim.TerminalSize) (*dockertypes.ContainerExecInspect, error) {
	containerId, err := d.GetDockerIdByPodName(podName)
	if err != nil {
		log.Logger.Warnf("can not get containerId by podName=%s, err=%v, uid=%s", podName, err, uid)
		return nil, err
	}

	log.Logger.Infof("execute command for podName=%s,containerId=%s, uid=%s,ExecConfig=%#v", podName,containerId, uid, execCfg)
	insp, err := d.GetClient().Exec(ctx, uid, containerId, podName, execCfg, iostream, resize)
	log.Logger.Infof("execute command for podName=%s, uid=%s, result=%+v", podName, uid, execCfg)

	if err != nil {
		log.Logger.Warnf("execute cmd=%#v in root mod failed, podName=%s, err=%s",
			execCfg, podName, err.Error())
		return nil, err
	}

	return insp, nil
}

func (d *DockerService) initContainerList() {
	log.Logger.Infof("initialize to get container list")
	containerList, err := d.GetClient().ContainerList()
	if err != nil {
		log.Logger.Errorf("get container list failed, err = %s", err.Error())
		return
	}
	d.parseContainerList(containerList)
}

func (d *DockerService) updateContainerTask() {
	log.Logger.Infof("Start to updateContainerTask")
	for {
		select {
		case <-d.dockerTicker.C:
			log.Logger.Debugf("timer to get container list")
			log.Logger.Infof("timer to get container list")
			containerList, err := d.GetClient().ContainerList()
			if err != nil {
				log.Logger.Warnf("get container list failed, err = %s", err.Error())
				continue
			}
			d.parseContainerList(containerList)
		case <-d.close:
			log.Logger.Info("close timer to exit docker service")
			d.dockerTicker.Stop()
		}
	}
}

func (d *DockerService) parseContainerList(cons []dockertypes.Container) {
	for _, c := range cons {
		if strings.Contains(strings.ToLower(c.Image), "pause") {
			continue
		}

		if !strings.Contains(strings.ToLower(c.State),"running"){
			continue
		}

		containerName,ok := c.Labels[PodContainerLabel]
		if !ok || strings.Contains(strings.ToLower(containerName),"init-"){
			log.Logger.Warnf("no containerName label exist or init container for id= %s", c.ID)
			continue
		}

		podName, ok := c.Labels[PodNameLabel]
		if !ok {
			log.Logger.Warnf("no podName label exist for id= %s", c.ID)
			continue
		}
		existId, ok := d.Container2Pod.Get(podName)
		if ok && existId != c.ID {
			log.Logger.Infof("pod name=%s updated, modify the container id from=%s, to %s",
				podName, existId.(string), c.ID)
		}
		log.Logger.Infof("add PodName=%s, ContainerId=%s", podName, c.ID)
		d.Container2Pod.Set(podName, c.ID)

		podUid, ok := c.Labels[PodUidLabel]
		if !ok {
			log.Logger.Warnf("no podUid label exist for id= %s", c.ID)
			continue
		}

		//log.Logger.Infof("add podUid=%s, podName=%s", podUid, podName)
		d.Uid2Pod.Set(podUid, podName)
	}
}

func (d *DockerService) Shutdown() {
	log.Logger.Infof("shutdown docker service")
	d.close <- nil
}
