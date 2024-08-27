// Copyright 2020 Northern.tech AS
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
	"testing"

	"github.com/stretchr/testify/assert"

	cmocks "github.com/mendersoftware/mender-server/pkg/config/mocks"

	. "github.com/mendersoftware/mender-server/services/useradm/config"
)

func TestDataStoreMongoConfigFromAppConfig(t *testing.T) {
	appConf := &cmocks.Reader{}
	appConf.On("GetString", SettingDb).Return("192.123.123.123")
	appConf.On("GetBool", SettingDbSSL).Return(true)
	appConf.On("GetBool", SettingDbSSLSkipVerify).Return(false)
	appConf.On("GetString", SettingDbUsername).Return("Steven")
	appConf.On("GetString", SettingDbPassword).Return("Shamballa")

	dbConf := dataStoreMongoConfigFromAppConfig(appConf)
	assert.Equal(t, "192.123.123.123", dbConf.ConnectionString)
	assert.Equal(t, true, dbConf.SSL)
	assert.Equal(t, false, dbConf.SSLSkipVerify)
	assert.Equal(t, "Steven", dbConf.Username)
	assert.Equal(t, "Shamballa", dbConf.Password)
}
