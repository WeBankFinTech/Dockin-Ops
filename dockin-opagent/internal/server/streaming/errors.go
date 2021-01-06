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

package streaming

import (
	"fmt"
	"net/http"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorStreamingDisabled(method string) error {
	return status.Errorf(codes.NotFound, fmt.Sprintf("streaming method %s disabled", method))
}

func ErrorTooManyInFlight() error {
	return status.Errorf(codes.ResourceExhausted, "maximum number of in-flight requests exceeded")
}

func WriteError(err error, w http.ResponseWriter) error {
	var status int
	switch grpc.Code(err) {
	case codes.NotFound:
		status = http.StatusNotFound
	case codes.ResourceExhausted:
		// We only expect to hit this if there is a DoS, so we just wait the full TTL.
		// If this is ever hit in steady-state operations, consider increasing the MaxInFlight requests,
		// or plumbing through the time to next expiration.
		w.Header().Set("Retry-After", strconv.Itoa(int(CacheTTL.Seconds())))
		status = http.StatusTooManyRequests
	default:
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
	_, writeErr := w.Write([]byte(err.Error()))
	return writeErr
}
