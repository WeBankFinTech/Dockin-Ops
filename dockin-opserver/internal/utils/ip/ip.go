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

package ip

import (
	"net"
	"net/http"

	"github.com/webankfintech/dockin-opserver/internal/log"
)

func GetIp(req *http.Request) string {
	realIp := req.Header.Get("X-Real-IP")
	if realIp != "" {
		log.Logger.Infof("parse remote ip from X-Real-IP=%s", realIp)
		return realIp
	}

	reqIp, _, _ := net.SplitHostPort(req.RemoteAddr)
	log.Logger.Infof("parse remote ip from request.RemoteAddr=%s", reqIp)
	return reqIp
}
