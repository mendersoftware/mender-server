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
	"bytes"
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	omitted        = "...<omitted>"
	lengthOmission = 3
)

// configuration saves the configuration as a binary blob
// It unmarshals any JSON type into a binary blob. Strings
// are escaped as it's decoded.
// The value marshals into a JSON string.
type configuration []byte

func (c *configuration) UnmarshalJSON(b []byte) error {
	if b == nil {
		return errors.New("error decoding configuration: received nil pointer buffer")
	} else if len(b) >= 2 {
		// Check if JSON value is string
		bs := bytes.Trim(b, " ")
		if bs[0] == '"' && bs[len(b)-1] == '"' {
			var str string
			err := json.Unmarshal(b, &str)
			if err == nil {
				*c = []byte(str)
				return nil
			}
		}
	}
	*c = append((*c)[0:0], b...)
	return nil
}

// ConfigurationDeploymentConstructor represent input data needed for creating new Configuration
// Deployment
type ConfigurationDeploymentConstructor struct {
	// Deployment name, required
	Name string `json:"name"`

	// A string containing a configuration object.
	// The deployments service will use it to generate configuration
	// artifact for the device.
	// The artifact will be generated when the device will ask
	// for an update.
	Configuration configuration `json:"configuration,omitempty"`

	// Retries represents the number of retries in case of deployment failures
	Retries uint `json:"retries,omitempty"`
}

// Validate validates the structure.
func (c ConfigurationDeploymentConstructor) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, lengthIn1To4096),
		validation.Field(&c.Configuration, validation.Required),
	)
}

// NewDeploymentWithID creates new configuration deployment object, sets create data by default.
func NewDeploymentWithID(ID string) (*Deployment, error) {
	now := time.Now()

	return &Deployment{
		Created: &now,
		Id:      ID,
		Stats:   NewDeviceDeploymentStats(),
		Status:  DeploymentStatusPending,
	}, nil
}

// NewConfigurationDeploymentFromConstructor creates new Deployments object based onconfiguration
// deployment constructor data
func NewDeploymentFromConfigurationDeploymentConstructor(
	constructor *ConfigurationDeploymentConstructor,
	deploymentID string,
) (*Deployment, error) {

	deployment, err := NewDeploymentWithID(deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deployment from constructor")
	}

	deployment.DeploymentConstructor = &DeploymentConstructor{
		Name: constructor.Name,
		// save constructor name as artifact name just to make deployment valid;
		// this field will be overwritten by the name of the auto-generated
		// configuration artifact
		ArtifactName: constructor.Name,
	}

	deviceCount := 0
	deployment.DeviceCount = &deviceCount

	return deployment, nil
}

type deploymentConfiguration []byte

func (c deploymentConfiguration) MarshalJSON() ([]byte, error) {
	var configuration map[string]string
	err := json.Unmarshal(c, &configuration)
	if err != nil {
		return json.Marshal(c)
	}
	for key, value := range configuration {
		if len(value) > lengthOmission {
			configuration[key] = value[0:lengthOmission] + omitted
		}
	}
	data, err := json.Marshal(configuration)
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}
