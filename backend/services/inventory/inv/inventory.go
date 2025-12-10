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

package inv

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/inventory/client/devicemonitor"
	"github.com/mendersoftware/mender-server/services/inventory/client/workflows"
	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/store"
	"github.com/mendersoftware/mender-server/services/inventory/store/mongo"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

const reindexBatchSize = 100

var (
	ErrETagDoesntMatch   = errors.New("ETag does not match")
	ErrTooManyAttributes = errors.New("the number of attributes in the scope is above the limit")
)

// this inventory service interface
//
//go:generate ../../../utils/mockgen.sh
type InventoryApp interface {
	WithReporting(c workflows.Client) InventoryApp
	HealthCheck(ctx context.Context) error
	ListDevices(ctx context.Context, q store.ListQuery) ([]model.Device, int, error)
	GetDevice(ctx context.Context, id model.DeviceID) (*model.Device, error)
	AddDevice(ctx context.Context, d *model.Device) error
	UpsertAttributes(ctx context.Context, id model.DeviceID, attrs model.DeviceAttributes,
		notModifiedAfter *time.Time) error
	UpsertAttributesWithUpdated(
		ctx context.Context,
		id model.DeviceID,
		attrs model.DeviceAttributes,
		scope string,
		etag string,
	) error
	UpsertDevicesStatuses(
		ctx context.Context,
		devices []model.DeviceUpdate,
		attrs model.DeviceAttributes,
	) (*model.UpdateResult, error)
	ReplaceAttributes(
		ctx context.Context,
		id model.DeviceID,
		upsertAttrs model.DeviceAttributes,
		scope string,
		etag string,
	) error
	GetFiltersAttributes(ctx context.Context) ([]model.FilterAttribute, error)
	DeleteGroup(ctx context.Context, groupName model.GroupName) (*model.UpdateResult, error)
	UnsetDeviceGroup(ctx context.Context, id model.DeviceID, groupName model.GroupName) error
	UnsetDevicesGroup(
		ctx context.Context,
		deviceIDs []model.DeviceID,
		groupName model.GroupName,
	) (*model.UpdateResult, error)
	UpdateDeviceGroup(ctx context.Context, id model.DeviceID, group model.GroupName) error
	UpdateDevicesGroup(
		ctx context.Context,
		ids []model.DeviceID,
		group model.GroupName,
	) (*model.UpdateResult, error)
	ListGroups(ctx context.Context, filters []model.FilterPredicate) ([]model.GroupName, error)
	ListDevicesByGroup(
		ctx context.Context,
		group model.GroupName,
		skip int,
		limit int,
	) ([]model.DeviceID, int, error)
	GetDeviceGroup(ctx context.Context, id model.DeviceID) (model.GroupName, error)
	DeleteDevice(ctx context.Context, id model.DeviceID) error
	DeleteDevices(
		ctx context.Context,
		ids []model.DeviceID,
	) (*model.UpdateResult, error)
	CreateTenant(ctx context.Context, tenant model.NewTenant) error
	SearchDevices(ctx context.Context, searchParams model.SearchParams) ([]model.Device, int, error)
	CheckAlerts(ctx context.Context, deviceId string) (int, error)
	WithLimits(attributes, tags int) InventoryApp
	WithDevicemonitor(client devicemonitor.Client) InventoryApp
	GetDeviceStatistics(ctx context.Context) (*model.DeviceStatistics, error)
}

type inventory struct {
	db              store.DataStore
	limitAttributes int
	limitTags       int
	dmClient        devicemonitor.Client
	enableReporting bool
	wfClient        workflows.Client
}

func NewInventory(d store.DataStore) InventoryApp {
	return &inventory{db: d}
}

func (i *inventory) WithDevicemonitor(client devicemonitor.Client) InventoryApp {
	i.dmClient = client
	return i
}

func (i *inventory) WithLimits(limitAttributes, limitTags int) InventoryApp {
	i.limitAttributes = limitAttributes
	i.limitTags = limitTags
	return i
}

