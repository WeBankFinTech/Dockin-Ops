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

package remote

import "context"

type Executor interface {
	Exec(execParam *DockinExecParam, streams *IOStreams) error

	ExecInteractive(execParam *DockinExecParam, cmdStream *InteractStream) error

	Resize(width, height int) error

	Shell(ctx context.Context, execParam *DockinExecParam, interStream *InteractStream) error

	DebugShell(ctx context.Context, execParam *DockinExecParam, interStream *InteractStream) error
}
