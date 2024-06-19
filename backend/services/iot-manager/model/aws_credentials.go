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
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// nolint:lll
type AWSCredentials struct {
	AccessKeyID      *string        `json:"access_key_id,omitempty" bson:"access_key_id,omitempty"`
	SecretAccessKey  *crypto.String `json:"secret_access_key,omitempty" bson:"secret_access_key,omitempty"`
	Region           *string        `json:"region,omitempty" bson:"region,omitempty"`
	DevicePolicyName *string        `json:"device_policy_name,omitempty" bson:"device_policydevice_policy_name,omitempty"`
}

func (c AWSCredentials) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.AccessKeyID, validation.Required),
		validation.Field(&c.SecretAccessKey, validation.Required),
		validation.Field(&c.Region, validation.Required),
		validation.Field(&c.DevicePolicyName, validation.Required),
	)
}
