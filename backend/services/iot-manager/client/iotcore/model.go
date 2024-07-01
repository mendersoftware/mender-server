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

package iotcore

import (
	"bytes"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

type Device struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Version       int64   `json:"version,omitempty"`
	Status        Status  `json:"status,omitempty"`
	CertificateID string  `json:"certificate_id,omitempty"`
	Certificate   string  `json:"certificate,omitempty"`
	PrivateKey    string  `json:"private_key,omitempty"`
	Endpoint      *string `json:"endpoint,omitempty"`
}

type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

func NewStatusFromMenderStatus(status model.Status) Status {
	switch status {
	case model.StatusAccepted, model.StatusPreauthorized:
		return StatusEnabled
	default:
		return StatusDisabled
	}
}

func (s *Status) UnmarshalText(b []byte) error {
	*s = Status(bytes.ToLower(b))
	return s.Validate()
}

var validateStatus = validation.In(
	StatusEnabled,
	StatusDisabled,
)

func (s Status) Validate() error {
	return validateStatus.Validate(s)
}

type DeviceShadow struct {
	Payload model.DeviceState `json:"state"`
}

type DesiredState struct {
	Desired map[string]interface{} `json:"desired"`
}

type DeviceShadowUpdate struct {
	State DesiredState `json:"state"`
}
