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

package utils

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func GetCommandList(cmd string) ([]string, error) {
	var (
		End uint
	)
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
			if len(callExpr.Args) <= 0 || End >= callExpr.Pos().Col(){
				break
			}

			cmdList = append(cmdList, callExpr.Args[0].Lit())
		//case *syntax.DblQuoted:
		//	dbl := node.(*syntax.DblQuoted)
		//	if len(dbl.Parts) > 0 {
		//		cmdList = append(cmdList, dbl.Parts[0].(*syntax.Lit).Value)
		//	}
			End = callExpr.End().Col()
		case *syntax.DeclClause:
			dcl := node.(*syntax.DeclClause)
			cmdList = append(cmdList, dcl.Variant.Value)
		case *syntax.Redirect:
			rdr := node.(*syntax.Redirect)
			//pos := rdr.Pos().Col()
			end := rdr.End().Col()
			if End >= end{
				break
			}

			End = end
		case *syntax.ForClause:
			fc := node.(*syntax.ForClause)
			end := fc.End().Col()
			if End >= end{
				break
			}

			End = end
		case *syntax.IfClause:
			ic := node.(*syntax.IfClause)
			end := ic.End().Col()
			if End >= end{
				break
			}

			End = end
		case *syntax.WhileClause:
			wc := node.(*syntax.WhileClause)
			end := wc.End().Col()
			if End >= end{
				break
			}

			End = end
		}
		return true
	})

	return cmdList, nil
}

func CommandParser(cmd string,aliasMap map[string]string) ([]string,string, error) {
	var (
		aliasCmd 	string
		End			uint
		flag		bool
	)
	End = 0
	flag = false
	cmdReader := strings.NewReader(cmd)
	prog, err := syntax.NewParser().Parse(cmdReader, "")
	if err != nil {
		return nil,"", err
	}
	var cmdList []string
	syntax.Walk(prog, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.CallExpr:
			callExpr := node.(*syntax.CallExpr)
			if len(callExpr.Args) > 0 {
				cmdList = append(cmdList, callExpr.Args[0].Lit())
			}
			if aliasMap != nil{
				var subCmdList []string
				for _,v := range callExpr.Args{
					_,ok := aliasMap[v.Lit()]
					if ok{
						flag = true
						subCmdList = append(subCmdList,aliasMap[v.Lit()])
					}else {
						subCmdList = append(subCmdList,v.Lit())
					}

				}

				pos := callExpr.Pos().Col()
				end := callExpr.End().Col()
				if End != 0{
					aliasCmd += cmd[End-1:pos-1]
				}
				End = end
				aliasCmd += strings.Join(subCmdList," ")

			}
			//case *syntax.DblQuoted:
			//	dbl := node.(*syntax.DblQuoted)
			//	if len(dbl.Parts) > 0 {
			//		cmdList = append(cmdList, dbl.Parts[0].(*syntax.Lit).Value)
			//	}
		case *syntax.DeclClause:
			dcl := node.(*syntax.DeclClause)
			cmdList = append(cmdList, dcl.Variant.Value)
		}
		return true
	})

	if flag{
		cmd = aliasCmd
	}
	return cmdList,cmd, nil
}
