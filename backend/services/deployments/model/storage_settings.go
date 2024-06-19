// Copyright 2023 Northern.tech AS
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
	"io"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type StorageType uint32

const (
	StorageTypeS3 StorageType = iota
	StorageTypeAzure
	storageTypeMax

	storageTypeStrS3    = "s3"
	storageTypeStrAzure = "azure"
)

func (typ *StorageType) UnmarshalText(b []byte) error {
	switch {
	case bytes.Equal(b, []byte(storageTypeStrS3)):
		*typ = StorageTypeS3

	case bytes.Equal(b, []byte(storageTypeStrAzure)):
		*typ = StorageTypeAzure
	default:
		return errors.New("storage type invalid")
	}
	return nil
}

func (typ StorageType) MarshalText() ([]byte, error) {
	switch typ {
	case StorageTypeS3:
		return []byte(storageTypeStrS3), nil
	case StorageTypeAzure:
		return []byte(storageTypeStrAzure), nil
	default:
		return nil, errors.New("storage type invalid")
	}
}

type StorageSettings struct {
	// Type is the provider type (azblob/s3) for the given settings
	Type StorageType `json:"type" bson:"type"`
	// Region sets the s3 bucket region (required when StorageType == StorageTypeAWS)
	Region string `json:"region" bson:"region"`
	// Bucket is the name of the bucket (s3) or container (azblob) storing artifacts.
	Bucket string `json:"bucket" bson:"bucket"`

	// Uri contains the (private) URI used to call the storage APIs.
	Uri string `json:"uri" bson:"uri"`
	// ExternalUri contains the public bucket / container URI.
	ExternalUri string `json:"external_uri" bson:"external_uri"`
	// ProxyURI is used for rewriting presigned requests, pointing the
	// requests to the proxy URL instead of the direct URL to s3.
	ProxyURI *string `json:"proxy_uri,omitempty" bson:"proxy_uri,omitempty"`

	// Key contains the key identifier (azblob: account name) used to
	// authenticate with the storage APIs.
	Key string `json:"key,omitempty" bson:"key,omitempty"`
	// Secret holds the secret part of the authentication credentials.
	Secret string `json:"secret,omitempty" bson:"secret,omitempty"`
	// Token (s3) stores the optional session token.
	Token string `json:"token,omitempty" bson:"token,omitempty"`
	// ConnectionString (azblob) contains the Azure connection string as an
	// alternative set of credentials from (Uri, Key, Secret).
	ConnectionString *string `json:"connection_string,omitempty" bson:"connection_string,omitempty"`
	// ForcePathStyle (s3) enables path-style URL scheme for the s3 API.
	ForcePathStyle bool `json:"force_path_style" bson:"force_path_style"`
	// UseAccelerate (s3) enables AWS transfer acceleration.
	UseAccelerate bool `json:"use_accelerate" bson:"use_accelerate"`
}

func ParseStorageSettingsRequest(source io.Reader) (settings *StorageSettings, err error) {
	// NOTE: by wrapping StorageSettings as an embedded struct field,
	// passing an empty object `{}` will unmarshall as nil.
	type extendedSettingsSchema struct {
		AccountName   *string `json:"account_name"`
		AccountKey    *string `json:"account_key"`
		ContainerName *string `json:"container_name"`
		*StorageSettings
	}
	var s extendedSettingsSchema

	err = json.NewDecoder(source).Decode(&s)
	if err == nil && s.StorageSettings != nil {
		if s.Type == StorageTypeAzure {
			if s.AccountName != nil {
				s.Key = *s.AccountName
			}
			if s.AccountKey != nil {
				s.Secret = *s.AccountKey
			}
			if s.ContainerName != nil {
				s.Bucket = *s.ContainerName
			}
		}
		settings = s.StorageSettings
		settings.UseAccelerate = settings.UseAccelerate && settings.Uri == ""
		err = errors.WithMessage(
			settings.Validate(),
			"invalid settings schema",
		)
	}
	return settings, err
}

var (
	ruleStorageType = validation.Max(storageTypeMax).
			Exclusive().
			Error("storage type invalid")
	ruleLen5_20   = validation.Length(5, 20)
	ruleLen5_50   = validation.Length(5, 50)
	ruleLen5_100  = validation.Length(5, 100)
	ruleLen3_2000 = validation.Length(3, 2000)
)

// Validate checks structure according to valid tags
func (s StorageSettings) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Type, ruleStorageType),
		validation.Field(&s.Region, validation.When(s.Type == StorageTypeS3,
			validation.Required, ruleLen5_20,
		)),
		validation.Field(&s.Bucket, validation.Required, ruleLen5_100),
		validation.Field(&s.Key, validation.When(
			s.Type == StorageTypeS3 || s.ConnectionString == nil,
			validation.Required, ruleLen5_50,
		)),
		validation.Field(&s.Secret, validation.When(
			s.Type == StorageTypeS3 || s.ConnectionString == nil,
			validation.Required, ruleLen5_100,
		)),
		validation.Field(&s.Uri, ruleLen3_2000),
		validation.Field(&s.ExternalUri, ruleLen3_2000),
		validation.Field(&s.Token, ruleLen5_100),
	)
}
