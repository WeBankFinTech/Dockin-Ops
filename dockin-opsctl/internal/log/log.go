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

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	_consoleDebug = "off"
	_loggerFile *os.File
	_loggerDir = "/tmp"
	_enabled = false
)

func init() {
	_consoleDebug = os.Getenv("ConsoleDebug")
	if _consoleDebug == "on" {
		EnableLogger()
	}
}

func EnableLogger() {
	if _enabled {
		return
	}
	if err := os.MkdirAll(_loggerDir, 0644); err != nil {
		Output("failed to mkdir %s", _loggerDir)
		return
	}

	loggerFile := filepath.Join(_loggerDir, "terminal.log")
	fp, err := os.OpenFile(loggerFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		Output("failed to open logger file in %s", loggerFile)
		return
	}

	_loggerFile = fp
	_enabled = true
}

func Debugf(format string, a ...interface{}) {
	if _enabled {
		funcName, _, line, _ := runtime.Caller(1)
		fm := fmt.Sprintf("[%s]-%s:%d %s",
			time.Now().Format("2006:01:02 03:04:15"), runtime.FuncForPC(funcName).Name(), line, format)
		fmt.Fprintln(_loggerFile, fmt.Sprintf(fm, a...))
	}
}

func Close() {
	Debugf("close file")
	_loggerFile.Close()
}

func Output(format string, a ...interface{}) {
	data := fmt.Sprintf(format, a...)
	println(data)
}
