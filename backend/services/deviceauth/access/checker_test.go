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
	"net/http"
	"regexp"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/addons"
	hdr "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/plan"
)

func init() {
	addonRules = append(addonRules, addonRule{
		Name:    "test",
		Methods: []string{http.MethodPost, http.MethodPut},
		URI:     regexp.MustCompile("^/api/devices/v1/test"),
	})
}

func TestValidateAddons(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX context.Context

		Error error
	}{{
		Name: "ok",

		CTX: func() context.Context {
			ctx := context.Background()
			ctx = hdr.WithContext(ctx, http.Header{
				hdrForwardedMethod: []string{"GET"},
				hdrForwardedURI: []string{
					"/api/devices/v1/deviceconnect/connect",
				},
			}, hdrForwardedMethod, hdrForwardedURI)
			return identity.WithContext(ctx, &identity.Identity{
				Plan:   plan.PlanEnterprise,
				Addons: addons.AllAddonsEnabled,
			})
		}(),
	}, {
		Name: "ok, addon disabled but trial mode",

		CTX: func() context.Context {
			ctx := context.Background()
			ctx = hdr.WithContext(ctx, http.Header{
				hdrForwardedMethod: []string{"GET"},
				hdrForwardedURI: []string{
					"/api/devices/v1/deviceconfig/configuration",
				},
			}, hdrForwardedMethod, hdrForwardedURI)
			return identity.WithContext(ctx, &identity.Identity{
				Plan:   plan.PlanEnterprise,
				Addons: addons.AllAddonsDisabled,
				Trial:  true,
			})
		}(),
	}, {
		Name: "error, addon disabled",

		CTX: func() context.Context {
			ctx := context.Background()
			ctx = hdr.WithContext(ctx, http.Header{
				hdrForwardedMethod: []string{"GET"},
				hdrForwardedURI: []string{
					"/api/devices/v1/deviceconfig/configuration",
				},
			}, hdrForwardedMethod, hdrForwardedURI)
			return identity.WithContext(ctx, &identity.Identity{
				Plan:   plan.PlanEnterprise,
				Addons: addons.AllAddonsDisabled,
			})
		}(),

		Error: errors.Errorf(
			"operation requires addon: %s",
			addons.MenderConfigure,
		),
	}, {
		Name: "error, addon not present",

		CTX: func() context.Context {
			ctx := context.Background()
			ctx = hdr.WithContext(ctx, http.Header{
				hdrForwardedMethod: []string{http.MethodPut},
				hdrForwardedURI: []string{
					"/api/devices/v1/test/foobar",
				},
			}, hdrForwardedMethod, hdrForwardedURI)
			return identity.WithContext(ctx, &identity.Identity{
				Plan: plan.PlanEnterprise,
			})
		}(),

		Error: errors.New("operation requires addon: test"),
	}, {
		Name: "error, identity not present",

		CTX: func() context.Context {
			ctx := context.Background()
			return hdr.WithContext(ctx, http.Header{
				hdrForwardedMethod: []string{http.MethodPut},
				hdrForwardedURI: []string{
					"/api/devices/v1/test/foobar",
				},
			}, hdrForwardedMethod, hdrForwardedURI)
		}(),

		Error: errors.New("missing tenant context"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			c := Merge(NewAddonChecker())
			err := c.ValidateWithContext(tc.CTX)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
