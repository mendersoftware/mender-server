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

package iothub

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"reflect"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Key []byte

func (k Key) MarshalText() ([]byte, error) {
	n := base64.StdEncoding.EncodedLen(len(k))
	ret := make([]byte, n)
	base64.StdEncoding.Encode(ret, k)
	return ret, nil
}

type SymmetricKey struct {
	Primary   Key `json:"primaryKey"`
	Secondary Key `json:"secondaryKey"`
}

type AuthType string

const (
	AuthTypeSymmetric   AuthType = "sas"
	AuthTypeCertificate AuthType = "certificate"
	AuthTypeNone        AuthType = "none"
	AuthTypeAuthority   AuthType = "Authority"
	AuthTypeSelfSigned  AuthType = "selfSigned"
)

type Auth struct {
	Type          AuthType `json:"type"`
	*SymmetricKey `json:"symmetricKey,omitempty"`
}

func NewSymmetricAuth() (*Auth, error) {
	var primKey, secKey [48]byte
	_, err := io.ReadFull(rand.Reader, primKey[:])
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(rand.Reader, secKey[:])
	if err != nil {
		return nil, err
	}
	return &Auth{
		Type: AuthTypeSymmetric,
		SymmetricKey: &SymmetricKey{
			Primary:   Key(primKey[:]),
			Secondary: Key(secKey[:]),
		},
	}, nil
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

type DeviceCapabilities struct {
	IOTEdge bool `json:"iotEdge"`
}

type TwinProperties struct {
	Desired  map[string]interface{} `json:"desired"`
	Reported map[string]interface{} `json:"reported"`
}

type X509ThumbPrint struct {
	Primary   string `json:"primaryThumbprint"`
	Secondary string `json:"secondaryThumbprint"`
}

type Device struct {
	*Auth                  `json:"authentication,omitempty"`
	*DeviceCapabilities    `json:"capabilities,omitempty"`
	C2DMessageCount        int    `json:"cloudToDeviceMessageCount,omitempty"`
	ConnectionState        string `json:"connectionState,omitempty"`
	ConnectionStateUpdated string `json:"connectionStateUpdatedTime,omitempty"`

	DeviceID         string `json:"deviceId"`
	DeviceScope      string `json:"deviceScope,omitempty"`
	ETag             string `json:"etag,omitempty"`
	GenerationID     string `json:"generationId,omitempty"`
	LastActivityTime string `json:"lastActivityTime,omitempty"`
	Status           Status `json:"status,omitempty"`
	StatusReason     string `json:"statusReason,omitempty"`
	StatusUpdateTime string `json:"statusUpdateTime,omitempty"`
}

func mergeDevices(devices ...*Device) *Device {
	var device *Device
	for _, device = range devices {
		if device != nil {
			break
		}
	}
	if device == nil {
		return new(Device)
	}
	rDevice := reflect.ValueOf(device).Elem()
	for _, dev := range devices {
		if dev == nil {
			continue
		}
		rDev := reflect.ValueOf(*dev)
		for i := 0; i < rDev.NumField(); i++ {
			fDev := rDev.Field(i)
			if fDev.IsZero() {
				continue
			}
			fDevice := rDevice.Field(i)
			fDevice.Set(fDev)
		}
	}
	return device
}

type DeviceTwin struct {
	AuthenticationType string              `json:"authenticationType,omitempty"`
	Capabilities       *DeviceCapabilities `json:"capabilities,omitempty"`

	CloudToDeviceMessageCount int64 `json:"cloudToDeviceMessageCount,omitempty"`

	ConnectionState  string                 `json:"connectionState,omitempty"`
	DeviceEtag       string                 `json:"deviceEtag,omitempty"`
	DeviceID         string                 `json:"deviceId,omitempty"`
	DeviceScope      string                 `json:"deviceScope,omitempty"`
	ETag             string                 `json:"etag,omitempty"`
	LastActivityTime string                 `json:"lastActivityTime,omitempty"`
	ModuleID         string                 `json:"moduleId,omitempty"`
	Properties       TwinProperties         `json:"properties,omitempty"`
	Status           Status                 `json:"status,omitempty"`
	StatusReason     string                 `json:"statusReason,omitempty"`
	StatusUpdateTime string                 `json:"statusUpdateTime,omitempty"`
	Tags             map[string]interface{} `json:"tags,omitempty"`
	Version          int32                  `json:"version,omitempty"`
	X509ThumbPrint   X509ThumbPrint         `json:"x509Thumbprint,omitempty"`
}

type UpdateProperties struct {
	Desired map[string]interface{} `json:"desired"`
}

type DeviceTwinUpdate struct {
	Properties UpdateProperties       `json:"properties,omitempty"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
	ETag       string                 `json:"-"`
	Replace    bool                   `json:"-"`
}
