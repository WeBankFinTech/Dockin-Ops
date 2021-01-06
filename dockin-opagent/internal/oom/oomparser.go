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
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/log"
)

var (
	legacyContainerRegexp = regexp.MustCompile(`Task in (.*) killed as a result of limit of (.*)`)
	// Starting in 5.0 linux kernels, the OOM message changed
	containerRegexp = regexp.MustCompile(`oom-kill:constraint=(.*),nodemask=(.*),cpuset=(.*),mems_allowed=(.*),oom_memcg=(.*) (.*),task_memcg=(.*),task=(.*),pid=(.*),uid=(.*)`)
	lastLineRegexp  = regexp.MustCompile(`Killed process ([0-9]+) \((.+)\)`)
	firstLineRegexp = regexp.MustCompile(`invoked oom-killer:`)
)

type OomParser struct {
	parser Parser
}

type OomInstance struct {
	// process id of the killed process
	Pid int `json:"pid"`
	// the name of the killed process
	ProcessName string `json:"processName"`
	// the time that the process was reported to be killed,
	// accurate to the minute
	TimeOfDeath time.Time `json:"timeOfDeath"`
	// the absolute name of the container that OOMed
	ContainerName string `json:"containerName"`
	// the absolute name of the container that was killed
	// due to the OOM.
	VictimContainerName string `json:"victimContainerName"`
	// the constraint that triggered the OOM.  One of CONSTRAINT_NONE,
	// CONSTRAINT_CPUSET, CONSTRAINT_MEMORY_POLICY, CONSTRAINT_MEMCG
	Constraint string `json:"constraint"`
}

func getLegacyContainerName(line string, currentOomInstance *OomInstance) error {
	parsedLine := legacyContainerRegexp.FindStringSubmatch(line)
	if parsedLine == nil {
		return nil
	}
	currentOomInstance.ContainerName = path.Join("/", parsedLine[1])
	currentOomInstance.VictimContainerName = path.Join("/", parsedLine[2])
	return nil
}

func getContainerName(line string, currentOomInstance *OomInstance) (bool, error) {
	parsedLine := containerRegexp.FindStringSubmatch(line)
	if parsedLine == nil {
		// Fall back to the legacy format if it isn't found here.
		return false, getLegacyContainerName(line, currentOomInstance)
	}
	currentOomInstance.ContainerName = parsedLine[7]
	currentOomInstance.VictimContainerName = parsedLine[5]
	currentOomInstance.Constraint = parsedLine[1]
	pid, err := strconv.Atoi(parsedLine[9])
	if err != nil {
		return false, err
	}
	currentOomInstance.Pid = pid
	currentOomInstance.ProcessName = parsedLine[8]
	return true, nil
}

func getProcessNamePid(line string, currentOomInstance *OomInstance) (bool, error) {
	reList := lastLineRegexp.FindStringSubmatch(line)

	if reList == nil {
		return false, nil
	}

	pid, err := strconv.Atoi(reList[1])
	if err != nil {
		return false, err
	}
	currentOomInstance.Pid = pid
	currentOomInstance.ProcessName = reList[2]
	return true, nil
}

func checkIfStartOfOomMessages(line string) bool {
	potentialOomStart := firstLineRegexp.MatchString(line)
	return potentialOomStart
}

func (p *OomParser) StreamOoms(outStream chan<- *OomInstance) {
	kmsgEntries := p.parser.Parse()
	defer p.parser.Close()

	for msg := range kmsgEntries {
		isOomMessage := checkIfStartOfOomMessages(msg.Message)
		if isOomMessage {
			oomCurrentInstance := &OomInstance{
				ContainerName:       "/",
				VictimContainerName: "/",
				TimeOfDeath:         msg.Timestamp,
			}
			for msg := range kmsgEntries {
				finished, err := getContainerName(msg.Message, oomCurrentInstance)
				if err != nil {
					log.Logger.Errorf("%v", err)
				}
				if !finished {
					finished, err = getProcessNamePid(msg.Message, oomCurrentInstance)
					if err != nil {
						log.Logger.Errorf("%v", err)
					}
				}
				if finished {
					oomCurrentInstance.TimeOfDeath = msg.Timestamp
					break
				}
			}
			outStream <- oomCurrentInstance
		}
	}
	// Should not happen
	log.Logger.Errorf("exiting analyzeLines. OOM events will not be reported.")
}

func Newoomparser(kmsgPath string) (*OomParser, error) {
	parser, err := NewParser(kmsgPath)
	if err != nil {
		return nil, err
	}
	return &OomParser{parser: parser}, nil
}