func (i *inventory) WithReporting(client workflows.Client) InventoryApp {
	i.enableReporting = true
	i.wfClient = client
	return i
}

func (i *inventory) HealthCheck(ctx context.Context) error {
	err := i.db.Ping(ctx)
	if err != nil {
		return errors.Wrap(err, "error reaching MongoDB")
	}

	if i.enableReporting {
		err := i.wfClient.CheckHealth(ctx)
		if err != nil {
			return errors.Wrap(err, "error reaching workflows")
		}
	}

	return nil
}

func (i *inventory) ListDevices(
	ctx context.Context,
	q store.ListQuery,
) ([]model.Device, int, error) {
	devs, totalCount, err := i.db.GetDevices(ctx, q)

	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to fetch devices")
	}

	return devs, totalCount, nil
}

func (i *inventory) GetDevice(ctx context.Context, id model.DeviceID) (*model.Device, error) {
	dev, err := i.db.GetDevice(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch device")
	}
	return dev, nil
}

func (i *inventory) AddDevice(ctx context.Context, dev *model.Device) error {
	if dev == nil {
		return errors.New("no device given")
	}
	dev.Text = utils.GetTextField(dev)
	err := i.db.AddDevice(ctx, dev)
	if err != nil {
		return errors.Wrap(err, "failed to add device")
	}

	i.maybeTriggerReindex(ctx, []model.DeviceID{dev.ID})

	return nil
}

func (i *inventory) DeleteDevices(
	ctx context.Context,
	ids []model.DeviceID,
) (*model.UpdateResult, error) {
	res, err := i.db.DeleteDevices(ctx, ids)
	if err != nil {
		return nil, err
	}

	if i.enableReporting {
		for _, d := range ids {
			i.triggerReindex(ctx, []model.DeviceID{d})
		}
	}

	return res, err
}

func (i *inventory) DeleteDevice(ctx context.Context, id model.DeviceID) error {
	res, err := i.db.DeleteDevices(ctx, []model.DeviceID{id})
	if err != nil {
		return errors.Wrap(err, "failed to delete device")
	} else if res.DeletedCount < 1 {
		return store.ErrDevNotFound
	}
	i.maybeTriggerReindex(ctx, []model.DeviceID{id})

	return nil
}

func (i *inventory) UpsertAttributes(
	ctx context.Context,
	id model.DeviceID,
	attrs model.DeviceAttributes,
	notModifiedAfter *time.Time,
) error {
	res, err := i.db.UpsertDevicesAttributes(
		ctx, []model.DeviceID{id}, attrs, notModifiedAfter,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert attributes in db")
	}
	if res != nil && res.MatchedCount > 0 {
		i.reindexTextField(ctx, res.Devices)
		i.maybeTriggerReindex(ctx, []model.DeviceID{id})
	}
	return nil
}

func (i *inventory) checkAttributesLimits(
	ctx context.Context,
	id model.DeviceID,
	attrs model.DeviceAttributes,
	scope string,
) error {
	limit := 0
	switch scope {
	case model.AttrScopeInventory:
		limit = i.limitAttributes
	case model.AttrScopeTags:
		limit = i.limitTags
	}
	if limit == 0 {
		return nil
	}
	device, err := i.db.GetDevice(ctx, id)
	if err != nil && err != store.ErrDevNotFound {
		return errors.Wrap(err, "failed to get the device")
	} else if device == nil {
		return nil
	}
	count := 0
	for _, attr := range device.Attributes {
		if attr.Scope == scope {
			count += 1
			if count > limit {
				break
			}
		}
	}
	for _, attr := range attrs {
		if count > limit {
			break
		}
		found := false
		for _, devAttr := range device.Attributes {
			if attr.Scope == scope && attr.Name == devAttr.Name {
				found = true
			}
		}
		if !found {
			count++
		}
	}
	if count > limit {
		return ErrTooManyAttributes
	}
	return nil
}

const oneDay = 24 * time.Hour

