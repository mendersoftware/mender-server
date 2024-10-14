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

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

var (
	// device not found
	ErrDevNotFound = errors.New("Device not found")

	ErrGroupNotFound = errors.New("group not found")

	// ErrNoAttrName is returned if attributes are attempted upserted without
	// a Name identifier.
	ErrNoAttrName = errors.New("attribute name not present")

	// ErrWriteConflict represents a write conflict in the storage layer
	ErrWriteConflict = errors.New("write conflict")
)

//go:generate ../../../utils/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error

	GetDevices(ctx context.Context, q ListQuery) ([]model.Device, int, error)

	// find a device with given `id`, returns the device or nil,
	// if device was not found, error and returned device are nil
	GetDevice(ctx context.Context, id model.DeviceID) (*model.Device, error)

	// insert device into data store
	//
	// ds.AddDevice(&model.Device{
	// 	ID: "foo",
	// 	Attributes: map[string]string{"token": "123"),
	// })
	AddDevice(ctx context.Context, dev *model.Device) error

	// DeleteDevices removes devices with the given IDs from the database.
	DeleteDevices(ctx context.Context, ids []model.DeviceID) (*model.UpdateResult, error)

	// UpsertDevicesAttributesWithUpdated provides an interface to apply the same
	// attribute update to multiple devices. Attribute updates are performed
	// in a differential manner. Nonexistent attributes are created,
	// existing are overwritten; the device resource is also created if
	// necessary. It also sets the updated_ts timestamp to the current
	// date and time.
	UpsertDevicesAttributesWithUpdated(
		ctx context.Context,
		ids []model.DeviceID,
		attrs model.DeviceAttributes,
		scope string,
		etag string,
	) (*model.UpdateResult, error)

	// UpsertDevicesAttributes provides an interface to apply the same
	// attribute update to multiple devices. Attribute updates are performed
	// in a differential manner. Nonexistent attributes are created,
	// existing are overwritten; the device resource is also created if
	// necessary.
	UpsertDevicesAttributes(
		ctx context.Context,
		ids []model.DeviceID,
		attrs model.DeviceAttributes,
	) (*model.UpdateResult, error)

	// UpsertRemoveDeviceAttributes provides an interface to replace the
	// attributes for a device. It accepts two lists: a list of attributes
	// to upsert, and a list of attributes to remove. Nonexistent attributes
	// are created, existing are overwritten; the device resource is also
	// created if necessary
	UpsertRemoveDeviceAttributes(
		ctx context.Context,
		id model.DeviceID,
		updateAttrs model.DeviceAttributes,
		removeAttrs model.DeviceAttributes,
		scope string,
		etag string,
	) (*model.UpdateResult, error)
	// UpsertDevicesAttributesWithRevision upserts attributes for devices in the same way
	// UpsertDevicesAttributes does.
	// The only difference between this method and UpsertDevicesAttributes
	// method is that this method takes into accout device object revision
	// and upserts attributes only if the recorded object has revision lower than
	// the revision provided with the update.
	UpsertDevicesAttributesWithRevision(ctx context.Context,
		ids []model.DeviceUpdate,
		attrs model.DeviceAttributes,
	) (*model.UpdateResult, error)

	// UpdateDeviceText updates the device text field
	UpdateDeviceText(ctx context.Context,
		id model.DeviceID,
		text string,
	) error

	// GetFiltersAttributes returns the attributes which can be used
	// in filters
	GetFiltersAttributes(ctx context.Context) ([]model.FilterAttribute, error)

	// DeleteGroup removes a device group
	DeleteGroup(ctx context.Context, group model.GroupName) (chan model.DeviceID, error)

	// UnsetDevicesGroup removes a list of devices from their respective
	// groups returning the number of devices that were modified or an
	// error if any, respectively.
	UnsetDevicesGroup(ctx context.Context,
		deviceIDs []model.DeviceID,
		group model.GroupName,
	) (*model.UpdateResult, error)

	// UpdateDevicesGroup updates multiple devices' group, returning number
	// of matching devices, the number devices that changed group and error,
	// if any.
	UpdateDevicesGroup(ctx context.Context,
		devIDs []model.DeviceID,
		group model.GroupName,
	) (*model.UpdateResult, error)

	// ListGroups returns a list of all existing groups. Devices included
	// in the evaluation can be filtered by the filters argument.
	ListGroups(ctx context.Context, filters []model.FilterPredicate) ([]model.GroupName, error)

	// Lists devices belonging to a group
	GetDevicesByGroup(ctx context.Context,
		group model.GroupName,
		skip,
		limit int,
	) ([]model.DeviceID, int, error)

	// Get device's group
	GetDeviceGroup(ctx context.Context, id model.DeviceID) (model.GroupName, error)

	// Scan all devices in collection, grab all (unique) attribute names
	GetAllAttributeNames(ctx context.Context) ([]string, error)

	SearchDevices(ctx context.Context,
		searchParams model.SearchParams,
	) ([]model.Device, int, error)

	MigrateTenant(ctx context.Context, version string, tenantId string) error

	Migrate(ctx context.Context, version string) error

	WithAutomigrate() DataStore

	Maintenance(ctx context.Context, version string, tenantIDs ...string) error
}
