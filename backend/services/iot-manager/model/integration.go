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
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

type Integration struct {
	ID          uuid.UUID   `json:"id" bson:"_id"`
	Provider    Provider    `json:"provider" bson:"provider"`
	Credentials Credentials `json:"credentials" bson:"credentials"`
	Description string      `json:"description,omitempty" bson:"description,omitempty"`
}

var (
	lenLessThan1024 = validation.Length(0, 1024)
)

func (itg Integration) Validate() error {
	return validation.ValidateStruct(&itg,
		validation.Field(&itg.ID),
		validation.Field(&itg.Provider,
			validation.Required,
			validation.By(itg.compatibleCredentials)),
		validation.Field(&itg.Credentials),
		validation.Field(&itg.Description, lenLessThan1024),
	)
}

func (itg Integration) compatibleCredentials(interface{}) error {
	switch itg.Provider {
	case ProviderIoTHub:
		if itg.Credentials.Type == CredentialTypeSAS {
			return nil
		}
	case ProviderIoTCore:
		if itg.Credentials.Type == CredentialTypeAWS {
			return nil
		}
	case ProviderWebhook:
		if itg.Credentials.Type == CredentialTypeHTTP {
			return nil
		}
	}
	return fmt.Errorf(
		"'%s' incompatible with credential type '%s'",
		itg.Provider,
		itg.Credentials.Type,
	)
}

type CredentialType string

const (
	CredentialTypeAWS  CredentialType = "aws"
	CredentialTypeSAS  CredentialType = "sas"
	CredentialTypeHTTP CredentialType = "http"
)

var credentialTypeRule = validation.In(
	CredentialTypeAWS,
	CredentialTypeSAS,
	CredentialTypeHTTP,
)

func (typ CredentialType) Validate() error {
	return credentialTypeRule.Validate(typ)
}

//nolint:lll
type Credentials struct {
	Type CredentialType `json:"type" bson:"type"`

	// AWS Iot Core
	AWSCredentials *AWSCredentials `json:"aws,omitempty" bson:"aws,omitempty"`

	// Azure IoT Hub
	//nolint:lll
	ConnectionString *ConnectionString `json:"connection_string,omitempty" bson:"connection_string,omitempty"`

	// Webhooks
	HTTP *HTTPCredentials `json:"http,omitempty" bson:"http,omitempty"`
}

func (s Credentials) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Type, validation.Required),
		validation.Field(&s.ConnectionString,
			validation.When(s.Type == CredentialTypeSAS, validation.Required)),
		validation.Field(&s.AWSCredentials,
			validation.When(s.Type == CredentialTypeAWS, validation.Required)),
		validation.Field(&s.HTTP,
			validation.When(s.Type == CredentialTypeHTTP, validation.Required)),
	)
}

type IntegrationFilter struct {
	Skip     int64
	Limit    int64
	Provider Provider
	IDs      []uuid.UUID
}
