// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package netutils

import "net"

type IPFilter interface {
	IsAllowed(ip net.IP) bool
}

func NewIPFilter(blacklist, whitelist []*net.IPNet) IPFilter {
	return &ipFilter{Blacklist: blacklist, Whitelist: whitelist}
}

type ipFilter struct {
	Blacklist []*net.IPNet
	Whitelist []*net.IPNet
}

func (filter *ipFilter) IsAllowed(ip net.IP) bool {
	for _, mask := range filter.Blacklist {
		if mask.Contains(ip) {
			return false
		}
	}
	if len(filter.Whitelist) > 0 {
		for _, mask := range filter.Whitelist {
			if mask.Contains(ip) {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

type ipFilterFunc func(ip net.IP) bool

func (filter ipFilterFunc) IsAllowed(ip net.IP) bool {
	return filter(ip)
}
