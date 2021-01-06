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

package keys

import (
	"fmt"
)

const (
	_subsystem  = "5582"
	emptyString = ""

	versionKey        = "version"
	ruleRedisKey      = "rule_access"
	accountKey        = "user"
	rawCmdRedisKey    = "raw_cmd"
	commonCmdRedisKey = "common_cmd"
)

func PodWideAllNamespaceSlotKey(clusterID, rule string) string {
	return fmt.Sprintf("%s:p_a_%s_%s_w", _subsystem, clusterID, rule)
}

func PodWideNamespaceSlotKey(clusterID, rule, namespace string) string {
	return fmt.Sprintf("%s:p_%s_%s_%s_w", _subsystem, namespace, clusterID, rule)
}

func PodWideKey(podName string) string {
	return fmt.Sprintf("%s:p_%s_w", _subsystem, podName)
}

func PodYAMLAllNamespaceSlotKey(clusterID, rule string) string {
	return fmt.Sprintf("%s:p_a_%s_%s_y", _subsystem, clusterID, rule)
}

func PodYAMLNamespaceSlotKey(clusterID, rule, namespace string) string {
	return fmt.Sprintf("%s:p_%s_%s_%s_y", _subsystem, namespace, clusterID, rule)
}

func PodYAMLKey(podName string) string {
	return fmt.Sprintf("%s:p_%s_y", _subsystem, podName)
}

func PodUUIDKey(uid string) string {
	return fmt.Sprintf("%s:p_u_%s", _subsystem, uid)
}

func NodeYamlKey(nodeName string) string {
	return fmt.Sprintf("%s:n_%s_y", _subsystem, nodeName)
}

func NodeYAMLSlotKey(clusterID, rule string) string {
	return fmt.Sprintf("%s:n_%s_%s_y", _subsystem, clusterID, rule)
}

func NodeWideSlotKey(clusterID, rule string) string {
	return fmt.Sprintf("%s:n_%s_%s_w", _subsystem, clusterID, rule)
}

func NodeWideKey(nodeName string) string {
	return fmt.Sprintf("%s:n_%s_w", _subsystem, nodeName)
}

func NodeKey(nodeName string) string {
	return fmt.Sprintf("%s:n_%s_y", _subsystem, nodeName)
}

func AccessTokenKey(userName string) string {
	return fmt.Sprintf("%s:token_%s", _subsystem, userName)
}

func GetRedisWhiteKeyByRule(rule string) string {
	return fmt.Sprintf("%s_whitelist", rule)

}

func GetRuleRedisKey() string {
	return ruleRedisKey
}

func GetVersionKey() string {
	return versionKey
}

func GetAccountKey() string {
	return accountKey
}

func GetUserNameKey(userName string) string {
	return fmt.Sprintf("%s_user", userName)

}

func GetRawCmdRedisKey() string {
	return rawCmdRedisKey
}

func GetCommonCmdRedisKey() string {
	return commonCmdRedisKey
}
