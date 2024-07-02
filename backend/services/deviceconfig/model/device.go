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
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Device struct {
	// ID is the device id assigned by deviceauth
	ID string `bson:"_id" json:"id"`

	// ConfiguredAttributes are the configured attributes for the device.
	ConfiguredAttributes Attributes `bson:"configured,omitempty" json:"configured"`
	// ReportedAttributes are the configuration reported by the device.
	ReportedAttributes Attributes `bson:"reported,omitempty" json:"reported"`
	// DeploymentID is the ID of the latest configuration deployment
	DeploymentID *uuid.UUID `bson:"deployment_id,omitempty" json:"deployment_id,omitempty"`

	// UpdatedTS holds the timestamp for when the desired state changed,
	// including when the object was created.
	UpdatedTS *time.Time `bson:"updated_ts" json:"updated_ts"`
	// ReportTS holds the timestamp when the device last reported its' state.
	ReportTS *time.Time `bson:"reported_ts,omitempty" json:"reported_ts,omitempty"`
}

func (dev Device) Validate() error {
	err := validation.ValidateStruct(&dev,
		validation.Field(&dev.ID, validation.Required),
		validation.Field(&dev.ConfiguredAttributes),
		validation.Field(&dev.ReportedAttributes),
	)
	return errors.Wrap(err, "invalid device object")
}

type NewDevice struct {
	ID string `json:"device_id"`
}

func (dev NewDevice) Validate() error {
	return validation.ValidateStruct(&dev,
		validation.Field(&dev.ID, validation.Required),
	)
}
