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

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const KubeConfigEnvName = "KUBECONFIG"
const AppHome = "APP_HOME"
const EmptyString = ""

var appHome string

const AppHomeEnv = "APP_HOME"

func init() {
	appHome = os.Getenv(AppHomeEnv)
	if appHome == EmptyString {
		curDir := CurrentDir()
		index := strings.LastIndex(curDir, "dockin-opserver")
		if index == -1 {
			panic(fmt.Errorf("try to look src in current path, but failed"))
		}
		appHome = filepath.Join(curDir[0:index], "dockin-opserver")
	}
}

func GetAppHome() string {
	return appHome
}

func GetConfPath() string {
	return filepath.Join(appHome, "configs")
}

func CurrentDir() string {
	_, dir, _, ok := runtime.Caller(1)
	if !ok {
		panic(fmt.Errorf("can not get current dir"))
	}
	return dir
}
