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

package base

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func IsIp(input string) bool {
	pattern := "[\\d]+\\.[\\d]+\\.[\\d]+\\.[\\d]+"
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}

func IsSubsystem(input string) bool {
	cnt := strings.Count(input, "-")
	return cnt == 1
}

func IsPodName(input string) bool {
	cnt := strings.Count(input, "-")
	return cnt > 2
}

func IsSubSystemId(input string) bool {
	pattern := `^[0-9]\d*$`
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}

func IsUUid(input string) bool {
	if _, err := uuid.Parse(input); err != nil {
		return false
	}
	return true
}

func IsPodSet(input string) bool {
	sps := strings.Split(input, "-")
	cnt := len(sps)
	if cnt == 2 {
		_, err := strconv.ParseInt(sps[0], 10, 32)
		return err == nil
	}
	if cnt == 3 {
		return true
	}
	return false
}