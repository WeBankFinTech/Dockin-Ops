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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/webankfintech/dockin-opserver/internal/common"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var Logger *logrus.Logger
var CommandLogger *logrus.Logger

type LogConfig struct {
	AppLogPath     string `yaml:"appLogPath"`
	CommandLogPath string `yaml:"commandLogPath"`
	LogLevel       string `yaml:"logLevel"`
}

const (
	debugLevel = "debug"
	infoLevel  = "info"
	fatalLevel = "fatal"
	errorLevel = "error"
	traceLevel = "trace"
	panicLevel = "panic"
)

func init() {

	confFile := filepath.Join(common.GetConfPath(), "log.yaml")
	content, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	opts := &LogConfig{}
	if err = yaml.Unmarshal(content, opts); err != nil {
		panic(err)
	}
	Logger = getLogger(opts.LogLevel, opts.AppLogPath)
	CommandLogger = getLogger(opts.LogLevel, opts.CommandLogPath)

}

func getLogger(level, path string) *logrus.Logger {
	logger := logrus.New()
	l := getLogLevel(level)
	logger.SetLevel(l)
	writer := getLogWriter(path)
	logger.SetOutput(writer)

	return logger
}

func getLogWriter(path string) io.Writer {
	writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("Redirect log writer to %s failed, using default: stdout. Error: %e", path, err)
		return os.Stdout
	}
	return writer
}

func getLogLevel(l string) logrus.Level {
	var level logrus.Level
	switch l {
	case debugLevel:
		level = logrus.DebugLevel
	case infoLevel:
		level = logrus.InfoLevel
	case fatalLevel:
		level = logrus.FatalLevel
	case errorLevel:
		level = logrus.ErrorLevel
	case traceLevel:
		level = logrus.TraceLevel
	case panicLevel:
		level = logrus.PanicLevel
	default:
		level = logrus.InfoLevel
	}
	return level
}