func (i *inventory) needsUpsert(
	device *model.Device,
	upsertAttrs model.DeviceAttributes,
	removeAttrs model.DeviceAttributes,
) bool {
	needsUpsert := true
	if device != nil {
		// we update the inventory attributes at least once every (calendar) day
		if device.UpdatedTs != nil &&
			device.UpdatedTs.Truncate(oneDay) != time.Now().Truncate(oneDay) {
			return true
		}
		needsUpsert = false
		for _, attribute := range upsertAttrs {
			if attribute.Scope != model.AttrScopeInventory {
				needsUpsert = true
				break
			}
			attributeChanged := true
			for _, deviceAttribute := range device.Attributes {
				if attribute.Scope != deviceAttribute.Scope ||
					attribute.Name != deviceAttribute.Name {
					continue
				}
				if _, ok := deviceAttribute.Value.(primitive.A); ok {
					attributeChanged = !reflect.DeepEqual(attribute.Value,
						[]interface{}(deviceAttribute.Value.(primitive.A)))
				} else {
					attributeChanged = !reflect.DeepEqual(attribute.Value, deviceAttribute.Value)
				}
				break
			}
			if attributeChanged {
				needsUpsert = true
				break
			}
		}
		if !needsUpsert && len(removeAttrs) > 0 {
		OuterLoop:
			for _, attribute := range removeAttrs {
				for _, deviceAttribute := range device.Attributes {
					if attribute.Scope == deviceAttribute.Scope &&
						attribute.Name == deviceAttribute.Name {
						needsUpsert = true
						break OuterLoop
					}
				}
			}
		}
	}
	return needsUpsert
}

func (i *inventory) UpsertAttributesWithUpdated(
	ctx context.Context,
	id model.DeviceID,
	attrs model.DeviceAttributes,
	scope string,
	etag string,
) error {
	if err := i.checkAttributesLimits(ctx, id, attrs, scope); err != nil {
		return err
	}

	device, err := i.db.GetDevice(ctx, id)
	if err != nil && err != store.ErrDevNotFound {
		return errors.Wrap(err, "failed to get the device")
	} else if !i.needsUpsert(device, attrs, nil) {
		return nil
	}

	res, err := i.db.UpsertDevicesAttributesWithUpdated(
		ctx, []model.DeviceID{id}, attrs, scope, etag,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert attributes in db")
	}
	if scope == model.AttrScopeTags {
		if res != nil && res.MatchedCount == 0 && etag != "" {
			return ErrETagDoesntMatch
		}
	}

	if res != nil && res.MatchedCount > 0 {
		i.reindexTextField(ctx, res.Devices)
		i.maybeTriggerReindex(ctx, []model.DeviceID{id})
	}
	return nil
}

func getRemoveAttrs(
	device *model.Device,
	scope string,
	upsertAttrs model.DeviceAttributes,
) model.DeviceAttributes {
	removeAttrs := model.DeviceAttributes{}
	if device != nil {
		for _, attr := range device.Attributes {
			if attr.Scope == scope {
				update := false
				for _, upsertAttr := range upsertAttrs {
					if upsertAttr.Name == attr.Name {
						update = true
						break
					}
				}
				if !update {
					removeAttrs = append(removeAttrs, attr)
				}
			}
		}
	}
	return removeAttrs
}

func (i *inventory) ReplaceAttributes(
	ctx context.Context,
	id model.DeviceID,
	upsertAttrs model.DeviceAttributes,
	scope string,
	etag string,
) error {
	limit := 0
	switch scope {
	case model.AttrScopeInventory:
		limit = i.limitAttributes
	case model.AttrScopeTags:
		limit = i.limitTags
	}
	if limit > 0 && len(upsertAttrs) > limit {
		return ErrTooManyAttributes
	}

	device, err := i.db.GetDevice(ctx, id)
	if err != nil && err != store.ErrDevNotFound {
		return errors.Wrap(err, "failed to get the device")
	}

	removeAttrs := getRemoveAttrs(device, scope, upsertAttrs)
	if !i.needsUpsert(device, upsertAttrs, removeAttrs) {
		return nil
	}

	res, err := i.db.UpsertRemoveDeviceAttributes(ctx, id, upsertAttrs, removeAttrs, scope, etag)
	if err != nil {
		return errors.Wrap(err, "failed to replace attributes in db")
	}
	if scope == model.AttrScopeTags {
		if res != nil && res.MatchedCount == 0 && etag != "" {
			return ErrETagDoesntMatch
		}
	}
	if res != nil && res.MatchedCount > 0 {
		i.reindexTextField(ctx, res.Devices)
		i.maybeTriggerReindex(ctx, []model.DeviceID{id})
	}
	return nil
}

