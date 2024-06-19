// Copyright 2021 Northern.tech AS
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

func TestStatusUnmarshal(t *testing.T) {
	var report StatusReport

	err := json.Unmarshal([]byte(`{"status": "aborted"}`), &report)
	assert.EqualError(t, err, "status: must be a valid value.")

	err = json.Unmarshal([]byte(`"status": "bad"}`), &report)
	assert.Error(t, err)

	err = json.Unmarshal([]byte(`{"status": "installing"}`), &report)
	assert.NoError(t, err)
	assert.Equal(t,
		StatusReport{Status: DeviceDeploymentStatusInstalling},
		report)
}
