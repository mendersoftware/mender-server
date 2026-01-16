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

package mongo

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/mendersoftware/mender-server/pkg/tiers"

	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/store"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

const (
	DbVersion = "1.1.0"

	DbName        = "inventory"
	DbDevicesColl = "devices"

	DbDevId              = "_id"
	DbDevAttributes      = "attributes"
	DbDevGroup           = "group"
	DbDevRevision        = "revision"
	DbDevUpdatedTs       = "updated_ts"
	DbDevAttributesText  = "text"
	DbDevAttributesTs    = "timestamp"
	DbDevAttributesDesc  = "description"
	DbDevAttributesValue = "value"
	DbDevAttributesScope = "scope"
	DbDevAttributesName  = "name"
	DbDevAttributesGroup = DbDevAttributes + "." +
		model.AttrScopeSystem + "-" + model.AttrNameGroup
	DbDevAttributesGroupValue = DbDevAttributesGroup + "." +
		DbDevAttributesValue

	DbScopeInventory = "inventory"

	FiltersAttributesMaxDevices = 5000
	FiltersAttributesLimit      = 500

	attrIdentityStatus = model.AttrScopeIdentity + "-status"
	attrSystemTier     = model.AttrScopeSystem + "-tier"
)

var (
	//with offcial mongodb supported driver we keep client
	clientGlobal *mongo.Client

	// once ensures client is created only once
	once sync.Once

	ErrNotFound = errors.New("mongo: no documents in result")
)

type DataStoreMongoConfig struct {
	// connection string
	ConnectionString string

	// SSL support
	SSL           bool
	SSLSkipVerify bool

	// Overwrites credentials provided in connection string if provided
	Username string
	Password string
}

type DataStoreMongo struct {
	client      *mongo.Client
	automigrate bool
}

func NewDataStoreMongoWithSession(client *mongo.Client) store.DataStore {
	return &DataStoreMongo{client: client}
}

// config.ConnectionString must contain a valid
func NewDataStoreMongo(config DataStoreMongoConfig) (store.DataStore, error) {
	//init master session
	var err error
	once.Do(func() {
		if !strings.Contains(config.ConnectionString, "://") {
			config.ConnectionString = "mongodb://" + config.ConnectionString
		}
		clientOptions := mopts.Client().ApplyURI(config.ConnectionString)

		if config.Username != "" {
			clientOptions.SetAuth(mopts.Credential{
				Username: config.Username,
				Password: config.Password,
			})
		}

		if config.SSL {
			tlsConfig := &tls.Config{}
			tlsConfig.InsecureSkipVerify = config.SSLSkipVerify
			clientOptions.SetTLSConfig(tlsConfig)
		}

		ctx := context.Background()
		l := log.FromContext(ctx)
		clientGlobal, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			l.Errorf("mongo: error connecting to mongo '%s'", err.Error())
			return
		}
		if clientGlobal == nil {
			l.Errorf("mongo: client is nil. wow.")
			return
		}
		// from: https://www.mongodb.com/blog/post/mongodb-go-driver-tutorial
		/*
			It is best practice to keep a client that is connected to MongoDB around so that the
			application can make use of connection pooling - you don't want to open and close a
			connection for each query. However, if your application no longer requires a connection,
			the connection can be closed with client.Disconnect() like so:
		*/
		err = clientGlobal.Ping(ctx, nil)
		if err != nil {
			clientGlobal = nil
			l.Errorf("mongo: error pinging mongo '%s'", err.Error())
			return
		}
		if clientGlobal == nil {
			l.Errorf("mongo: global instance of client is nil.")
			return
		}
	})

	if clientGlobal == nil {
		return nil, errors.New("failed to open mongo-driver session")
	}
	db := &DataStoreMongo{client: clientGlobal}

	return db, nil
}

