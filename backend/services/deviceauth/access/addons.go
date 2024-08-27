// Copyright 2021 Northern.tech AS
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

package access

import (
	"context"
	"regexp"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/addons"
	hdr "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
)

const (
	hdrForwardedMethod = "X-Forwarded-Method"
	hdrForwardedURI    = "X-Forwarded-Uri"

	regexDeviceconfigURI  = "^/api/devices/v[1-9]/deviceconfig"
	regexDevicemonitorURI = "^/api/devices/v[1-9]/devicemonitor"
)

type addonRule struct {
	// URI regex for the restricted resource
	URI *regexp.Regexp
	// Methods to which the rule applies (nil means ALL)
	Methods []string
	// Name gives the name of the addon the feature belongs to
	Name string
}

// the only addon that impose restrictions to the devices API is
// the configure addon
var addonRules = []addonRule{{
	Name: addons.MenderConfigure,
	URI:  regexp.MustCompile(regexDeviceconfigURI),
}, {
	Name: addons.MenderMonitor,
	URI:  regexp.MustCompile(regexDevicemonitorURI),
}}

type addonChecker struct{}

func NewAddonChecker() Checker {
	return new(addonChecker)
}

func (c addonChecker) ValidateWithContext(ctx context.Context) error {
	method := hdr.FromContext(ctx, hdrForwardedMethod)
	URI := hdr.FromContext(ctx, hdrForwardedURI)
	id := identity.FromContext(ctx)
	if id == nil {
		return errors.New("missing tenant context")
	}

	// sidestep addon validation if the device is in trial mode
	if id.Trial {
		return nil
	}

	for _, rule := range addonRules {
		if rule.Methods != nil {
			var applies bool = false
			for _, m := range rule.Methods {
				if m == method {
					applies = true
					break
				}
			}
			if !applies {
				continue
			}
		}
		if !rule.URI.MatchString(URI) {
			continue
		}
		// The rule matches, check if the addon permits it
		var enabled bool = false
		for _, addon := range id.Addons {
			if addon.Name == rule.Name {
				if addon.Enabled {
					enabled = true
				}
				break
			}
		}
		if !enabled {
			return PermissionError{
				error: errors.Errorf(
					"operation requires addon: %s",
					rule.Name,
				),
			}
		}
	}
	return nil
}
