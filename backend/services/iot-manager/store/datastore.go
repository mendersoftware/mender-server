// Copyright 2024 Northern.tech AS
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

	"github.com/google/uuid"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

// DataStore interface for DataStore services
//
//nolint:lll
//go:generate ../../../utils/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error
	Close() error

	GetIntegrations(context.Context, model.IntegrationFilter) ([]model.Integration, error)
	GetIntegrationById(context.Context, uuid.UUID) (*model.Integration, error)
	CreateIntegration(context.Context, model.Integration) (*model.Integration, error)
	GetDevice(ctx context.Context, deviceID string) (*model.Device, error)
	GetDeviceByIntegrationID(
		ctx context.Context,
		deviceID string,
		integrationID uuid.UUID,
	) (*model.Device, error)
	DoDevicesExistByIntegrationID(context.Context, uuid.UUID) (bool, error)
	// RemoveDevicesFromIntegration integration with the given integrationID
	// from all devices belonging to the tenant within the context.
	RemoveDevicesFromIntegration(
		ctx context.Context,
		integrationID uuid.UUID,
	) (deviceCount int64, err error)
	// UsertDeviceIntegrations adds the list of integration IDs to the
	// device and creates it if it does not exist.
	UpsertDeviceIntegrations(
		ctx context.Context,
		deviceID string,
		integrationIDs []uuid.UUID,
	) (newDevice *model.Device, err error)
	DeleteDevice(ctx context.Context, deviceID string) error
	SetIntegrationCredentials(context.Context, uuid.UUID, model.Credentials) error
	RemoveIntegration(context.Context, uuid.UUID) error

	// GetAllDevices returns an iterator over ALL devices sorted by tenant ID.
	GetAllDevices(ctx context.Context) (Iterator, error)

	// GetEvents returns list of event objects
	GetEvents(ctx context.Context, fltr model.EventsFilter) ([]model.Event, error)
	// SaveEvent saves the event in the database
	SaveEvent(ctx context.Context, event model.Event) error
	// DeleteTenantData removes all data belonging to a given tenant
	DeleteTenantData(
		ctx context.Context,
	) error
}

type Iterator interface {
	Next(ctx context.Context) bool
	Decode(value interface{}) error
	Close(ctx context.Context) error
}

var (
	ErrSerialization  = errors.New("store: failed to serialize object")
	ErrObjectNotFound = errors.New("store: object not found")

	ErrObjectExists = errors.New("store: the object already exists")
)
