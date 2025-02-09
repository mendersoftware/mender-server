// Copyright 2024 Northern.tech AS
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
	"net"

	"github.com/mendersoftware/mender-server/pkg/config"
)

const (
	// SettingListen is the config key for the listen address
	SettingListen = "listen"
	// SettingListenDefault is the default value for the listen address
	SettingListenDefault = ":8080"

	// SettingMongo is the config key for the mongo URL
	SettingMongo = "mongo_url"
	// SettingMongoDefault is the default value for the mongo URL
	SettingMongoDefault = "mongodb://mender-mongo:27017"

	// SettingDbName is the config key for the mongo database name
	SettingDbName = "mongo_dbname"
	// SettingDbNameDefault is the default value for the mongo database name
	SettingDbNameDefault = "iot-manager"

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

	// SettingWorkflowsURL configures the workflows URL
	SettingWorkflowsURL = "workflows_url"
	// SettingWorkflowsURL defines the default workflows URL
	SettingWorkflowsURLDefault = "http://mender-workflows-server:8080"

	// SettingAESEncryptionKey stores the 32 bytes long key used to encrypt and
	// decrypt sensitive data
	SettingAESEncryptionKey = "aes_encryption_key"
	// SettingAESEncryptionFallbackKey stores the 32 bytes long fallback key used
	// to decrypt sensitive data when performing key-rotation
	SettingAESEncryptionFallbackKey = "aes_encryption_fallback_key"

	// SettingDeviceauthURL configures the deviceauth URL
	SettingDeviceauthURL = "deviceauth_url"
	// SettingDeviceauthURL defines the default deviceauth URL
	SettingDeviceauthURLDefault = "http://mender-device-auth:8080"

	// SettingDebugLog is the config key for the turning on the debug log
	SettingDebugLog = "debug_log"
	// SettingDebugLogDefault is the default value for the debug log enabling
	SettingDebugLogDefault = false

	// SettingDomainWhitelist sets the hostnames trusted by this service
	// to mitigate SSRF attacks. The whitelist accept DNS-like wildcards
	// to cover all subdomains.
	SettingDomainWhitelist = "domain_whitelist"
	// SettingDomainWhitelist sets the default to the set of potential IoT
	// Hub domains included in connection strings.
	SettingDomainWhitelistDefault = "*.azure-devices.net *.iot.*.amazonaws.com"

	// SettingEventExpirationTimeout sets the expiration timeout for stored
	// events. After this time events will be removed from the storage.
	SettingEventExpirationTimeout = "event_exp_timeout"
	// SettingEventExpirationTimeoutDefault define the default expiration
	// timeout for events.
	SettingEventExpirationTimeoutDefault = "604800" // one week

	// SettingWebhooksTimeoutSeconds sets the timeout for webook
	// requests. After this number of seconds requests will be cancelled.
	SettingWebhooksTimeoutSeconds = "webhooks_timeout_seconds"
	// SettingWebhooksTimeoutSecondsDefault define the default timeout
	// in seconds for webhook requests.
	SettingWebhooksTimeoutSecondsDefault = "10" // 10 seconds

	SettingWebhooksIPWhitelist = "webhooks_ip_filter_whitelist_cidrs"
	SettingWebhooksIPBlacklist = "webhooks_ip_filter_blacklist_cidrs"
)

var (
	// Defaults are the default configuration settings
	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingMongo, Value: SettingMongoDefault},
		{Key: SettingDbName, Value: SettingDbNameDefault},
		{Key: SettingDbSSL, Value: SettingDbSSLDefault},
		{Key: SettingDbSSLSkipVerify, Value: SettingDbSSLSkipVerifyDefault},
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
		{Key: SettingWorkflowsURL, Value: SettingWorkflowsURLDefault},
		{Key: SettingDeviceauthURL, Value: SettingDeviceauthURLDefault},
		{Key: SettingDomainWhitelist, Value: SettingDomainWhitelistDefault},
		{Key: SettingEventExpirationTimeout, Value: SettingEventExpirationTimeoutDefault},
		{Key: SettingWebhooksTimeoutSeconds, Value: SettingWebhooksTimeoutSecondsDefault},
	}
)

func LoadWebhookCIDRLists() (whitelist, blacklist []*net.IPNet, err error) {
	if !config.Config.IsSet(SettingWebhooksIPBlacklist) &&
		!config.Config.IsSet(SettingWebhooksIPWhitelist) {
		return nil, nil, nil
	}
	blacklistConf := config.Config.GetStringSlice(SettingWebhooksIPBlacklist)
	whitelistConf := config.Config.GetStringSlice(SettingWebhooksIPWhitelist)
	blacklist = make([]*net.IPNet, len(blacklistConf))
	whitelist = make([]*net.IPNet, len(whitelistConf))
	for i, cidr := range blacklistConf {
		_, blacklist[i], err = net.ParseCIDR(cidr)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing IP blacklist: %w", err)
		}
	}
	for i, cidr := range whitelistConf {
		_, whitelist[i], err = net.ParseCIDR(cidr)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing IP whitelist: %w", err)
		}
	}
	return whitelist, blacklist, nil
}
