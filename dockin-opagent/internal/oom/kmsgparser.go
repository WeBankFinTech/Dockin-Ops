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
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/webankfintech/dockin-opagent/internal/log"
)

type Parser interface {
	// SeekEnd moves the parser to the end of the kmsg queue.
	SeekEnd() error
	// Parse provides a channel of messages read from the kernel ring buffer.
	// When first called, it will read the existing ringbuffer, after which it will emit new messages as they occur.
	Parse() <-chan Message

	// Close closes the underlying kmsg reader for this parser
	Close() error
}

type Message struct {
	Priority       int
	SequenceNumber int
	Timestamp      time.Time
	Message        string
}

func NewParser(kmsgPath string) (Parser, error) {
	f, err := os.Open(kmsgPath)
	if err != nil {
		return nil, err
	}

	bootTime, err := getBootTime()
	if err != nil {
		return nil, err
	}

	return &parser{
		kmsgReader: f,
		bootTime:   bootTime,
	}, nil
}

type ReadSeekCloser interface {
	io.ReadCloser
	io.Seeker
}

type parser struct {
	kmsgReader ReadSeekCloser
	bootTime   time.Time
}

func getBootTime() (time.Time, error) {
	var sysinfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysinfo)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not get boot time: %v", err)
	}
	// sysinfo only has seconds
	return time.Now().Add(-1 * (time.Duration(sysinfo.Uptime) * time.Second)), nil
}

func (p *parser) Close() error {
	return p.kmsgReader.Close()
}

func (p *parser) SeekEnd() error {
	_, err := p.kmsgReader.Seek(0, os.SEEK_END)
	return err
}

func (p *parser) Parse() <-chan Message {

	output := make(chan Message, 1)

	go func() {
		defer close(output)
		msg := make([]byte, 8192)
		for {
			// Each read call gives us one full message.
			// https://www.kernel.org/doc/Documentation/ABI/testing/dev-kmsg
			n, err := p.kmsgReader.Read(msg)
			if err != nil {
				if err == syscall.EPIPE {
					log.Logger.Warnf("short read from kmsg; skipping")
					continue
				}

				if err == io.EOF {
					log.Logger.Infof("kmsg reader closed, shutting down")
					return
				}

				log.Logger.Errorf("error reading /dev/kmsg: %v", err)
				return
			}

			msgStr := string(msg[:n])

			message, err := p.parseMessage(msgStr)
			if err != nil {
				log.Logger.Warnf("unable to parse kmsg message %q: %v", msgStr, err)
				continue
			}

			output <- message
		}
	}()

	return output
}

func (p *parser) parseMessage(input string) (Message, error) {
	// Format:
	//   PRIORITY,SEQUENCE_NUM,TIMESTAMP,-;MESSAGE
	parts := strings.SplitN(input, ";", 2)
	if len(parts) != 2 {
		return Message{}, fmt.Errorf("invalid kmsg; must contain a ';'")
	}

	metadata, message := parts[0], parts[1]

	metadataParts := strings.Split(metadata, ",")
	if len(metadataParts) < 3 {
		return Message{}, fmt.Errorf("invalid kmsg: must contain at least 3 ',' separated pieces at the start")
	}

	priority, sequence, timestamp := metadataParts[0], metadataParts[1], metadataParts[2]

	prioNum, err := strconv.Atoi(priority)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse %q as priority: %v", priority, err)
	}

	sequenceNum, err := strconv.Atoi(sequence)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse %q as sequence number: %v", priority, err)
	}

	timestampUsFromBoot, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return Message{}, fmt.Errorf("could not parse %q as timestamp: %v", priority, err)
	}
	// timestamp is offset in microsecond from boottime.
	msgTime := p.bootTime.Add(time.Duration(timestampUsFromBoot) * time.Microsecond)

	return Message{
		Priority:       prioNum,
		SequenceNumber: sequenceNum,
		Timestamp:      msgTime,
		Message:        message,
	}, nil
}
