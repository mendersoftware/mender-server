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

package netutils

import (
	"net"
	"net/http"
	"strings"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
)

// GetIPFromXFFDepth gets the IP address at index proxyDepth from the end of the
// list of X-Forwarded-For values. If proxyDepth is 0, the remote address from
// the connection (IP packet) is used.
func GetIPFromXFFDepth(r *http.Request, proxyDepth int) net.IP {
	if proxyDepth == 0 {
		return net.ParseIP(strings.SplitN(r.RemoteAddr, ":", 2)[0])
	}

	xff := r.Header.Values(headerXForwardedFor)
	l := len(xff) - 1
	if l < 0 {
		return nil
	}
	var clientIP net.IP
	for i := len(xff) - 1; i >= 0; i-- {
		ipList := strings.Split(xff[i], ",")
		if numIPs := len(ipList); numIPs >= proxyDepth {
			ipStr := strings.TrimSpace(ipList[numIPs-proxyDepth])
			clientIP = net.ParseIP(ipStr)
			break
		} else {
			proxyDepth -= numIPs
		}
	}
	return clientIP
}