func (db *DataStoreMongo) Ping(ctx context.Context) error {
	res := db.client.Database(DbName).RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

func (db *DataStoreMongo) GetDevices(
	ctx context.Context,
	q store.ListQuery,
) ([]model.Device, int, error) {
	c := db.client.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)

	queryFilters := make([]bson.M, 0)
	for _, filter := range q.Filters {
		op := mongoOperator(filter.Operator)
		name := fmt.Sprintf(
			"%s-%s",
			filter.AttrScope,
			model.GetDeviceAttributeNameReplacer().Replace(filter.AttrName),
		)
		field := fmt.Sprintf("%s.%s.%s", DbDevAttributes, name, DbDevAttributesValue)
		switch filter.Operator {
		default:
			if filter.ValueFloat != nil {
				queryFilters = append(queryFilters, bson.M{"$or": []bson.M{
					{field: bson.M{op: filter.Value}},
					{field: bson.M{op: filter.ValueFloat}},
				}})
			} else if filter.ValueTime != nil {
				queryFilters = append(queryFilters, bson.M{"$or": []bson.M{
					{field: bson.M{op: filter.Value}},
					{field: bson.M{op: filter.ValueTime}},
				}})
			} else {
				queryFilters = append(queryFilters, bson.M{field: bson.M{op: filter.Value}})
			}
		}
	}
	if q.GroupName != "" {
		groupFilter := bson.M{DbDevAttributesGroupValue: q.GroupName}
		queryFilters = append(queryFilters, groupFilter)
	}
	if q.HasGroup != nil {
		groupExistenceFilter := bson.M{
			DbDevAttributesGroup: bson.M{
				"$exists": *q.HasGroup,
			},
		}
		queryFilters = append(queryFilters, groupExistenceFilter)
	}

	findQuery := bson.M{}
	if len(queryFilters) > 0 {
		findQuery["$and"] = queryFilters
	}

	findOptions := mopts.Find()
	if q.Skip > 0 {
		findOptions.SetSkip(int64(q.Skip))
	}
	if q.Limit > 0 {
		findOptions.SetLimit(int64(q.Limit))
	}
	if q.Sort != nil {
		name := fmt.Sprintf(
			"%s-%s",
			q.Sort.AttrScope,
			model.GetDeviceAttributeNameReplacer().Replace(q.Sort.AttrName),
		)
		sortField := fmt.Sprintf("%s.%s.%s", DbDevAttributes, name, DbDevAttributesValue)
		sortFieldQuery := bson.D{{Key: sortField, Value: 1}}
		if !q.Sort.Ascending {
			sortFieldQuery[0].Value = -1
		}
		findOptions.SetSort(sortFieldQuery)
	}

	cursor, err := c.Find(ctx, findQuery, findOptions)
	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to search devices")
	}
	defer cursor.Close(ctx)

	devices := []model.Device{}
	if err = cursor.All(ctx, &devices); err != nil {
		return nil, -1, errors.Wrap(err, "failed to search devices")
	}

	count, err := c.CountDocuments(ctx, findQuery)
	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to count devices")
	}

	return devices, int(count), nil
}

func (db *DataStoreMongo) GetDevice(
	ctx context.Context,
	id model.DeviceID,
) (*model.Device, error) {
	var res model.Device
	c := db.client.
		Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl)
	l := log.FromContext(ctx)

	if id == model.NilDeviceID {
		return nil, nil
	}
	if err := c.FindOne(ctx, bson.M{DbDevId: id}).Decode(&res); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			l.Errorf("GetDevice: %v", err)
			return nil, errors.Wrap(err, "failed to fetch device")
		}
	}
	return &res, nil
}

// AddDevice inserts a new device, initializing the inventory data.
func (db *DataStoreMongo) AddDevice(ctx context.Context, dev *model.Device) error {
	if dev.Group != "" {
		dev.Attributes = append(dev.Attributes, model.DeviceAttribute{
			Scope: model.AttrScopeSystem,
			Name:  model.AttrNameGroup,
			Value: dev.Group,
		})
	}
	_, err := db.UpsertDevicesAttributesWithUpdated(
		ctx, []model.DeviceID{dev.ID}, dev.Attributes, "", "",
	)
	if err != nil {
		return errors.Wrap(err, "failed to store device")
	}
	return nil
}

func (db *DataStoreMongo) UpsertDevicesAttributesWithRevision(
	ctx context.Context,
	devices []model.DeviceUpdate,
	attrs model.DeviceAttributes,
) (*model.UpdateResult, error) {
	return db.upsertAttributes(ctx, devices, attrs, false, true, "", "", nil)
}

func (db *DataStoreMongo) UpsertDevicesAttributesWithUpdated(
	ctx context.Context,
	ids []model.DeviceID,
	attrs model.DeviceAttributes,
	scope string,
	etag string,
) (*model.UpdateResult, error) {
	withUpdated := scope == model.AttrScopeInventory
	return db.upsertAttributes(
		ctx, makeDevsWithIds(ids),
		attrs, withUpdated,
		false, scope,
		etag, nil,
	)
}