func (i *inventory) GetFiltersAttributes(ctx context.Context) ([]model.FilterAttribute, error) {
	attributes, err := i.db.GetFiltersAttributes(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get filter attributes from the db")
	}
	return attributes, nil
}

func (i *inventory) DeleteGroup(
	ctx context.Context,
	groupName model.GroupName,
) (*model.UpdateResult, error) {
	deviceIDs, err := i.db.DeleteGroup(ctx, groupName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete group")
	}

	batchDeviceIDsLength := 0
	batchDeviceIDs := make([]model.DeviceID, reindexBatchSize)

	triggerReindex := func() {
		i.maybeTriggerReindex(ctx, batchDeviceIDs[0:batchDeviceIDsLength])
		batchDeviceIDsLength = 0
	}

	res := &model.UpdateResult{}
	for deviceID := range deviceIDs {
		batchDeviceIDs[batchDeviceIDsLength] = deviceID
		batchDeviceIDsLength++
		if batchDeviceIDsLength == reindexBatchSize {
			triggerReindex()
		}
		res.MatchedCount += 1
		res.UpdatedCount += 1
	}
	if batchDeviceIDsLength > 0 {
		triggerReindex()
	}

	return res, err
}

func (i *inventory) UpsertDevicesStatuses(
	ctx context.Context,
	devices []model.DeviceUpdate,
	attrs model.DeviceAttributes,
) (*model.UpdateResult, error) {
	res, err := i.db.UpsertDevicesAttributesWithRevision(ctx, devices, attrs)
	if err != nil {
		return nil, err
	}

	if i.enableReporting {
		deviceIDs := make([]model.DeviceID, len(devices))
		for i, d := range devices {
			deviceIDs[i] = d.Id
		}
		i.triggerReindex(ctx, deviceIDs)
	}

	return res, err
}

func (i *inventory) UnsetDevicesGroup(
	ctx context.Context,
	deviceIDs []model.DeviceID,
	groupName model.GroupName,
) (*model.UpdateResult, error) {
	res, err := i.db.UnsetDevicesGroup(ctx, deviceIDs, groupName)
	if err != nil {
		return nil, err
	}

	if i.enableReporting {
		i.triggerReindex(ctx, deviceIDs)
	}

	return res, nil
}

func (i *inventory) UnsetDeviceGroup(
	ctx context.Context,
	id model.DeviceID,
	group model.GroupName,
) error {
	result, err := i.db.UnsetDevicesGroup(ctx, []model.DeviceID{id}, group)
	if err != nil {
		return errors.Wrap(err, "failed to unassign group from device")
	} else if result.MatchedCount <= 0 {
		return store.ErrDevNotFound
	}

	i.maybeTriggerReindex(ctx, []model.DeviceID{id})

	return nil
}

func (i *inventory) UpdateDevicesGroup(
	ctx context.Context,
	deviceIDs []model.DeviceID,
	group model.GroupName,
) (*model.UpdateResult, error) {

	res, err := i.db.UpdateDevicesGroup(ctx, deviceIDs, group)
	if err != nil {
		return nil, err
	}

	if i.enableReporting {
		i.triggerReindex(ctx, deviceIDs)
	}

	return res, err
}

