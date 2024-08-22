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

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"

	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

var (
	// device not found
	ErrDevNotFound = errors.New("device not found")
	// token not found
	ErrTokenNotFound = errors.New("token not found")
	// authorization set not found
	ErrAuthSetNotFound = errors.New("authorization set not found")
	// limit  set not found
	ErrLimitNotFound = errors.New("limit not found")
	// device already exists
	ErrObjectExists = errors.New("object exists")
	// device status unknown
	ErrDevStatusBroken = errors.New("cannot qualify device status")
)

const (
	NoCollectionErrMsg = "ns doesn't exist"
)

type AuthSetFilter struct {
	DeviceID string `bson:"device_id,omitempty"`
	Status   string `bson:"status,omitempty"`
}

// MapFunc is the prototype for the function applied to each database in
// ForEachTenant. dbCtx contains a reference to the current database
// MapFunc is applied to.
type MapFunc func(dbCtx context.Context) error

//go:generate ../utils/mockgen.sh
type DataStore interface {
	// ForEachTenant loops over all tenants and applies opFunc with
	// for the existing database. If opFunc returns an error for one of
	// the elements, this function aborts with the same error.
	ForEachTenant(parentCtx context.Context, opFunc MapFunc) error

	Ping(ctx context.Context) error
	// retrieve device by Mender-assigned device ID
	//returns ErrDevNotFound if device not found
	GetDeviceById(ctx context.Context, id string) (*model.Device, error)

	// retrieve device by its identity data
	// returns ErrDevNotFound if device not found
	GetDeviceByIdentityDataHash(ctx context.Context, idataHash []byte) (*model.Device, error)

	// list devices
	GetDevices(
		ctx context.Context,
		skip,
		limit uint,
		filter model.DeviceFilter,
	) ([]model.Device, error)

	AddDevice(ctx context.Context, d model.Device) error

	// updates a single device with deviceID, using data from `up`
	UpdateDevice(ctx context.Context, deviceID string, up model.DeviceUpdate) error

	// deletes device
	DeleteDevice(ctx context.Context, id string) error

	AddAuthSet(ctx context.Context, set model.AuthSet) error
	UpsertAuthSetStatus(ctx context.Context, authSet *model.AuthSet) error

	GetAuthSetByIdDataHashKey(
		ctx context.Context,
		idDataHash []byte,
		key string,
	) (*model.AuthSet, error)

	GetAuthSetByIdDataHashKeyByStatus(
		ctx context.Context,
		idDataHash []byte,
		key string,
		status string,
	) (*model.AuthSet, error)

	GetAuthSetById(ctx context.Context, id string) (*model.AuthSet, error)

	GetAuthSetsForDevice(ctx context.Context, devid string) ([]model.AuthSet, error)

	// RejectAuthSetsForDevice updates the auth set status for all auth sets
	// with status "accepted" or "preauthorized" to "rejected"
	RejectAuthSetsForDevice(ctx context.Context, deviceID string) error
	UpdateAuthSetById(ctx context.Context, authId string, mod model.AuthSetUpdate) error

	// deletes all auth sets for device
	DeleteAuthSetsForDevice(ctx context.Context, devid string) error

	// deletes authentication set for device
	DeleteAuthSetForDevice(ctx context.Context, devId string, authId string) error

	// adds JWT to database
	AddToken(ctx context.Context, t *jwt.Token) error

	// retrieves JWT from database using JWT Id and device Id
	// returns ErrTokenNotFound if token not found
	GetToken(ctx context.Context, jti oid.ObjectID) (*jwt.Token, error)

	// deletes token
	DeleteToken(ctx context.Context, jti oid.ObjectID) error

	// deletes all (tenant's) tokens (identity in context)
	DeleteTokens(ctx context.Context) error

	// deletes device token
	DeleteTokenByDevId(ctx context.Context, dev_id oid.ObjectID) error

	// put limit information into data store
	PutLimit(ctx context.Context, lim model.Limit) error

	// delete limit information from data store
	DeleteLimit(ctx context.Context, lim string) error

	// fetch limit information from data store
	GetLimit(ctx context.Context, name string) (*model.Limit, error)

	// get the number of devices with a given admission status
	// computed based on aggregated auth set statuses
	GetDevCountByStatus(ctx context.Context, status string) (int, error)

	// gets device status
	GetDeviceStatus(ctx context.Context, dev_id string) (string, error)

	MigrateTenant(ctx context.Context, version string, tenant string) error
	WithAutomigrate() DataStore
	//call this one if you really know what you are doing. This is supposed to be called only
	//from cmdPropagateStatusesInventory
	StoreMigrationVersion(ctx context.Context, version *migrate.Version) error
	ListTenantsIds(
		ctx context.Context,
	) ([]string, error)
}