func (db *DataStoreMongo) UpsertDevicesAttributes(
	ctx context.Context,
	ids []model.DeviceID,
	attrs model.DeviceAttributes,
	notModifiedAfter *time.Time,
) (*model.UpdateResult, error) {
	return db.upsertAttributes(
		ctx, makeDevsWithIds(ids), attrs,
		false, false, "",
		"", notModifiedAfter,
	)
}

func makeDevsWithIds(ids []model.DeviceID) []model.DeviceUpdate {
	devices := make([]model.DeviceUpdate, len(ids))
	for i, id := range ids {
		devices[i].Id = id
	}
	return devices
}

func (db *DataStoreMongo) upsertAttributes(
	ctx context.Context,
	devices []model.DeviceUpdate,
	attrs model.DeviceAttributes,
	withUpdated bool,
	withRevision bool,
	scope string,
	etag string,
	notModifiedAfter *time.Time,
) (*model.UpdateResult, error) {
	const systemScope = DbDevAttributes + "." + model.AttrScopeSystem
	const createdField = systemScope + "-" + model.AttrNameCreated
	const etagField = model.AttrNameTagsEtag
	const updatedTS = "attributes.system-" + DbDevUpdatedTs + ".value"
	var (
		result *model.UpdateResult
		err    error
	)

	c := db.client.
		Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl)

	update, err := makeAttrUpsert(attrs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	oninsert := bson.M{
		createdField: model.DeviceAttribute{
			Scope: model.AttrScopeSystem,
			Name:  model.AttrNameCreated,
			Value: now,
		},
	}
	if !withRevision {
		oninsert["revision"] = 0
	}

	const updatedField = systemScope + "-" + model.AttrNameUpdated
	if withUpdated {
		update[updatedField] = model.DeviceAttribute{
			Scope: model.AttrScopeSystem,
			Name:  model.AttrNameUpdated,
			Value: now,
		}
	}

	switch len(devices) {
	case 0:
		return &model.UpdateResult{}, nil
	case 1:
		filter := bson.M{
			"_id": devices[0].Id,
		}
		updateOpts := mopts.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(mopts.After)

		if withRevision {
			filter[DbDevRevision] = bson.M{"$lt": devices[0].Revision}
			update[DbDevRevision] = devices[0].Revision
		}
		if scope == model.AttrScopeTags {
			update[etagField] = uuid.New().String()
			updateOpts = mopts.FindOneAndUpdate().
				SetUpsert(false).
				SetReturnDocument(mopts.After)
		}
		if etag != "" {
			filter[etagField] = bson.M{"$eq": etag}
		}
		if notModifiedAfter != nil {
			filter["$or"] = []bson.M{
				{updatedTS: bson.M{"$lte": notModifiedAfter}},
				{updatedTS: bson.M{"$exists": false}},
			}
		}

		update = bson.M{
			"$set":         update,
			"$setOnInsert": oninsert,
		}

		device := &model.Device{}
		res := c.FindOneAndUpdate(ctx, filter, update, updateOpts)
		err = res.Decode(device)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, store.ErrWriteConflict
			} else if err == mongo.ErrNoDocuments {
				return &model.UpdateResult{}, nil
			} else {
				return nil, err
			}
		}
		result = &model.UpdateResult{
			MatchedCount: 1,
			CreatedCount: 0,
			Devices:      []*model.Device{device},
		}
	default:
		var bres *mongo.BulkWriteResult
		// Perform single bulk-write operation
		// NOTE: Can't use UpdateMany as $in query operator does not
		//       upsert missing devices.

		models := make([]mongo.WriteModel, len(devices))
		for i, dev := range devices {
			umod := mongo.NewUpdateOneModel()
			filter := bson.M{"_id": dev.Id}
			if withRevision {
				filter[DbDevRevision] = bson.M{"$lt": dev.Revision}
				update[DbDevRevision] = dev.Revision
			}
			if notModifiedAfter != nil {
				filter["$or"] = []bson.M{
					{updatedTS: bson.M{"$lte": notModifiedAfter}},
					{updatedTS: bson.M{"$exists": false}},
				}
			}
			umod.Update = bson.M{
				"$set":         update,
				"$setOnInsert": oninsert,
			}
			umod.Filter = filter
			umod.SetUpsert(true)
			models[i] = umod
		}
		bres, err = c.BulkWrite(
			ctx, models, mopts.BulkWrite().SetOrdered(false),
		)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// bulk mode, swallow the error as we already updated the other devices
				// and the Matchedcount and CreatedCount values will tell the caller if
				// all the operations succeeded or not
				err = nil
			} else {
				return nil, err
			}
		}
		result = &model.UpdateResult{
			MatchedCount: bres.MatchedCount,
			CreatedCount: bres.UpsertedCount,
		}
	}
	return result, err
}

