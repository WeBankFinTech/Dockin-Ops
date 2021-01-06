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

package ssh

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func ParseCmdlineToCmdList(cmd string) ([]string, error) {
	var cmdList []string
	cmdReader := strings.NewReader(cmd)
	prog, err := syntax.NewParser().Parse(cmdReader, "")
	if err != nil {
		return nil, err
	}

	syntax.Walk(prog, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.CallExpr:
			callExpr := node.(*syntax.CallExpr)
			if len(callExpr.Args) > 0 {
				cmdList = append(cmdList, callExpr.Args[0].Lit())
			}
		case *syntax.DeclClause:
			dcl := node.(*syntax.DeclClause)
			cmdList = append(cmdList, dcl.Variant.Value)
		}
		return true
	})

	var filterCmd []string
	for _, c := range cmdList {
		idx := strings.LastIndex(c, "/")
		if idx != -1 {
			c = c[idx+1:]
		}
		filterCmd = append(filterCmd, c)
	}

	return filterCmd, nil
}

type CmdUnit struct {
	// target executable binary
	cmd string
	// cmd params
	params []string
	// the target operate resources, such file and directory
	target string
}

func ParseCmdlineToCmdUnitList(cmdline string) ([]*CmdUnit, error) {
	cmdReader := strings.NewReader(cmdline)
	prog, err := syntax.NewParser().Parse(cmdReader, "")
	if err != nil {
		return nil, err
	}

	var cmdList []*CmdUnit
	syntax.Walk(prog, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.CallExpr:
			callExpr := node.(*syntax.CallExpr)
			num := len(callExpr.Args)
			if num == 1 {
				cmdList = append(cmdList, &CmdUnit{
					cmd: callExpr.Args[0].Lit(),
				})
			} else if num == 2 {
				cmdList = append(cmdList, &CmdUnit{
					cmd:    callExpr.Args[0].Lit(),
					target: callExpr.Args[1].Lit(),
				})
			} else if num > 2 {
				cu := &CmdUnit{
					cmd:    callExpr.Args[0].Lit(),
					target: callExpr.Args[num-1].Lit(),
				}
				for _, p := range callExpr.Args[1:] {
					if strings.HasPrefix(p.Lit(), "-") {
						cu.params = append(cu.params, p.Lit())
					} else {
						cu.target = p.Lit()
					}
				}
				cmdList = append(cmdList, cu)
			}
		case *syntax.DeclClause:
			dcl := node.(*syntax.DeclClause)
			cmdList = append(cmdList, &CmdUnit{
				cmd: dcl.Variant.Value,
			})
		}
		return true
	})

	return cmdList, nil
}
