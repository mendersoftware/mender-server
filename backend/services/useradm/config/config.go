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
	SettingListen        = "listen"
	SettingListenDefault = ":8080"

	SettingMiddleware        = "middleware"
	SettingMiddlewareDefault = "prod"

	SettingServerPrivKeyPath                   = "server_priv_key_path"
	SettingServerPrivKeyPathDefault            = "/etc/useradm/rsa/private.pem"
	SettingServerPrivKeyFileNamePattern        = "server_priv_key_filename_pattern"
	SettingServerPrivKeyFileNamePatternDefault = "private\\.id\\.([0-9]*)\\.pem"

	SettingServerFallbackPrivKeyPath        = "server_fallback_priv_key_path"
	SettingServerFallbackPrivKeyPathDefault = ""

	SettingJWTIssuer        = "jwt_issuer"
	SettingJWTIssuerDefault = "mender.useradm"

	SettingJWTExpirationTimeout        = "jwt_exp_timeout"
	SettingJWTExpirationTimeoutDefault = "604800" //one week

	SettingDb        = "mongo"
	SettingDbDefault = "mongo-useradm"

	SettingTenantAdmAddr        = "tenantadm_addr"
	SettingTenantAdmAddrDefault = ""

	SettingDbSSL        = "mongo_ssl"
	SettingDbSSLDefault = false

	SettingDbSSLSkipVerify        = "mongo_ssl_skipverify"
	SettingDbSSLSkipVerifyDefault = false

	SettingDbUsername = "mongo_username"
	SettingDbPassword = "mongo_password"

	SettingRedisConnectionString        = "redis_connection_string"
	SettingRedisConnectionStringDefault = ""

	SettingRedisKeyPrefix        = "redis_key_prefix"
	SettingRedisKeyPrefixDefault = "useradm:v1"

	SettingLimitSessionsPerUser        = "limit_sessions_per_user"
	SettingLimitSessionsPerUserDefault = 10

	SettingLimitTokensPerUser        = "limit_tokens_per_user"
	SettingLimitTokensPerUserDefault = 10

	SettingTokenLastUsedUpdateFreqMinutes        = "token_last_used_update_freq_minutes"
	SettingTokenLastUsedUpdateFreqMinutesDefault = 5

	SettingTokenMaxExpirationSeconds        = "token_max_expiration_seconds"
	SettingTokenMaxExpirationSecondsDefault = 31536000

	SettingPlanDefinitions        = "plan_definitions_path"
	SettingPlanDefinitionsDefault = "/etc/useradm/plans.yaml"
)

var (
	ConfigDefaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingMiddleware, Value: SettingMiddlewareDefault},
		{Key: SettingServerPrivKeyPath, Value: SettingServerPrivKeyPathDefault},
		{Key: SettingServerPrivKeyFileNamePattern,
			Value: SettingServerPrivKeyFileNamePatternDefault},
		{Key: SettingServerFallbackPrivKeyPath, Value: SettingServerFallbackPrivKeyPathDefault},
		{Key: SettingJWTIssuer, Value: SettingJWTIssuerDefault},
		{Key: SettingJWTExpirationTimeout, Value: SettingJWTExpirationTimeoutDefault},
		{Key: SettingDb, Value: SettingDbDefault},
		{Key: SettingTenantAdmAddr, Value: SettingTenantAdmAddrDefault},
		{Key: SettingDbSSL, Value: SettingDbSSLDefault},
		{Key: SettingDbSSLSkipVerify, Value: SettingDbSSLSkipVerifyDefault},
		{Key: SettingRedisConnectionString, Value: SettingRedisConnectionStringDefault},
		{Key: SettingRedisKeyPrefix, Value: SettingRedisKeyPrefixDefault},
		{Key: SettingLimitSessionsPerUser, Value: SettingLimitSessionsPerUserDefault},
		{Key: SettingLimitTokensPerUser, Value: SettingLimitTokensPerUserDefault},
		{Key: SettingTokenLastUsedUpdateFreqMinutes,
			Value: SettingTokenLastUsedUpdateFreqMinutesDefault},
		{Key: SettingTokenMaxExpirationSeconds,
			Value: SettingTokenMaxExpirationSecondsDefault},
		{Key: SettingPlanDefinitions,
			Value: SettingPlanDefinitionsDefault},
	}
)
