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

package store

import (
	"context"
	"errors"
	"io"

	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

// DataStore interface for DataStore services
//
//nolint:lll - skip line length check for interface declaration.
//go:generate ../utils/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error
	ProvisionDevice(ctx context.Context, tenantID string, deviceID string) error
	DeleteDevice(ctx context.Context, tenantID, deviceID string) error
	GetDevice(ctx context.Context, tenantID, deviceID string) (*model.Device, error)
	SetDeviceConnected(ctx context.Context, tenantID, deviceID string) (int64, error)
	SetDeviceDisconnected(ctx context.Context, tenantID, deviceID string, version int64) error
	AllocateSession(ctx context.Context, sess *model.Session) error
	GetSession(ctx context.Context, sessionID string) (*model.Session, error)
	WriteSessionRecords(ctx context.Context, sessionID string, w io.Writer) error
	InsertSessionRecording(ctx context.Context, sessionID string, sessionBytes []byte) error
	InsertControlRecording(ctx context.Context, sessionID string, sessionBytes []byte) error
	DeleteSession(ctx context.Context, sessionID string) (*model.Session, error)
	Close() error
}

var (
	ErrSessionNotFound = errors.New("store: session not found")
)
