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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	LoadBalancerWidth = 16

	LabelNodeRolePrefix = "node-role.kubernetes.io/"

	NodeLabelRole = "kubernetes.io/role"
)

type DescriberFunc func(restClientGetter genericclioptions.RESTClientGetter, mapping *meta.RESTMapping) (Describer, error)

type Describer interface {
	Describe(namespace, name string, describerSettings DescriberSettings) (output string, err error)
}

type DescriberSettings struct {
	ShowEvents bool
}

type ObjectDescriber interface {
	DescribeObject(object interface{}, extra ...interface{}) (output string, err error)
}

type ErrNoDescriber struct {
	Types []string
}

func (e ErrNoDescriber) Error() string {
	return fmt.Sprintf("no describer has been defined for %v", e.Types)
}
