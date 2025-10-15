// Copyright 2025 Northern.tech AS
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

package ratelimits

import (
	"fmt"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/mendersoftware/mender-server/pkg/config"
)

var pathRegex = regexp.MustCompile(`\{[^/]+\}`)

type RatelimitConfig struct {
	// RejectUnmatched rejects requests that does not resolve to a
	// ratelimit group. That is, if either there's no APIPattern matching
	// the request or if the GroupExpression does not match a
	// RatelimitGroups.
	// Defaults to false - disable ratelimiting for unmatched requests.
	RejectUnmatched bool
	// RatelimitGroups configures the ratelimiter parameters for a named ratelimit
	// group.
	RatelimitGroups []RatelimitGroupParams `json:"groups"`
	// MatchExpressions configures mathing expressions (API pattern) and mapping
	// them to a group.
	MatchExpressions []MatchGroup `json:"match"`
}

func (rc RatelimitConfig) Validate() error {
	validGroupNames := make([]interface{}, len(rc.RatelimitGroups))
	for i, group := range rc.RatelimitGroups {
		validGroupNames[i] = group.Name
	}

	return validation.ValidateStruct(&rc,
		validation.Field(&rc.MatchExpressions,
			validation.By(validatePatterns),
		),
		validation.Field(&rc.RatelimitGroups,
			validation.By(validateLimitGroups),
		),
	)

}

func validatePatterns(value interface{}) error {
	groups, ok := value.([]MatchGroup)
	if !ok {
		return fmt.Errorf("value is not []MatchGroup")
	}

	seenPatterns := make(map[string]struct{})
	for _, group := range groups {
		normalizedPattern := normalizePattern(group.APIPattern)
		if _, seen := seenPatterns[normalizedPattern]; seen {
			return fmt.Errorf("duplicate API pattern: '%s'", group.APIPattern)
		}
		seenPatterns[normalizedPattern] = struct{}{}
	}
	return nil
}

func validateLimitGroups(value interface{}) error {
	groups, ok := value.([]RatelimitGroupParams)
	if !ok {
		return fmt.Errorf("value is not []RatelimitGroupParams")
	}

	seenGroups := make(map[string]struct{})
	for _, group := range groups {
		if _, seen := seenGroups[group.Name]; seen {
			return fmt.Errorf("duplicate limit group name: '%s'", group.Name)
		}
		seenGroups[group.Name] = struct{}{}
	}
	return nil
}

// Replace all placeholder or wildcard with "*"
func normalizePattern(pattern string) string {
	return pathRegex.ReplaceAllString(pattern, "*")
}

type RatelimitGroupParams struct {
	// Name of the group
	Name string `json:"name"`
	RatelimitParams
}

type RatelimitParams struct {
	// Quota is the number of requests that can be made within Interval
	Quota int64 `json:"quota"`

	// Interval is the time for the rate limit algorithm to reset.
	Interval config.Duration `json:"interval"`

	// EventExpression specifies a Go template for grouping events (requests)
	// when invoking the rate limiter. For example:
	// {{.Identity.Subject}}{{/* Group by JWT subject (user ID) */}}
	// {{.Identity.Tenant}}{{/* Group by tenant ID (shared quota) */}}
	EventExpression string `json:"event_expression"`
}

type MatchGroup struct {
	// APIPattern matches method and path of the incoming request using pattern
	// from Go standard library ServeMux.
	// https://pkg.go.dev/net/http#hdr-Patterns-ServeMux
	APIPattern string `json:"api_pattern"`

	// GroupExpression is a template string for selecting rate limit group.
	GroupExpression string `json:"group_expression,omitempty"`
}
