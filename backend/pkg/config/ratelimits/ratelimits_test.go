package ratelimits

import (
	"testing"

	"github.com/mendersoftware/mender-server/pkg/config"
)

func TestRatelimitConfigValidate(t *testing.T) {
	validGroup := RatelimitGroupParams{
		Name: "group1",
		RatelimitParams: RatelimitParams{
			Quota:           10,
			Interval:        config.Duration(60),
			EventExpression: "{{.Identity.Subject}}",
		},
	}
	validMatch := MatchGroup{
		APIPattern:      "/api/v1/resource",
		GroupExpression: "group1",
	}

	testCases := map[string]struct {
		cfg RatelimitConfig
		err bool
	}{
		"ok, valid config": {
			cfg: RatelimitConfig{
				RejectUnmatched:  false,
				RatelimitGroups:  []RatelimitGroupParams{validGroup},
				MatchExpressions: []MatchGroup{validMatch},
			},
			err: false,
		},
		"ok, empty config": {
			cfg: RatelimitConfig{},
			err: false,
		},
		"err, duplicate group names": {
			cfg: RatelimitConfig{
				RatelimitGroups: []RatelimitGroupParams{
					validGroup,
					{Name: "group1"},
				},
				MatchExpressions: []MatchGroup{validMatch},
			},
			err: true,
		},
		"err, duplicate API patterns": {
			cfg: RatelimitConfig{
				RatelimitGroups: []RatelimitGroupParams{validGroup},
				MatchExpressions: []MatchGroup{
					validMatch,
					{APIPattern: "/api/v1/resource", GroupExpression: "group1"},
				},
			},
			err: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := tc.cfg.Validate()
			if tc.err && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.err && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}