func (i *inventory) UpdateDeviceGroup(
	ctx context.Context,
	devid model.DeviceID,
	group model.GroupName,
) error {
	result, err := i.db.UpdateDevicesGroup(
		ctx, []model.DeviceID{devid}, group,
	)
	if err != nil {
		return errors.Wrap(err, "failed to add device to group")
	} else if result.MatchedCount <= 0 {
		return store.ErrDevNotFound
	}

	i.maybeTriggerReindex(ctx, []model.DeviceID{devid})

	return nil
}

func (i *inventory) ListGroups(
	ctx context.Context,
	filters []model.FilterPredicate,
) ([]model.GroupName, error) {
	groups, err := i.db.ListGroups(ctx, filters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list groups")
	}

	if groups == nil {
		return []model.GroupName{}, nil
	}
	return groups, nil
}

func (i *inventory) ListDevicesByGroup(
	ctx context.Context,
	group model.GroupName,
	skip,
	limit int,
) ([]model.DeviceID, int, error) {
	ids, totalCount, err := i.db.GetDevicesByGroup(ctx, group, skip, limit)
	if err != nil {
		if err == store.ErrGroupNotFound {
			return nil, -1, err
		} else {
			return nil, -1, errors.Wrap(err, "failed to list devices by group")
		}
	}

	return ids, totalCount, nil
}

func (i *inventory) GetDeviceGroup(
	ctx context.Context,
	id model.DeviceID,
) (model.GroupName, error) {
	group, err := i.db.GetDeviceGroup(ctx, id)
	if err != nil {
		if err == store.ErrDevNotFound {
			return "", err
		} else {
			return "", errors.Wrap(err, "failed to get device's group")
		}
	}

	return group, nil
}

func (i *inventory) CreateTenant(ctx context.Context, tenant model.NewTenant) error {
	if err := i.db.WithAutomigrate().
		MigrateTenant(ctx, mongo.DbVersion, tenant.ID); err != nil {
		return errors.Wrapf(err, "failed to apply migrations for tenant %v", tenant.ID)
	}
	return nil
}

func (i *inventory) SearchDevices(
	ctx context.Context,
	searchParams model.SearchParams,
) ([]model.Device, int, error) {
	devs, totalCount, err := i.db.SearchDevices(ctx, searchParams)

	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to fetch devices")
	}

	return devs, totalCount, nil
}

func (i *inventory) CheckAlerts(ctx context.Context, deviceId string) (int, error) {
	return i.dmClient.CheckAlerts(ctx, deviceId)
}

// maybeTriggerReindex conditionally triggers the reindex_reporting workflow for a device
func (i *inventory) maybeTriggerReindex(ctx context.Context, deviceIDs []model.DeviceID) {
	if i.enableReporting {
		i.triggerReindex(ctx, deviceIDs)
	}
}

// triggerReindex triggers the reindex_reporting workflow for a device
func (i *inventory) triggerReindex(ctx context.Context, deviceIDs []model.DeviceID) {
	err := i.wfClient.StartReindex(ctx, deviceIDs)
	if err != nil {
		l := log.FromContext(ctx)
		l.Errorf("failed to start reindex_reporting for devices %v, error: %v", deviceIDs, err)
	}
}

// reindexTextField reindex the device's text field
func (i *inventory) reindexTextField(ctx context.Context, devices []*model.Device) {
	l := log.FromContext(ctx)
	for _, device := range devices {
		text := utils.GetTextField(device)
		if device.Text != text {
			err := i.db.UpdateDeviceText(ctx, device.ID, text)
			if err != nil {
				l.Errorf("failed to reindex the text field for device %v, error: %v",
					device.ID, err)
			}
		}
	}
}

func (i *inventory) GetDeviceStatistics(ctx context.Context) (*model.DeviceStatistics, error) {
	byStatus, err := i.db.GetDeviceTierStatisticsByStatus(ctx)
	if err != nil {
		return nil, err
	}

	if byStatus == nil {
		return nil, errors.New("found no device statistics")
	}

	return &model.DeviceStatistics{
		ByStatus: *byStatus,
	}, nil
}