// makeAttrField is a convenience function for composing attribute field names.
func makeAttrField(attrName, attrScope string, subFields ...string) string {
	field := fmt.Sprintf(
		"%s.%s-%s",
		DbDevAttributes,
		attrScope,
		model.GetDeviceAttributeNameReplacer().Replace(attrName),
	)
	if len(subFields) > 0 {
		field = strings.Join(
			append([]string{field}, subFields...), ".",
		)
	}
	return field
}

// makeAttrUpsert creates a new upsert document for the given attributes.
func makeAttrUpsert(attrs model.DeviceAttributes) (bson.M, error) {
	var fieldName string
	upsert := make(bson.M)

	for i := range attrs {
		if attrs[i].Name == "" {
			return nil, store.ErrNoAttrName
		}
		if attrs[i].Scope == "" {
			// Default to inventory scope
			attrs[i].Scope = model.AttrScopeInventory
		}

		fieldName = makeAttrField(
			attrs[i].Name,
			attrs[i].Scope,
			DbDevAttributesScope,
		)
		upsert[fieldName] = attrs[i].Scope

		fieldName = makeAttrField(
			attrs[i].Name,
			attrs[i].Scope,
			DbDevAttributesName,
		)
		upsert[fieldName] = attrs[i].Name

		if attrs[i].Value != nil {
			fieldName = makeAttrField(
				attrs[i].Name,
				attrs[i].Scope,
				DbDevAttributesValue,
			)
			upsert[fieldName] = attrs[i].Value
		}

		if attrs[i].Description != nil {
			fieldName = makeAttrField(
				attrs[i].Name,
				attrs[i].Scope,
				DbDevAttributesDesc,
			)
			upsert[fieldName] = attrs[i].Description
		}

		if attrs[i].Timestamp != nil {
			fieldName = makeAttrField(
				attrs[i].Name,
				attrs[i].Scope,
				DbDevAttributesTs,
			)
			upsert[fieldName] = attrs[i].Timestamp
		}
	}
	return upsert, nil
}

// makeAttrRemove creates a new unset document to remove attributes
func makeAttrRemove(attrs model.DeviceAttributes) (bson.M, error) {
	var fieldName string
	remove := make(bson.M)

	for i := range attrs {
		if attrs[i].Name == "" {
			return nil, store.ErrNoAttrName
		}
		if attrs[i].Scope == "" {
			// Default to inventory scope
			attrs[i].Scope = model.AttrScopeInventory
		}
		fieldName = makeAttrField(
			attrs[i].Name,
			attrs[i].Scope,
		)
		remove[fieldName] = true
	}
	return remove, nil
}

func mongoOperator(co store.ComparisonOperator) string {
	switch co {
	case store.Eq:
		return "$eq"
	}
	return ""
}

func (db *DataStoreMongo) UpsertRemoveDeviceAttributes(
	ctx context.Context,
	id model.DeviceID,
	updateAttrs model.DeviceAttributes,
	removeAttrs model.DeviceAttributes,
	scope string,
	etag string,
) (*model.UpdateResult, error) {
	const systemScope = DbDevAttributes + "." + model.AttrScopeSystem
	const updatedField = systemScope + "-" + model.AttrNameUpdated
	const createdField = systemScope + "-" + model.AttrNameCreated
	const etagField = model.AttrNameTagsEtag
	var (
		err error
	)

	c := db.client.
		Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl)

	update, err := makeAttrUpsert(updateAttrs)
	if err != nil {
		return nil, err
	}
	remove, err := makeAttrRemove(removeAttrs)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": id}
	if etag != "" {
		filter[etagField] = bson.M{"$eq": etag}
	}

	updateOpts := mopts.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(mopts.After)
	if scope == model.AttrScopeTags {
		update[etagField] = uuid.New().String()
		updateOpts = updateOpts.SetUpsert(false)
	}
	now := time.Now()
	if scope != model.AttrScopeTags {
		update[updatedField] = model.DeviceAttribute{
			Scope: model.AttrScopeSystem,
			Name:  model.AttrNameUpdated,
			Value: now,
		}
	}
	update = bson.M{
		"$set": update,
		"$setOnInsert": bson.M{
			createdField: model.DeviceAttribute{
				Scope: model.AttrScopeSystem,
				Name:  model.AttrNameCreated,
				Value: now,
			},
		},
	}
	if len(remove) > 0 {
		update["$unset"] = remove
	}

	device := &model.Device{}
	res := c.FindOneAndUpdate(ctx, filter, update, updateOpts)
	err = res.Decode(device)
	if err == mongo.ErrNoDocuments {
		return &model.UpdateResult{
			MatchedCount: 0,
			CreatedCount: 0,
			Devices:      []*model.Device{},
		}, nil
	} else if err == nil {
		return &model.UpdateResult{
			MatchedCount: 1,
			CreatedCount: 0,
			Devices:      []*model.Device{device},
		}, nil
	}
	return nil, err
}

