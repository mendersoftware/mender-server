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
	"strings"
	"sync"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Values for the session status attribute
const (
	SessionStatusDisconnected = "disconnected"
	SessiontatusConnected     = "connected"
)

// Values for the session types attribute
const (
	SessionTypeTerminal    = "terminal"
	SessionTypePortForward = "portforward"
)

func GetSessionSubject(tenantID, sessionID string) string {
	if tenantID == "" {
		return strings.Join([]string{
			"session", sessionID,
		}, ".")
	}
	return strings.Join([]string{
		"session",
		tenantID,
		sessionID,
	}, ".")
}

func GetDeviceSubject(tenantID, deviceID string) string {
	if tenantID == "" {
		return strings.Join([]string{
			"device",
			deviceID,
		}, ".")
	}
	return strings.Join([]string{
		"device",
		tenantID,
		deviceID,
	}, ".")
}

// Session represents a session from a user to a device and its attributes
type Session struct {
	ID                 string      `json:"id" bson:"_id"`
	UserID             string      `json:"user_id" bson:"user_id"`
	DeviceID           string      `json:"device_id" bson:"device_id"`
	Types              []string    `json:"-" bson:"-"`
	StartTS            time.Time   `json:"start_ts" bson:"start_ts"`
	TenantID           string      `json:"tenant_id" bson:"tenant_id"`
	BytesRecordedMutex *sync.Mutex `json:"-" bson:"-"`
	BytesRecorded      int         `json:"bytes_transferred" bson:"bytes_transferred"`
}

func (sess Session) Subject(tenantID string) string {
	return GetSessionSubject(tenantID, sess.ID)
}

func (sess Session) Validate() error {
	return validation.ValidateStruct(&sess,
		validation.Field(&sess.ID, validation.Required),
		validation.Field(&sess.UserID, validation.Required),
		validation.Field(&sess.DeviceID, validation.Required),
		validation.Field(&sess.StartTS, validation.Required),
	)
}

// ActiveSession stores the data about an active session in memory
type ActiveSession struct {
	RemoteTerminal bool
}
