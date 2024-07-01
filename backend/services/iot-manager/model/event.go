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
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

type DeliveryStatus struct {
	IntegrationID uuid.UUID `json:"integration_id" bson:"integration_id"`
	Success       bool      `json:"success" bson:"success"`
	Error         string    `json:"error,omitempty" bson:"err,omitempty"`
	StatusCode    *int      `json:"status_code,omitempty" bson:"status,omitempty"`
}

type WebhookEvent struct {
	// ID is a unique UUID for the event.
	ID uuid.UUID `json:"id" bson:"_id"`
	// Type is the type of event.
	Type EventType `json:"type" bson:"type"`
	// Data contains the event payload (depends on type)
	Data interface{} `json:"data" bson:"data"`
	// EventTS is the timestamp when the event has been produced.
	EventTS time.Time `json:"time" bson:"event_ts"`
}

type Event struct {
	// WebhookEvent contains the part of the event exposed to webhook
	// handlers.
	WebhookEvent `bson:"inline"`
	// ExpireTS contains the timestamp when this event entry expires from the
	// database.
	ExpireTS time.Time `json:"-" bson:"expire_ts"`

	DeliveryStatus []DeliveryStatus `json:"delivery_statuses,omitempty" bson:"status,omitempty"`
}

func (event Event) Validate() error {
	return validation.ValidateStruct(&event,
		validation.Field(&event.ID),
		validation.Field(&event.Type, validation.Required),
	)
}

type EventType string

const (
	EventTypeDeviceProvisioned    EventType = "device-provisioned"
	EventTypeDeviceDecommissioned EventType = "device-decommissioned"
	EventTypeDeviceStatusChanged  EventType = "device-status-changed"
)

var eventTypeRule = validation.In(
	EventTypeDeviceProvisioned,
	EventTypeDeviceDecommissioned,
	EventTypeDeviceStatusChanged,
)

func (typ EventType) Validate() error {
	return eventTypeRule.Validate(typ)
}

type EventsFilter struct {
	Skip  int64
	Limit int64
}

// AuthSet contains a subset of the deviceauth AuthSet definition
type AuthSet struct {
	ID       string `json:"id" bson:"id"`
	DeviceID string `json:"device_id" bson:"device_id"`

	IdentityData map[string]interface{} `json:"identity_data" bson:"identity_data"`
	PublicKey    string                 `json:"pubkey" bson:"pubkey"`

	Status    string     `json:"status" bson:"status"`
	CreatedTS *time.Time `json:"ts,omitempty" bson:"ts,omitempty"`
}

// DeviceEvent contains the representation of a device.
// For device decommissioning events, the DeviceEvent only contains the device
// ID. For StatusChangeEvents, the Status is also included. For
// ProvisioningEvents, the entire struct is expected to be populated.
type DeviceEvent struct {
	// ID is the device ID
	ID string `json:"id" bson:"id"`
	// The Status (reported by deviceauth) of the device.
	Status Status `json:"status,omitempty" bson:"status,omitempty"`
	// AuthSets (partial) list of auth sets associated with the device.
	AuthSets []AuthSet `json:"auth_sets,omitempty" bson:"auth_sets,omitempty"`
	// CreatedTS is the time when the device was created.
	CreatedTS *time.Time `json:"created_ts,omitempty" bson:"created_ts,omitempty"`
}
