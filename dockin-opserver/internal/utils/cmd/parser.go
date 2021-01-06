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

package cmd

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func GetCommandList(cmd string) ([]string, error) {
	cmdReader := strings.NewReader(cmd)
	prog, err := syntax.NewParser().Parse(cmdReader, "")
	if err != nil {
		return nil, err
	}
	var cmdList []string
	syntax.Walk(prog, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.CallExpr:
			callExpr := node.(*syntax.CallExpr)
			for _, args := range callExpr.Args {
				cmdList = append(cmdList, args.Lit())
			}
		case *syntax.DblQuoted:
			dbl := node.(*syntax.DblQuoted)
			for _, args := range dbl.Parts {
				cmdList = append(cmdList, args.(*syntax.Lit).Value)
			}
		}
		return true
	})

	return cmdList, nil
}
