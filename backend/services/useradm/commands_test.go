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

func TestCommandCreateUser(t *testing.T) {
	conf := &cmocks.Reader{}
	conf.On("GetString", SettingDb).Return("foo")
	conf.On("GetBool", SettingDbSSL).Return(false)
	conf.On("GetBool", SettingDbSSLSkipVerify).Return(false)
	conf.On("GetString", SettingDbUsername).Return("siala")
	conf.On("GetString", SettingDbPassword).Return("haha")

	// not an email, password too short
	err := commandCreateUser(conf, "foo", "bar", "", "")
	assert.Error(t, err)

	if !testing.Short() {
		err = commandCreateUser(conf, "foo@bar.com", "foobarbarbar", "", "")
		assert.Error(t, err)
	}
}
