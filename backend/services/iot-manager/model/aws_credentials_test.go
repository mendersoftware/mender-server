// Copyright 2022 Northern.tech AS
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

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSCredentialsJSON(t *testing.T) {
	awsCredentials := &AWSCredentials{
		AccessKeyID:      str2ptr("accessKeyID"),
		SecretAccessKey:  str2cyptoptr("secretAccessKey"),
		Region:           str2ptr("c"),
		DevicePolicyName: str2ptr("d"),
	}

	data, err := json.Marshal(awsCredentials)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "accessKeyID")
	assert.NotContains(t, string(data), "secretAccessKey")
}
