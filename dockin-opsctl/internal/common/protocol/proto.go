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

package protocol

import "fmt"

type Proto struct {
	Command     string                 `json:"command"`
	Resource    string                 `json:"resource"`
	Name        string                 `json:"name"`
	PrintType   string                 `json:"printType"`
	Flags       []string               `json:"flags"`
	Params      map[string]interface{} `json:"params"`
	UserName    string                 `json:"userName"`
	Password    string                 `json:"password"`
	AccessToken string                 `json:"accessToken"`
	Env         []string               `json:"env"`
	User        string                 `json:"user"`
	WorkDir     string                 `json:"workDir"`
	Container   string                 `json:"container"`
}

func NewProto() *Proto {
	proto := &Proto{}
	proto.Params = make(map[string]interface{})
	proto.PrintType = "wide"
	proto.Flags = nil
	return proto
}

func (p *Proto) String() string {
	return fmt.Sprintf("Command:[%s],Resource:[%s], Name[%s], PrintType:[%s],Flags:[%s], Params:[%s]",
		p.Command, p.Resource, p.Name, p.PrintType, p.Flags, p.Params)
}
