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
	"github.com/mendersoftware/mender-server/pkg/config"
)

const (
	// SettingListen is the config key for the listen address
	SettingListen = "listen"
	// SettingListenDefault is the default value for the listen address
	SettingListenDefault = ":8080"

	// SettingNatsURI is the config key for the nats uri
	SettingNatsURI = "nats_uri"
	// SettingNatsURIDefault is the default value for the nats uri
	SettingNatsURIDefault = "nats://localhost:4222"

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

	// SettingWorkflowsURL sets the base URL for the workflows orchestrator.
	SettingWorkflowsURL = "workflows_url"
	// SettingWorkflowsURLDefault sets the default workflows URL.
	SettingWorkflowsURLDefault = "http://mender-workflows-server:8080"

	// SettingEnableAuditLogs enables/disables audit logging.
	SettingEnableAuditLogs = "enable_audit"
	// SettingEnableAuditLogsDefault is disabled by default.
	SettingEnableAuditLogsDefault = false

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
)

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
		{Key: SettingWorkflowsURL, Value: SettingWorkflowsURLDefault},
		{Key: SettingEnableAuditLogs, Value: SettingEnableAuditLogsDefault},
		{Key: SettingRecordingExpireSec, Value: SettingRecordingExpireDefault},
		{Key: SettingWSAllowedOrigins, Value: SettingWSAllowedOriginsDefault},
		{Key: SettingGracefulShutdownTimeout, Value: SettingGracefulShutdownTimeoutDefault},
		{Key: SettingMaxRequestSize, Value: SettingMaxRequestSizeDefault},
		{Key: SettingMaxFileUploadSize, Value: SettingMaxFileUploadSizeDefault},
	}
)
