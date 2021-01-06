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

package option

import (
	"fmt"
	"time"

	"github.com/webankfintech/dockin-opserver/internal/common/option/utils"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type LogsOption struct {
	Namespace     string
	ResourceArg   string
	AllContainers bool
	Options       runtime.Object
	Resources     []string

	SinceTime       string
	SinceSeconds    time.Duration
	Follow          bool
	Previous        bool
	Timestamps      bool
	IgnoreLogErrors bool
	LimitBytes      int64
	Tail            int64
	Container       string

	ContainerNameSpecified bool
	Selector               string
	MaxFollowConcurency    int

	Object        runtime.Object
	GetPodTimeout time.Duration

	TailSpecified bool
}

func (option *LogsOption) Complete(flags *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	option.ContainerNameSpecified = cmd.Flag("container").Changed
	option.TailSpecified = cmd.Flag("tail").Changed
	option.Resources = args
	switch len(args) {
	case 0:
		if len(option.Selector) == 0 {
			return utils.UsageErrorf(cmd, "%s", "expected POD or TYPE/NAME is a required argument for the logs command")
		}
	case 1:
		option.ResourceArg = args[0]
		if len(option.Selector) != 0 {
			return utils.UsageErrorf(cmd, "only a selector (-l) or a POD name is allowed")
		}
	case 2:
		option.ResourceArg = args[0]
		option.Container = args[1]
	default:
		return utils.UsageErrorf(cmd, "%s", "expected POD or TYPE/NAME is a required argument for the logs command")
	}

	option.Namespace = *flags.Namespace
	return nil
}

func (o *LogsOption) Validate() error {
	if len(o.SinceTime) > 0 && o.SinceSeconds != 0 {
		return fmt.Errorf("at most one of `sinceTime` or `sinceSeconds` may be specified")
	}
	//
	//logsOptions, ok := o.Options.(*corev1.PodLogOptions)
	//if !ok {
	//	return errors.New("unexpected logs options object")
	//}
	//if o.AllContainers && len(logsOptions.Container) > 0 {
	//	return fmt.Errorf("--all-containers=true should not be specified with container name %s", logsOptions.Container)
	//}
	//
	//if o.ContainerNameSpecified && len(o.Resources) == 2 {
	//	return fmt.Errorf("only one of -c or an inline [CONTAINER] arg is allowed")
	//}
	//
	//if o.LimitBytes < 0 {
	//	return fmt.Errorf("--limit-bytes must be greater than 0")
	//}
	//
	//if logsOptions.SinceSeconds != nil && *logsOptions.SinceSeconds < int64(0) {
	//	return fmt.Errorf("--since must be greater than 0")
	//}
	//
	//if logsOptions.TailLines != nil && *logsOptions.TailLines < -1 {
	//	return fmt.Errorf("--tail must be greater than or equal to -1")
	//}

	return nil
}

func (option *LogsOption) getUrl() string {
	return fmt.Sprintf("http://%s?command=%s", utils.CommonUrl, "logs")
}

func (option *LogsOption) Run() error {
	body, err := jsoniter.MarshalToString(option)
	if err != nil {
		fmt.Printf("json encode failed %v", err)
	}
	resp, err := utils.HttpPost(option.getUrl(), body)
	if err != nil {
		return err
	}
	fmt.Println(string(resp))
	return nil
}
