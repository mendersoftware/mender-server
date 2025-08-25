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
	"net"
	"net/http"
	"os"
	"strconv"

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