func (db *DataStoreMongo) UpdateDevicesGroup(
	ctx context.Context,
	devIDs []model.DeviceID,
	group model.GroupName,
) (*model.UpdateResult, error) {
	database := db.client.Database(mstore.DbFromContext(ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)

	var filter = bson.M{}
	switch len(devIDs) {
	case 0:
		return &model.UpdateResult{}, nil
	case 1:
		filter[DbDevId] = devIDs[0]
	default:
		filter[DbDevId] = bson.M{"$in": devIDs}
	}
	update := bson.M{
		"$set": bson.M{
			DbDevAttributesGroup: model.DeviceAttribute{
				Scope: model.AttrScopeSystem,
				Name:  DbDevGroup,
				Value: group,
			},
		},
	}
	res, err := collDevs.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &model.UpdateResult{
		MatchedCount: res.MatchedCount,
		UpdatedCount: res.ModifiedCount,
	}, nil
}

// UpdateDeviceText updates the device text field
func (db *DataStoreMongo) UpdateDeviceText(
	ctx context.Context,
	deviceID model.DeviceID,
	text string,
) error {
	filter := bson.M{
		DbDevId: deviceID.String(),
	}

	update := bson.M{
		"$set": bson.M{
			DbDevAttributesText: text,
		},
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)

	_, err := collDevs.UpdateOne(ctx, filter, update)
	return err
}

func (db *DataStoreMongo) GetFiltersAttributes(
	ctx context.Context,
) ([]model.FilterAttribute, error) {
	database := db.client.Database(mstore.DbFromContext(ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)

	const DbCount = "count"

	cur, err := collDevs.Aggregate(ctx, []bson.M{
		// Sample up to 5,000 devices to get a representative sample
		{
			"$limit": FiltersAttributesMaxDevices,
		},
		{
			"$project": bson.M{
				"_id": 0,
				"attributes": bson.M{
					"$objectToArray": "$" + DbDevAttributes,
				},
			},
		},
		{
			"$unwind": "$" + DbDevAttributes,
		},
		{
			"$group": bson.M{
				DbDevId: bson.M{
					DbDevAttributesName:  "$" + DbDevAttributes + ".v." + DbDevAttributesName,
					DbDevAttributesScope: "$" + DbDevAttributes + ".v." + DbDevAttributesScope,
				},
				DbCount: bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$limit": FiltersAttributesLimit,
		},
		{
			"$sort": bson.D{
				{Key: DbCount, Value: -1},
				{Key: DbDevId + "." + DbDevAttributesScope, Value: 1},
				{Key: DbDevId + "." + DbDevAttributesName, Value: 1},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var attributes []model.FilterAttribute
	type Result struct {
		Group struct {
			Name  string `bson:"name"`
			Scope string `bson:"scope"`
		} `bson:"_id"`
		Count int32 `bson:"count"`
	}
	for cur.Next(ctx) {
		var elem Result
		err = cur.Decode(&elem)
		if err != nil {
			break
		}
		attributes = append(attributes, model.FilterAttribute{
			Name:  elem.Group.Name,
			Scope: elem.Group.Scope,
			Count: elem.Count,
		})
	}

	return attributes, nil
}

func (db *DataStoreMongo) DeleteGroup(
	ctx context.Context,
	group model.GroupName,
) (chan model.DeviceID, error) {
	deviceIDs := make(chan model.DeviceID)

	database := db.client.Database(mstore.DbFromContext(ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)

	filter := bson.M{DbDevAttributesGroupValue: group}

	const batchMaxSize = 100
	batchSize := int32(batchMaxSize)
	findOptions := &mopts.FindOptions{
		Projection: bson.M{DbDevId: 1},
		BatchSize:  &batchSize,
	}
	cursor, err := collDevs.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	go func() {
		defer cursor.Close(ctx)
		batch := make([]model.DeviceID, batchMaxSize)
		batchSize := 0

		update := bson.M{"$unset": bson.M{DbDevAttributesGroup: 1}}
		device := &model.Device{}
		defer close(deviceIDs)

	next:
		for {
			hasNext := cursor.Next(ctx)
			if !hasNext {
				if batchSize > 0 {
					break
				}
				return
			}
			if err = cursor.Decode(&device); err == nil {
				batch[batchSize] = device.ID
				batchSize++
				if len(batch) == batchSize {
					break
				}
			}
		}

		_, _ = collDevs.UpdateMany(ctx, bson.M{DbDevId: bson.M{"$in": batch[:batchSize]}}, update)
		for _, item := range batch[:batchSize] {
			deviceIDs <- item
		}
		batchSize = 0
		goto next
	}()

	return deviceIDs, nil
}

func (db *DataStoreMongo) UnsetDevicesGroup(
	ctx context.Context,
	deviceIDs []model.DeviceID,
	group model.GroupName,
) (*model.UpdateResult, error) {
	database := db.client.Database(mstore.DbFromContext(ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)

	var filter bson.D
	// Add filter on device id (either $in or direct indexing)
	switch len(deviceIDs) {
	case 0:
		return &model.UpdateResult{}, nil
	case 1:
		filter = bson.D{{Key: DbDevId, Value: deviceIDs[0]}}
	default:
		filter = bson.D{{Key: DbDevId, Value: bson.M{"$in": deviceIDs}}}
	}
	// Append filter on group
	filter = append(
		filter,
		bson.E{Key: DbDevAttributesGroupValue, Value: group},
	)
	// Create unset operation on group attribute
	update := bson.M{
		"$unset": bson.M{
			DbDevAttributesGroup: "",
		},
	}
	res, err := collDevs.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &model.UpdateResult{
		MatchedCount: res.MatchedCount,
		UpdatedCount: res.ModifiedCount,
	}, nil
}

func predicateToQuery(pred model.FilterPredicate) (bson.D, error) {
	if err := pred.Validate(); err != nil {
		return nil, err
	}
	name := fmt.Sprintf(
		"%s.%s-%s.value",
		DbDevAttributes,
		pred.Scope,
		model.GetDeviceAttributeNameReplacer().Replace(pred.Attribute),
	)
	return bson.D{{
		Key: name, Value: bson.D{{Key: pred.Type, Value: pred.Value}},
	}}, nil
}

func (db *DataStoreMongo) ListGroups(
	ctx context.Context,
	filters []model.FilterPredicate,
) ([]model.GroupName, error) {
	c := db.client.
		Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl)

	fltr := bson.D{{
		Key: DbDevAttributesGroupValue, Value: bson.M{"$exists": true},
	}}
	if len(fltr) > 0 {
		for _, p := range filters {
			q, err := predicateToQuery(p)
			if err != nil {
				return nil, errors.Wrap(
					err, "store: bad filter predicate",
				)
			}
			fltr = append(fltr, q...)
		}
	}
	results, err := c.Distinct(
		ctx, DbDevAttributesGroupValue, fltr,
	)
	if err != nil {
		return nil, err
	}

	groups := make([]model.GroupName, len(results))
	for i, d := range results {
		groups[i] = model.GroupName(d.(string))
	}
	return groups, nil
}

func (db *DataStoreMongo) GetDevicesByGroup(
	ctx context.Context,
	group model.GroupName,
	skip,
	limit int,
) ([]model.DeviceID, int, error) {
	c := db.client.
		Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl)

	filter := bson.M{DbDevAttributesGroupValue: group}
	result := c.FindOne(ctx, filter)
	if result == nil {
		return nil, -1, store.ErrGroupNotFound
	}

	var dev model.Device
	err := result.Decode(&dev)
	if err != nil {
		return nil, -1, store.ErrGroupNotFound
	}

	hasGroup := group != ""
	devices, totalDevices, e := db.GetDevices(ctx,
		store.ListQuery{
			Skip:      skip,
			Limit:     limit,
			Filters:   nil,
			Sort:      nil,
			HasGroup:  &hasGroup,
			GroupName: string(group)})
	if e != nil {
		return nil, -1, errors.Wrap(e, "failed to get device list for group")
	}

	resIds := make([]model.DeviceID, len(devices))
	for i, d := range devices {
		resIds[i] = d.ID
	}
	return resIds, totalDevices, nil
}

func (db *DataStoreMongo) GetDeviceGroup(
	ctx context.Context,
	id model.DeviceID,
) (model.GroupName, error) {
	dev, err := db.GetDevice(ctx, id)
	if err != nil || dev == nil {
		return "", store.ErrDevNotFound
	}

	return dev.Group, nil
}

func (db *DataStoreMongo) DeleteDevices(
	ctx context.Context, ids []model.DeviceID,
) (*model.UpdateResult, error) {
	var filter = bson.M{}
	database := db.client.Database(mstore.DbFromContext(ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)

	switch len(ids) {
	case 0:
		// This is a no-op, don't bother requesting mongo.
		return &model.UpdateResult{DeletedCount: 0}, nil
	case 1:
		filter[DbDevId] = ids[0]
	default:
		filter[DbDevId] = bson.M{"$in": ids}
	}
	res, err := collDevs.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &model.UpdateResult{
		DeletedCount: res.DeletedCount,
	}, nil
}

func (db *DataStoreMongo) GetAllAttributeNames(ctx context.Context) ([]string, error) {
	c := db.client.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)

	project := bson.M{
		"$project": bson.M{
			"arrayofkeyvalue": bson.M{
				"$objectToArray": "$$ROOT.attributes",
			},
		},
	}

	unwind := bson.M{
		"$unwind": "$arrayofkeyvalue",
	}

	group := bson.M{
		"$group": bson.M{
			"_id": nil,
			"allkeys": bson.M{
				"$addToSet": "$arrayofkeyvalue.v.name",
			},
		},
	}

	l := log.FromContext(ctx)
	cursor, err := c.Aggregate(ctx, []bson.M{
		project,
		unwind,
		group,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	cursor.Next(ctx)
	elem := &bson.D{}
	err = cursor.Decode(elem)
	if err != nil {
		if err != io.EOF {
			return nil, errors.Wrap(err, "failed to get attributes")
		} else {
			return make([]string, 0), nil
		}
	}
	bsonValue, err := bson.Marshal(elem)
	if err != nil {
		return make([]string, 0), nil
	}
	var mapValue bson.M
	err = bson.Unmarshal(bsonValue, &mapValue)
	if err != nil {
		return make([]string, 0), nil
	}
	results := mapValue["allkeys"].(primitive.A)
	attributeNames := make([]string, len(results))
	for i, d := range results {
		attributeNames[i] = d.(string)
		l.Debugf("GetAllAttributeNames got: '%v'", d)
	}

	return attributeNames, nil
}

func (db *DataStoreMongo) SearchDevices(
	ctx context.Context,
	searchParams model.SearchParams,
) ([]model.Device, int, error) {
	c := db.client.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)

	queryFilters := make([]bson.M, 0)
	for _, filter := range searchParams.Filters {
		op := filter.Type
		var field string
		if filter.Scope == model.AttrScopeIdentity && filter.Attribute == model.AttrNameID {
			field = DbDevId
		} else {
			name := fmt.Sprintf(
				"%s-%s",
				filter.Scope,
				model.GetDeviceAttributeNameReplacer().Replace(filter.Attribute),
			)
			field = fmt.Sprintf("%s.%s.%s", DbDevAttributes, name, DbDevAttributesValue)
		}
		queryFilters = append(queryFilters, bson.M{field: bson.M{op: filter.Value}})
	}

	// FIXME: remove after migrating ids to attributes
	if len(searchParams.DeviceIDs) > 0 {
		queryFilters = append(queryFilters, bson.M{"_id": bson.M{"$in": searchParams.DeviceIDs}})
	}

	if searchParams.Text != "" {
		queryFilters = append(queryFilters, bson.M{
			"$text": bson.M{
				"$search": utils.TextToKeywords(searchParams.Text),
			},
		})
	}

	findQuery := bson.M{}
	if len(queryFilters) > 0 {
		findQuery["$and"] = queryFilters
	}

	findOptions := mopts.Find()
	findOptions.SetSkip(int64((searchParams.Page - 1) * searchParams.PerPage))
	findOptions.SetLimit(int64(searchParams.PerPage))

	if len(searchParams.Attributes) > 0 {
		name := fmt.Sprintf(
			"%s-%s",
			model.AttrScopeSystem,
			model.GetDeviceAttributeNameReplacer().Replace(DbDevUpdatedTs),
		)
		field := fmt.Sprintf("%s.%s", DbDevAttributes, name)
		projection := bson.M{field: 1}
		for _, attribute := range searchParams.Attributes {
			name := fmt.Sprintf(
				"%s-%s",
				attribute.Scope,
				model.GetDeviceAttributeNameReplacer().Replace(attribute.Attribute),
			)
			field := fmt.Sprintf("%s.%s", DbDevAttributes, name)
			projection[field] = 1
		}
		findOptions.SetProjection(projection)
	}

	if searchParams.Text != "" {
		findOptions.SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})
	} else if len(searchParams.Sort) > 0 {
		sortField := make(bson.D, len(searchParams.Sort))
		for i, sortQ := range searchParams.Sort {
			var field string
			if sortQ.Scope == model.AttrScopeIdentity && sortQ.Attribute == model.AttrNameID {
				field = DbDevId
			} else {
				name := fmt.Sprintf(
					"%s-%s",
					sortQ.Scope,
					model.GetDeviceAttributeNameReplacer().Replace(sortQ.Attribute),
				)
				field = fmt.Sprintf("%s.%s.value", DbDevAttributes, name)
			}
			sortField[i] = bson.E{Key: field, Value: 1}
			if sortQ.Order == "desc" {
				sortField[i].Value = -1
			}
		}
		findOptions.SetSort(sortField)
	}

	cursor, err := c.Find(ctx, findQuery, findOptions)
	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to search devices")
	}
	defer cursor.Close(ctx)

	devices := []model.Device{}

	if err = cursor.All(ctx, &devices); err != nil {
		return nil, -1, errors.Wrap(err, "failed to search devices")
	}

	count, err := c.CountDocuments(ctx, findQuery)
	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to search devices")
	}

	return devices, int(count), nil
}

func (db *DataStoreMongo) GetDeviceTierStatisticsByStatus(
	ctx context.Context,
) (
	*model.DeviceStatisticsByStatus,
	error,
) {
	collection := db.client.Database(
		mstore.DbFromContext(ctx, DbName),
	).Collection(DbDevicesColl)

	match := bson.M{
		"$match": bson.M{
			indexAttrName(attrIdentityStatus): bson.M{
				"$in": []string{
					model.DeviceStatusAccepted,
					model.DeviceStatusPending,
				},
			},
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"status": "$" + indexAttrName(attrIdentityStatus),
				"tier": bson.M{
					"$ifNull": bson.A{
						"$" + indexAttrName(attrSystemTier),
						tiers.StandardTier,
					},
				},
			},
			"devices": bson.M{
				"$sum": 1,
			},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":     0,
			"status":  "$_id.status",
			"tier":    "$_id.tier",
			"devices": 1,
		},
	}

	cursor, err := collection.Aggregate(
		ctx,
		[]bson.M{match, group, project},
	)
	if err != nil {
		return nil, err
	}

	var deviceCounts []struct {
		Status  string `bson:"status"`
		Tier    string `bson:"tier"`
		Devices uint64 `bson:"devices"`
	}

	err = cursor.All(ctx, &deviceCounts)
	if err != nil {
		return nil, err
	}

	var statistics model.DeviceStatisticsByStatus
	for _, count := range deviceCounts {
		var dest *model.DeviceCountPerTier
		if count.Status == model.DeviceStatusAccepted {
			dest = &statistics.Accepted
		} else {
			dest = &statistics.Pending
		}

		switch count.Tier {
		case tiers.StandardTier:
			dest.Standard = count.Devices
		case tiers.MicroTier:
			dest.Micro = count.Devices
		case tiers.SystemTier:
			dest.System = count.Devices
		}
	}

	return &statistics, nil
}

func indexAttr(s *mongo.Client, ctx context.Context, attr string) error {
	l := log.FromContext(ctx)
	c := s.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)

	indexView := c.Indexes()
	keys := bson.D{
		{Key: indexAttrName(attrIdentityStatus), Value: 1},
		{Key: indexAttrName(attr), Value: 1},
	}
	_, err := indexView.CreateOne(ctx, mongo.IndexModel{Keys: keys, Options: &mopts.IndexOptions{
		Name: &attr,
	}})

	if err != nil {
		if isTooManyIndexes(err) {
			l.Warnf(
				"failed to index attr %s in db %s: too many indexes",
				attr,
				mstore.DbFromContext(ctx, DbName),
			)
		} else {
			return errors.Wrapf(
				err,
				"failed to index attr %s in db %s",
				attr,
				mstore.DbFromContext(ctx, DbName),
			)
		}
	}

	return nil
}

func indexAttrName(attr string) string {
	return fmt.Sprintf("attributes.%s.value", attr)
}

func isTooManyIndexes(e error) bool {
	return strings.HasPrefix(e.Error(), "add index fails, too many indexes for inventory.devices")
}
