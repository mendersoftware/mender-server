// Copyright 2023 Northern.tech AS
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
	"fmt"
	"strconv"
	"strings"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/services/deviceconnect/utils/memlimit"
)

const (
	// SettingListen is the config key for the listen address
	SettingListen = "listen"
	// SettingListenDefault is the default value for the listen address
	SettingListenDefault = ":8080"

	// SettingNatsURI is the config key for the nats uri
	SettingNatsURI = "nats_uri"
	// SettingNatsURIDefault is the default value for the nats uri
	SettingNatsURIDefault = "nats://mender-nats:4222"

	// SettingMongo is the config key for the mongo URL
	SettingMongo = "mongo_url"
	// SettingMongoDefault is the default value for the mongo URL
	SettingMongoDefault = "mongodb://mender-mongo:27017"

	// SettingDbName is the config key for the mongo database name
	SettingDbName = "mongo_dbname"
	// SettingDbNameDefault is the default value for the mongo database name
	SettingDbNameDefault = "deviceconnect"

	// SettingDbSSL is the config key for the mongo SSL setting
	SettingDbSSL = "mongo_ssl"
	// SettingDbSSLDefault is the default value for the mongo SSL setting
	SettingDbSSLDefault = false

	// SettingDbSSLSkipVerify is the config key for the mongo SSL skip verify setting
	SettingDbSSLSkipVerify = "mongo_ssl_skipverify"
	// SettingDbSSLSkipVerifyDefault is the default value for the mongo SSL skip verify setting
	SettingDbSSLSkipVerifyDefault = false

	// SettingDbUsername is the config key for the mongo username
	SettingDbUsername = "mongo_username"

	// SettingDbPassword is the config key for the mongo password
	SettingDbPassword = "mongo_password"

	// SettingDebugLog is the config key for the turning on the debug log
	SettingDebugLog = "debug_log"
	// SettingDebugLogDefault is the default value for the debug log enabling
	SettingDebugLogDefault = false

	// SettingLogExpireSec is the config key for how long logs will be
	// retained in the database.
	SettingRecordingExpireSec     = "recording_expire_seconds"
	SettingRecordingExpireDefault = 30 * 24 * 60 * 60

	// SettingWSAllowedOrigin configures the allowed origins to use the websocket APIs.
	// An empty list will disable cors checks
	SettingWSAllowedOrigins        = "ws.allowed_origins"
	SettingWSAllowedOriginsDefault = ""

	// SettingGracefulShutdownTimeout is the config key for the
	// graceful shutdown timeout.
	SettingGracefulShutdownTimeout        = "graceful_shutdown_timeout"
	SettingGracefulShutdownTimeoutDefault = "60s"

	// Max Request body size
	SettingMaxRequestSize        = "request_size_limit"
	SettingMaxRequestSizeDefault = 1024 * 1024 // 1 MiB

	// Max Upload size
	SettingMaxFileUploadSize        = "file_upload_limit"
	SettingMaxFileUploadSizeDefault = 1024 * 1024 * 1024 // 1 GiB

	SettingReadinessSource = "readiness.source"
	SettingReadinessMax    = "readiness.max"
	SettingReadinessHigh   = "readiness.high"
	SettingReadinessLow    = "readiness.low"
)

type ReadinessSource interface {
	Usage() (uint64, error)
	String() string
}

type ReadinessLimits struct {
	Source         ReadinessSource
	Max, Low, High uint64
}

func parsePercentOrAbsolute(c config.Reader, key string, percentOf uint64) (uint64, error) {
	var (
		err error
		max uint64
	)
	if !c.IsSet(key) {
		return percentOf, nil
	}
	stringValue := c.GetString(key)
	if p, found := strings.CutSuffix(stringValue, "%"); found {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse %s: %w", key, err)
		}
		if f <= 0.0 || f > 100.0 {
			return 0, fmt.Errorf("invalid configuration %s: value out of range", key)
		}
		max = uint64(float64(percentOf) * f / 100.0)
	} else {
		max, err = strconv.ParseUint(c.GetString(SettingReadinessMax), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid configuration %s: %w", key, err)
		}
	}
	return max, nil
}

func LoadReadiness(c config.Reader) (*ReadinessLimits, error) {
	var (
		max uint64
		err error
		cfg ReadinessLimits
	)
	switch c.GetString(SettingReadinessSource) {
	case "disabled":
		return nil, nil
	case "memory":
		cfg.Max, cfg.Source, err = memlimit.LimitBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to load memory limit: %w", err)
		}
		if c.IsSet(SettingReadinessMax) {
			cfg.Max, err = strconv.ParseUint(c.GetString(SettingReadinessMax), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid configuration %s: %w", SettingReadinessMax, err)
			}
		}
	default:
		return nil, fmt.Errorf("invalid configuration '%s': must be one of [disabled, memory]",
			SettingReadinessSource)
	}
	cfg.High, err = parsePercentOrAbsolute(c, SettingReadinessHigh, max)
	if err != nil {
		return nil, err
	}
	cfg.Low, err = parsePercentOrAbsolute(c, SettingReadinessLow, max)
	if err != nil {
		return nil, err
	}
	if cfg.High < cfg.Low {
		return nil, fmt.Errorf("invalid readiness limits: "+
			"high watermark (%d) cannot be lower than low watermark (%d)",
			cfg.High, cfg.Low)
	}
	return &cfg, err
}

var (
	// Defaults are the default configuration settings
	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingNatsURI, Value: SettingNatsURIDefault},
		{Key: SettingMongo, Value: SettingMongoDefault},
		{Key: SettingDbName, Value: SettingDbNameDefault},
		{Key: SettingDbSSL, Value: SettingDbSSLDefault},
		{Key: SettingDbSSLSkipVerify, Value: SettingDbSSLSkipVerifyDefault},
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
		{Key: SettingRecordingExpireSec, Value: SettingRecordingExpireDefault},
		{Key: SettingWSAllowedOrigins, Value: SettingWSAllowedOriginsDefault},
		{Key: SettingGracefulShutdownTimeout, Value: SettingGracefulShutdownTimeoutDefault},
		{Key: SettingMaxRequestSize, Value: SettingMaxRequestSizeDefault},
		{Key: SettingMaxFileUploadSize, Value: SettingMaxFileUploadSizeDefault},
		{Key: SettingReadinessSource, Value: "disabled"}, // disabled / memory / websockets
		{Key: SettingReadinessMax, Value: nil},           // Target maximum (default from source)
		{Key: SettingReadinessHigh, Value: "90%"},        // High watermark
		{Key: SettingReadinessLow, Value: "75%"},         // Low watermark
	}
)
