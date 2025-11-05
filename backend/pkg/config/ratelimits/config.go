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
	"time"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rate"
	"github.com/mendersoftware/mender-server/pkg/redis"
)

type ConfigDisabledError struct {
	Path string
}

func (err *ConfigDisabledError) Error() string {
	return `configuration "` + err.Path + `" disabled`
}

const (
	SettingRatelimits                    = "ratelimits"
	SettingRatelimitsAuth                = SettingRatelimits + ".auth"
	SettingRatelimitsAuthEnable          = SettingRatelimitsAuth + ".enable"
	SettingRatelimitsAuthGroups          = SettingRatelimitsAuth + ".groups"
	SettingRatelimitsAuthMatch           = SettingRatelimitsAuth + ".match"
	SettingRatelimitsAuthRejectUnmatched = SettingRatelimitsAuth + ".reject_unmatched"
)

func LoadRatelimits(c config.Reader) (*RatelimitConfig, error) {
	if !c.GetBool(SettingRatelimitsAuthEnable) {
		return nil, nil
	}
	ratelimitConfig := &RatelimitConfig{
		RejectUnmatched: c.GetBool(SettingRatelimitsAuthRejectUnmatched),
	}
	err := config.UnmarshalSliceSetting(c,
		SettingRatelimitsAuthGroups,
		&ratelimitConfig.RatelimitGroups,
	)
	if err != nil {
		return nil, fmt.Errorf("error loading rate limit groups: %w", err)
	}

	err = config.UnmarshalSliceSetting(c,
		SettingRatelimitsAuthMatch,
		&ratelimitConfig.MatchExpressions,
	)
	if err != nil {
		return nil, fmt.Errorf("error loading rate limit match expressions: %w", err)
	}
	return ratelimitConfig, nil
}

func SetupRedisRateLimits(
	redisClient redis.Client,
	keyPrefix string,
	c config.Reader,
) (*rate.HTTPLimiter, error) {
	if !c.GetBool(SettingRatelimitsAuthEnable) {
		return nil, &ConfigDisabledError{
			Path: SettingRatelimitsAuthEnable,
		}
	}
	lims, err := LoadRatelimits(c)
	if err != nil {
		return nil, err
	}
	err = lims.Validate()
	if err != nil {
		return nil, err
	}
	log.NewEmpty().Debugf("loaded rate limit configuration: %v", lims)
	mux := rate.NewHTTPLimiter()
	if c.GetBool(SettingRatelimitsAuthRejectUnmatched) {
		mux.WithRejectUnmatched()
	}
	for _, group := range lims.RatelimitGroups {
		groupPrefix := fmt.Sprintf("%s:rate:g:%s", keyPrefix, group.Name)
		limiter := redis.NewFixedWindowRateLimiter(
			redisClient, groupPrefix, time.Duration(group.Interval), group.Quota,
		)
		err = mux.AddRateLimitGroup(limiter, group.Name, group.EventExpression)
		if err != nil {
			return nil, fmt.Errorf("error setting up rate limit group %s: %w", group.Name, err)
		}
	}
	for _, expr := range lims.MatchExpressions {
		err = mux.AddMatchExpression(expr.APIPattern, expr.GroupExpression)
		if err != nil {
			return nil, fmt.Errorf("error setting up match patterns: %w", err)
		}
	}
	return mux, nil
}
