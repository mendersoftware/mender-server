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
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentConfigurationMarshalJSON(t *testing.T) {
	data := map[string]string{
		"key":         "value",
		"another-key": "another-value",
	}
	dataJSON, err := json.Marshal(data)
	assert.NoError(t, err)

	var c deploymentConfiguration
	c = deploymentConfiguration(dataJSON)

	dataJSON, err = json.Marshal(c)
	assert.NoError(t, err)
	expected := base64.StdEncoding.EncodeToString([]byte("{\"another-key\":\"ano...\\u003comitted\\u003e\",\"key\":\"val...\\u003comitted\\u003e\"}"))
	assert.Equal(t, "\""+expected+"\"", string(dataJSON))
}

func TestConfigurationDeploymentValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inputConstructor ConfigurationDeploymentConstructor
		outputError      error
	}{
		"ok": {
			inputConstructor: ConfigurationDeploymentConstructor{
				Name:          "foo",
				Configuration: []byte("foo"),
			},
		},
		"ko, missing name": {
			inputConstructor: ConfigurationDeploymentConstructor{
				Configuration: []byte("foo"),
			},
			outputError: errors.New("name: cannot be blank."),
		},
		"ko, missing configuration": {
			inputConstructor: ConfigurationDeploymentConstructor{
				Name: "foo"},
			outputError: errors.New("configuration: cannot be blank."),
		},
		"ko, missing name and configuration": {
			inputConstructor: ConfigurationDeploymentConstructor{},
			outputError:      errors.New("configuration: cannot be blank; name: cannot be blank."),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.inputConstructor.Validate()
			if tc.outputError != nil {
				assert.EqualError(t, err, tc.outputError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestNewDeploymentFromConfigurationDeploymentConstructor(t *testing.T) {

	t.Parallel()

	testCases := map[string]struct {
		inputConstructor  *ConfigurationDeploymentConstructor
		inputDeploymentID string

		outputError error
	}{
		"ok": {
			inputConstructor: &ConfigurationDeploymentConstructor{
				Name:          "foo",
				Configuration: []byte("bar"),
			},
			inputDeploymentID: "baz",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deployment, err := NewDeploymentFromConfigurationDeploymentConstructor(tc.inputConstructor, tc.inputDeploymentID)
			if tc.outputError != nil {
				assert.EqualError(t, err, tc.outputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, deployment.DeploymentConstructor.Name, tc.inputConstructor.Name)
				assert.Equal(t, deployment.Id, tc.inputDeploymentID)
			}
		})
	}

}
