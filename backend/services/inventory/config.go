// Copyright 2023 Northern.tech AS
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
package main

import (
	"github.com/mendersoftware/mender-server/services/inventory/config"
)

const (
	SettingListen        = "listen"
	SettingListenDefault = ":8080"

	SettingDb        = "mongo"
	SettingDbDefault = "mongo-inventory:27017"

	SettingDbSSL        = "mongo_ssl"
	SettingDbSSLDefault = false

	SettingDbSSLSkipVerify        = "mongo_ssl_skipverify"
	SettingDbSSLSkipVerifyDefault = false

	SettingDbUsername = "mongo_username"
	SettingDbPassword = "mongo_password"

	SettingLimitAttributes        = "limit_attributes"
	SettingLimitAttributesDefault = 100

	SettingLimitTags        = "limit_tags"
	SettingLimitTagsDefault = 20

	SettingDevicemonitorAddr        = "devicemonitor_addr"
	SettingDevicemonitorAddrDefault = "http://mender-devicemonitor:8080"

	SettingEnableReporting        = "enable_reporting"
	SettingEnableReportingDefault = false

	SettingOrchestratorAddr        = "orchestrator_addr"
	SettingOrchestratorAddrDefault = "http://mender-workflows-server:8080"
)

var (
	configDefaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingDb, Value: SettingDbDefault},
		{Key: SettingDbSSL, Value: SettingDbSSLDefault},
		{Key: SettingDbSSLSkipVerify, Value: SettingDbSSLSkipVerifyDefault},
		{Key: SettingLimitAttributes, Value: SettingLimitAttributesDefault},
		{Key: SettingLimitTags, Value: SettingLimitTagsDefault},
		{Key: SettingDevicemonitorAddr, Value: SettingDevicemonitorAddrDefault},
		{Key: SettingEnableReporting, Value: SettingEnableReportingDefault},
		{Key: SettingOrchestratorAddr, Value: SettingOrchestratorAddrDefault},
	}
)
