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
// +build linux
package ssh

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/webankfintech/dockin-opsctl/internal/log"

	"golang.org/x/crypto/ssh/terminal"
)

func UpdateTerminalSize(winSize chan WindowSize, exitChan chan byte) {
	go func() {
		// SIGWINCH is sent to the process when the window size of the terminal has
		// changed.
		sigwinchCh := make(chan os.Signal, 1)
		signal.Notify(sigwinchCh, syscall.SIGWINCH)

		fd := int(os.Stdout.Fd())
		termWidth, termHeight, err := terminal.GetSize(fd)
		if err != nil {
			fmt.Println(err)
		}

		for {
			select {
			// The client updated the size of the local PTY. This change needs to occur
			// on the server side PTY as well.
			case sigwinch := <-sigwinchCh:
				if sigwinch == nil {
					return
				}
				currTermWidth, currTermHeight, err := terminal.GetSize(fd)
				if err != nil {
					log.Debugf("Unable to send window-change reqest: %s.", err)
					continue
				}

				// Terminal size has not changed, don't do anything.
				if currTermHeight == termHeight && currTermWidth == termWidth {
					continue
				}

				log.Debugf("update window size change, from:%dx%d to %dx%d", termWidth, termHeight, currTermWidth, currTermHeight)

				winSize <- WindowSize{
					Width:  currTermWidth,
					Height: currTermHeight,
				}
				termWidth, termHeight = currTermWidth, currTermHeight
			case <-exitChan:
				log.Debugf("exit the window handle")
				return
			}
		}
	}()

}
