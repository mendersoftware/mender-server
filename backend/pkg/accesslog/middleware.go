// Copyright 2024 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package accesslog

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/mendersoftware/mender-server/pkg/netutils"
)

const (
	StatusClientClosedConnection = 499

	envProxyDepth = "ACCESSLOG_PROXY_DEPTH"
)

func getClientIPFromEnv() func(r *http.Request) net.IP {
	if proxyDepthEnv, ok := os.LookupEnv(envProxyDepth); ok {
		proxyDepth, err := strconv.ParseUint(proxyDepthEnv, 10, 8)
		if err == nil {
			return func(r *http.Request) net.IP {
				return netutils.GetIPFromXFFDepth(r, int(proxyDepth))
			}
		}
	}
	return nil
}

const MaxTraceback = 32

func collectTrace() string {
	var (
		trace     [MaxTraceback]uintptr
		traceback strings.Builder
	)
	// Skip 4
	// = accesslog.LogFunc
	// + accesslog.collectTrace
	// + runtime.Callers
	// + runtime.gopanic
	n := runtime.Callers(4, trace[:])
	frames := runtime.CallersFrames(trace[:n])
	for frame, more := frames.Next(); frame.PC != 0 &&
		n >= 0; frame, more = frames.Next() {
		funcName := frame.Function
		if funcName == "" {
			fmt.Fprint(&traceback, "???\n")
		} else {
			fmt.Fprintf(&traceback, "%s@%s:%d",
				frame.Function,
				path.Base(frame.File),
				frame.Line,
			)
		}
		if more {
			fmt.Fprintln(&traceback)
		}
		n--
	}
	return traceback.String()
}
