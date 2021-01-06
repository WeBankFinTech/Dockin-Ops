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

package oom

import (
	"github.com/webankfintech/dockin-opagent/internal/log"

	"k8s.io/apimachinery/pkg/util/runtime"
)

type realWatcher struct {
	kmsgPath string
	recorder EventRecorder
}

var _ Watcher = &realWatcher{}

func NewWatcher(recorder EventRecorder, kmsgPath string) Watcher {
	return &realWatcher{
		kmsgPath: kmsgPath,
		recorder: recorder,
	}
}

const systemOOMEvent = "SystemOOM"

func (ow *realWatcher) Start() error {
	oomLog, err := Newoomparser(ow.kmsgPath)
	if err != nil {
		return err
	}
	outStream := make(chan *OomInstance, 10)
	go oomLog.StreamOoms(outStream)

	go func() {
		defer runtime.HandleCrash()

		for event := range outStream {
			if event.ContainerName == "/" {
				log.Logger.Infof("Got sys oom event: %v", event)
				ow.recorder.OnOOMEvent(event)
			}
		}
		log.Logger.Errorf("Unexpectedly stopped receiving OOM notifications")
	}()
	return nil
}
