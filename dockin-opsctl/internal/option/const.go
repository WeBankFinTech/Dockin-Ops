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

const (
	NoTypeOrNameErr     = "no resource type or name provided"
	NoResourceNameErr   = "no resource name provided"
	NoCommandErr        = "no command provide"
	LogsCommandErr      = "logs args err"
	CommandParamMissing = "some command param missing"

	GetCommandSuggest    = "See 'dockin-opsctl get -h' for help and examples."
	ListCommandSuggest   = "See 'dockin-opsctl list -h' for help and examples."
	ScriptCommandSuggest = "See 'dockin-opsctl script -h' for help and examples."
	ExecCommandSuggest   = "See 'dockin-opsctl exec -h' for help and examples."
	DevopsCommandSuggest = "See 'dockin-opsctl devops -h' for help and examples."
	JvmCommandSuggest    = "See 'dockin-opsctl jvm -h' for help and examples."
	FileCommandSuggest   = "See 'dockin-opsctl file -h' for help and examples."
	LogsCommandSuggest   = "See 'dockin-opsctl logs -h for help and examples."
	DescCommandSuggest   = "See 'dockin-opsctl descript -h' for help and examples."
	SSHCommandSuggest    = "See 'dockin-opsctl ssh -h' for help and examples."
	DebugCommandSuggest  = "See 'dockin-opsctl debug -h' for help and examples."
)

const (
	UploadFileErr   = "upload file failed!"
	DownloadFileErr = "download file failed!"
	ListErr         = ""
	GetErr          = ""
	ExecErr         = ""
	JvmErr          = ""
	ScriptErr       = ""
)
