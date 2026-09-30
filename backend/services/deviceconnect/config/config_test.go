// Copyright 2020 Northern.tech AS
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

package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgconfig "github.com/mendersoftware/mender-server/pkg/config"
)

func TestLoadReadinessDisabled(t *testing.T) {
	v := viper.New()
	v.Set(SettingReadinessSource, "disabled")

	cfg, err := LoadReadiness(v)

	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadReadinessInvalidSource(t *testing.T) {
	testCases := []struct {
		name   string
		source string
		setSrc bool
	}{
		{name: "unset source", setSrc: false},
		{name: "unknown source", source: "bogus", setSrc: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			if tc.setSrc {
				v.Set(SettingReadinessSource, tc.source)
			}

			cfg, err := LoadReadiness(v)

			assert.Nil(t, cfg)
			require.Error(t, err)
			assert.ErrorContains(t, err, "must be one of [disabled, memory]")
		})
	}
}

func TestLoadReadinessMemoryPercentage(t *testing.T) {
	v := viper.New()
	v.Set(SettingReadinessSource, "memory")
	v.Set(SettingReadinessMax, "1000000")
	v.Set(SettingReadinessHigh, "90%")
	v.Set(SettingReadinessLow, "75%")

	cfg, err := LoadReadiness(v)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint64(1000000), cfg.Max)
	assert.Equal(t, uint64(900000), cfg.High)
	assert.Equal(t, uint64(750000), cfg.Low)
	assert.NotNil(t, cfg.Source)
}

func TestLoadReadinessMemoryAbsolute(t *testing.T) {
	v := viper.New()
	v.Set(SettingReadinessSource, "memory")
	v.Set(SettingReadinessMax, "1000000")
	v.Set(SettingReadinessHigh, "800000")
	v.Set(SettingReadinessLow, "600000")

	cfg, err := LoadReadiness(v)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint64(1000000), cfg.Max)
	assert.Equal(t, uint64(800000), cfg.High)
	assert.Equal(t, uint64(600000), cfg.Low)
}

func TestLoadReadinessMemoryMixedPercentAndAbsolute(t *testing.T) {
	v := viper.New()
	v.Set(SettingReadinessSource, "memory")
	v.Set(SettingReadinessMax, "1000000")
	v.Set(SettingReadinessHigh, "80%")
	v.Set(SettingReadinessLow, "200000")

	cfg, err := LoadReadiness(v)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint64(1000000), cfg.Max)
	assert.Equal(t, uint64(800000), cfg.High)
	assert.Equal(t, uint64(200000), cfg.Low)
}

func TestLoadReadinessMemoryDefaultMax(t *testing.T) {
	// readiness.max is left unset, so the max is derived from the
	// process's actual memory limit (cgroup/meminfo) rather than from
	// configuration; watermarks are still computed as a share of it.
	v := viper.New()
	v.Set(SettingReadinessSource, "memory")
	v.Set(SettingReadinessHigh, "50%")
	v.Set(SettingReadinessLow, "25%")

	cfg, err := LoadReadiness(v)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.NotNil(t, cfg.Source)
	assert.Greater(t, cfg.Max, uint64(0))
	assert.Equal(t, cfg.Max/2, cfg.High)
	assert.Equal(t, cfg.Max/4, cfg.Low)
}

func TestLoadReadinessWithPackageDefaults(t *testing.T) {
	// Exercises the shipped Defaults (90%/75% watermarks) the way the
	// service wires them up in production via config.SetDefaults.
	v := viper.New()
	pkgconfig.SetDefaults(v, Defaults)
	v.Set(SettingReadinessSource, "memory")
	v.Set(SettingReadinessMax, "1000000")

	cfg, err := LoadReadiness(v)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, uint64(1000000), cfg.Max)
	assert.Equal(t, uint64(900000), cfg.High)
	assert.Equal(t, uint64(750000), cfg.Low)
}

func TestLoadReadinessInvalidMax(t *testing.T) {
	v := viper.New()
	v.Set(SettingReadinessSource, "memory")
	v.Set(SettingReadinessMax, "not-a-number")

	cfg, err := LoadReadiness(v)

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.ErrorContains(t, err, SettingReadinessMax)
}

func TestLoadReadinessInvalidWatermarks(t *testing.T) {
	testCases := []struct {
		name      string
		high, low string
		errSubstr string
	}{
		{
			name:      "high percent out of range",
			high:      "150%",
			low:       "10%",
			errSubstr: "value out of range",
		},
		{
			name:      "high percent zero",
			high:      "0%",
			low:       "0%",
			errSubstr: "value out of range",
		},
		{
			name:      "high percent malformed",
			high:      "abc%",
			low:       "10%",
			errSubstr: "failed to parse " + SettingReadinessHigh,
		},
		{
			name:      "high absolute malformed",
			high:      "not-a-number",
			low:       "10",
			errSubstr: "invalid configuration " + SettingReadinessHigh,
		},
		{
			name:      "low percent out of range",
			high:      "90%",
			low:       "-1%",
			errSubstr: "value out of range",
		},
		{
			name:      "high below low",
			high:      "60%",
			low:       "80%",
			errSubstr: "cannot be lower than low watermark",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			v.Set(SettingReadinessSource, "memory")
			v.Set(SettingReadinessMax, "1000000")
			v.Set(SettingReadinessHigh, tc.high)
			v.Set(SettingReadinessLow, tc.low)

			cfg, err := LoadReadiness(v)

			assert.Nil(t, cfg)
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.errSubstr)
		})
	}
}
