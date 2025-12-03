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
	"iter"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store/v2"

	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
	uto "github.com/mendersoftware/mender-server/services/deviceauth/utils/to"
)

const (
	DbVersion     = "2.0.0"
	DbName        = "deviceauth"
	DbDevicesColl = "devices"
	DbAuthSetColl = "auth_sets"
	DbTokensColl  = "tokens"
	DbLimitsColl  = "limits"

	DbKeyDeviceRevision = "revision"
	dbFieldID           = "_id"
	dbFieldTenantID     = "tenant_id"
	dbFieldIDDataSha    = "id_data_sha256"
	dbFieldStatus       = "status"
	dbFieldDeviceID     = "device_id"
	dbFieldPubKey       = "pubkey"
	dbFieldExpTime      = "exp.time"
	dbFieldTenantClaim  = "mender.tenant"
	dbFieldName         = "name"
	dbFieldSubject      = "sub"
)

var (
	indexDevices_IdentityData                       = "devices:IdentityData"
	indexDevices_IdentityDataSha256                 = "devices:IdentityDataSha256"
	indexDevices_Status                             = "devices:Status"
	indexAuthSet_DeviceId_IdentityData_PubKey       = "auth_sets:DeviceId:IdData:PubKey"
	indexAuthSet_DeviceId_IdentityDataSha256_PubKey = "auth_sets:IdDataSha256:PubKey"
	indexAuthSet_IdentityDataSha256_PubKey          = "auth_sets:NoDeviceId:IdDataSha256:PubKey"
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
	multitenant bool
}

func NewDataStoreMongoWithClient(client *mongo.Client) *DataStoreMongo {
	return &DataStoreMongo{
		client: client,
	}
}

func NewDataStoreMongo(config DataStoreMongoConfig) (*DataStoreMongo, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mongo client")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify mongodb connection")
	}

	return NewDataStoreMongoWithClient(client), nil
}

func (db *DataStoreMongo) ForEachTenant(
	ctx context.Context,
	mapFunc store.MapFunc,
) error {
	var (
		dbCtx   context.Context
		err     error
		errChan = make(chan error, 1)
	)
	tenantsIds, err := db.ListTenantsIds(ctx)
	if err != nil {
		return errors.Wrap(err, "store: failed to retrieve tenants")
	}
	go func() {
		for _, tenantID := range tenantsIds {
			if ctx.Err() != nil {
				return
			}
			if tenantID != "" {
				dbCtx = identity.WithContext(ctx,
					&identity.Identity{
						Tenant: tenantID,
					},
				)
			} else {
				dbCtx = ctx
			}
			err := mapFunc(dbCtx)
			if err != nil {
				errChan <- errors.Wrapf(err,
					`store: failed to apply mapFunc to tenant "%s"`,
					tenantID,
				)
			}
		}
		errChan <- nil
	}()

	select {
	case err = <-errChan:
	case <-ctx.Done():
		err = errors.Wrap(ctx.Err(),
			"store: database operations stopped prematurely",
		)
	}
	return err
}

func (db *DataStoreMongo) Ping(ctx context.Context) error {
	return db.client.Ping(ctx, nil)
}

type DeviceFilter model.DeviceFilter

func (fltr DeviceFilter) MarshalBSON() (b []byte, err error) {
	doc := bson.D{}
	switch len(fltr.IDs) {
	case 0:
		break
	case 1:
		doc = append(doc, bson.E{Key: "_id", Value: fltr.IDs[0]})
	default:
		doc = append(doc, bson.E{
			Key: "_id", Value: bson.D{{
				Key: "$in", Value: fltr.IDs,
			}},
		})
	}
	switch len(fltr.Status) {
	case 0:
		break
	case 1:
		doc = append(doc, bson.E{Key: "status", Value: fltr.Status[0]})
	default:
		doc = append(doc, bson.E{
			Key: "status", Value: bson.D{{
				Key: "$in", Value: fltr.Status,
			}},
		})
	}

	return bson.Marshal(doc)
}

func (db *DataStoreMongo) GetDevices(
	ctx context.Context,
	skip,
	limit uint,
	filter model.DeviceFilter,
) ([]model.Device, error) {
	const MaxInt64 = int64(^uint64(1 << 63))
	var (
		res  = []model.Device{}
		fltr = DeviceFilter(filter)
	)
	collDevs := db.client.
		Database(DbName).
		Collection(DbDevicesColl)

	findOpts := mopts.Find().
		SetSort(bson.D{
			{Key: dbFieldStatus, Value: 1},
			{Key: dbFieldID, Value: 1},
		}).
		SetSkip(int64(skip) & MaxInt64)

	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	cursor, err := collDevs.Find(ctx, ctxstore.WithTenantID(ctx, fltr), findOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch device list")
	}
	if err := cursor.All(ctx, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (db *DataStoreMongo) StoreMigrationVersion(
	ctx context.Context,
	version *migrate.Version,
) error {
	if version == nil {
		return errors.New("version cant be nil.")
	}

	c := db.client.Database(DbName).
		Collection(migrate.DbMigrationsColl)

	migrationInfo := migrate.MigrationEntry{
		Version:   *version,
		Timestamp: time.Now(),
	}
	_, err := c.InsertOne(ctx, migrationInfo)
	return err
}

func (db *DataStoreMongo) GetDeviceById(ctx context.Context, id string) (*model.Device, error) {

	c := db.client.Database(DbName).Collection(DbDevicesColl)

	res := model.Device{}

	err := c.FindOne(ctx, ctxstore.WithTenantID(ctx, bson.M{"_id": id})).Decode(&res)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrDevNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch device")
		}
	}

	return &res, nil
}

func (db *DataStoreMongo) GetDeviceByIdentityDataHash(
	ctx context.Context,
	idataHash []byte,
) (*model.Device, error) {
	c := db.client.Database(DbName).Collection(DbDevicesColl)

	filter := ctxstore.WithTenantID(ctx, bson.M{dbFieldIDDataSha: idataHash})
	res := model.Device{}

	err := c.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrDevNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch device")
		}
	}

	return &res, nil
}

func (db *DataStoreMongo) AddDevice(ctx context.Context, d model.Device) error {

	if d.Id == "" {
		uid := oid.NewUUIDv4()
		d.Id = uid.String()
	}

	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}
	d.TenantID = tenantId

	c := db.client.Database(DbName).Collection(DbDevicesColl)

	if _, err := c.InsertOne(ctx, d); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return store.ErrObjectExists
		}
		return errors.Wrap(err, "failed to store device")
	}
	return nil
}

func (db *DataStoreMongo) UpdateDevice(ctx context.Context,
	deviceID string, updev model.DeviceUpdate) error {

	c := db.client.Database(DbName).Collection(DbDevicesColl)

	updev.UpdatedTs = uto.TimePtr(time.Now().UTC())
	update := bson.M{
		"$inc": bson.M{
			DbKeyDeviceRevision: 1,
		},
		"$set": updev,
	}

	res, err := c.UpdateOne(ctx, ctxstore.WithTenantID(ctx, bson.M{"_id": deviceID}), update)
	if err != nil {
		return errors.Wrap(err, "failed to update device")
	} else if res.MatchedCount < 1 {
		return store.ErrDevNotFound
	}

	return nil
}

func (db *DataStoreMongo) DeleteDevice(ctx context.Context, id string) error {

	c := db.client.Database(DbName).Collection(DbDevicesColl)

	filter := ctxstore.WithTenantID(ctx, bson.M{"_id": id})
	result, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "failed to remove device")
	} else if result.DeletedCount < 1 {
		return store.ErrDevNotFound
	}

	return nil
}

func (db *DataStoreMongo) AddToken(ctx context.Context, t *jwt.Token) error {
	database := db.client.Database(DbName)
	collTokens := database.Collection(DbTokensColl)

	filter := bson.M{"_id": t.Claims.ID}
	update := bson.M{"$set": t}
	updateOpts := mopts.Update()
	updateOpts.SetUpsert(true)

	if _, err := collTokens.UpdateOne(
		ctx, filter, update, updateOpts,
	); err != nil {
		return errors.Wrap(err, "failed to store token")
	}

	return nil
}

func (db *DataStoreMongo) GetToken(
	ctx context.Context,
	jti oid.ObjectID,
) (*jwt.Token, error) {

	c := db.client.Database(DbName).Collection(DbTokensColl)

	res := jwt.Token{}

	err := c.FindOne(ctx, bson.M{"_id": jti}).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrTokenNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch token")
		}
	}

	return &res, nil
}

func (db *DataStoreMongo) DeleteToken(ctx context.Context, jti oid.ObjectID) error {
	database := db.client.Database(DbName)
	collTokens := database.Collection(DbTokensColl)

	filter := bson.M{"_id": jti}
	result, err := collTokens.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "failed to remove token")
	} else if result.DeletedCount < 1 {
		return store.ErrTokenNotFound
	}

	return nil
}

func (db *DataStoreMongo) DeleteTokens(ctx context.Context) error {
	c := db.client.Database(DbName).
		Collection(DbTokensColl)

	id := identity.FromContext(ctx)
	filter := bson.M{dbFieldTenantClaim: nil}
	if id != nil && id.Tenant != "" {
		filter[dbFieldTenantClaim] = id.Tenant
	}
	_, err := c.DeleteMany(ctx, filter)

	return err
}

func (db *DataStoreMongo) DeleteTokenByDevId(
	ctx context.Context,
	devID oid.ObjectID,
) error {
	c := db.client.Database(DbName).
		Collection(DbTokensColl)
	filter := bson.M{"sub": devID, dbFieldTenantClaim: nil}
	id := identity.FromContext(ctx)
	if id != nil && id.Tenant != "" {
		filter[dbFieldTenantClaim] = id.Tenant
	}
	ci, err := c.DeleteMany(ctx, filter)

	if err != nil {
		return errors.Wrap(err, "failed to remove tokens")
	}

	if ci.DeletedCount == 0 {
		return store.ErrTokenNotFound
	}

	return nil
}

func (db *DataStoreMongo) Migrate(ctx context.Context, version string) error {
	l := log.FromContext(ctx)

	dbs := []string{DbName}

	if db.multitenant {
		l.Infof("running migrations in multitenant mode")

		tdbs, err := migrate.GetTenantDbs(ctx, db.client, ctxstore.IsTenantDb(DbName))
		if err != nil {
			return errors.Wrap(err, "failed go retrieve tenant DBs")
		}
		dbs = append(dbs, tdbs...)
	} else {
		l.Infof("running migrations in single tenant mode")
	}

	if db.automigrate {
		l.Infof("automigrate is ON, will apply migrations")
	} else {
		l.Infof("automigrate is OFF, will check db version compatibility")
	}

	for _, d := range dbs {
		// if not in multi tenant, then tenant will be "" and identity
		// will be the same as default
		tenant := ctxstore.TenantFromDbName(d, DbName)

		tenantCtx := identity.WithContext(ctx, &identity.Identity{
			Tenant: tenant,
		})

		// TODO: we should aim to unify context usage across migrators and migrations
		// migrators use 'string' db name, migrations use DbFromContext
		// both should use one or the other; for now - just redundantly pass both ctx and db name
		err := db.MigrateTenant(tenantCtx, d, version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DataStoreMongo) MigrateTenant(ctx context.Context, database, version string) error {
	l := log.FromContext(ctx)

	l.Infof("migrating %s", database)

	m := migrate.SimpleMigrator{
		Client:      db.client,
		Db:          database,
		Automigrate: db.automigrate,
	}

	migrations := []migrate.Migration{
		&migration_1_1_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_2_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_3_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_4_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_5_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_6_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_7_0{
			ms:  db,
			ctx: ctx,
		},
		&migration_1_8_0{
			ds:  db,
			ctx: ctx,
		},
		&migration_1_9_0{
			ds:  db,
			ctx: ctx,
		},
		&migration_1_10_0{
			ds:  db,
			ctx: ctx,
		},
		&migration_1_11_0{
			ds:  db,
			ctx: ctx,
		},
		&migration_2_0_0{
			ds:  db,
			ctx: ctx,
		},
	}

	ver, err := migrate.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "failed to parse service version")
	}

	err = m.Apply(ctx, *ver, migrations)
	if err != nil {
		return errors.Wrap(err, "failed to apply migrations")
	}

	return nil
}

func fillAuthSet(ctx context.Context, authSet *model.AuthSet) {
	if authSet.Id == "" {
		uid := oid.NewUUIDv4()
		authSet.Id = uid.String()
	}
	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}
	authSet.TenantID = tenantId
	if authSet.Timestamp == nil {
		ts := time.Now()
		authSet.Timestamp = &ts
	}
}

func (db *DataStoreMongo) AddAuthSet(ctx context.Context, set model.AuthSet) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)
	fillAuthSet(ctx, &set)

	if _, err := c.InsertOne(ctx, set); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return store.ErrObjectExists
		}
		return errors.Wrap(err, "failed to store device")
	}

	return nil
}

// UpsertAuthSetStatus inserts a new auth set and if it exists, ensures the
// AuthSet.Status matches the one in authSet.
func (db *DataStoreMongo) UpsertAuthSetStatus(ctx context.Context, authSet *model.AuthSet) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	fillAuthSet(ctx, authSet)

	fltr := bson.D{
		{Key: dbFieldTenantID, Value: authSet.TenantID},
		{Key: dbFieldIDDataSha, Value: authSet.IdDataSha256},
		{Key: dbFieldPubKey, Value: authSet.PubKey},
	}
	setOnInsert := *authSet
	setOnInsert.Status = ""

	update := bson.D{
		{Key: "$setOnInsert", Value: setOnInsert},
		{Key: "$set", Value: bson.D{
			{Key: dbFieldStatus, Value: authSet.Status},
		}},
	}
	_, err := c.UpdateOne(ctx, fltr, update, mopts.Update().SetUpsert(true))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return store.ErrObjectExists
		}
		return errors.Wrap(err, "failed to store device")
	}

	return nil
}

func (db *DataStoreMongo) GetAuthSetByIdDataHashKey(
	ctx context.Context,
	idDataHash []byte,
	key string,
) (*model.AuthSet, error) {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}

	filter := model.AuthSet{
		IdDataSha256: idDataHash,
		PubKey:       key,
		TenantID:     tenantId,
	}
	res := model.AuthSet{}

	err := c.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrAuthSetNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch authentication set")
		}
	}

	return &res, nil
}

func (db *DataStoreMongo) GetAuthSetByIdDataHashKeyByStatus(
	ctx context.Context,
	idDataHash []byte,
	key string,
	status string,
) (*model.AuthSet, error) {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}

	filter := model.AuthSet{
		IdDataSha256: idDataHash,
		PubKey:       key,
		TenantID:     tenantId,
		Status:       status,
	}
	res := model.AuthSet{}

	err := c.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrAuthSetNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch authentication set")
		}
	}

	return &res, nil
}

func (db *DataStoreMongo) GetAuthSetById(
	ctx context.Context,
	auth_id string,
) (*model.AuthSet, error) {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	res := model.AuthSet{}
	err := c.FindOne(ctx, ctxstore.WithTenantID(ctx, bson.M{"_id": auth_id})).Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrAuthSetNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch authentication set")
		}
	}

	return &res, nil
}

func (db *DataStoreMongo) GetAuthSetsForDevice(
	ctx context.Context,
	devid string,
) ([]model.AuthSet, error) {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	res := []model.AuthSet{}

	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}
	cursor, err := c.Find(ctx, model.AuthSet{DeviceId: devid, TenantID: tenantId})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrAuthSetNotFound
		}
		return nil, errors.Wrap(err, "failed to fetch authentication sets")
	}

	return res, nil
}

func (db *DataStoreMongo) RejectAuthSetsForDevice(
	ctx context.Context,
	deviceID string,
) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)
	var tenantID string
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}

	filter := bson.M{
		model.AuthSetKeyDeviceId: deviceID,
		model.AuthSetKeyStatus:   model.DevStatusAccepted,
		dbFieldTenantID:          tenantID,
	}
	update := bson.M{
		"$set": bson.M{
			model.AuthSetKeyStatus: model.DevStatusRejected,
		},
	}

	if res, err := c.UpdateMany(ctx, filter, update); err != nil {
		return errors.Wrap(err, "failed to update auth set")
	} else if res.MatchedCount == 0 {
		return store.ErrAuthSetNotFound
	}

	return nil
}

func (db *DataStoreMongo) UpdateAuthSet(
	ctx context.Context,
	filter interface{},
	mod model.AuthSetUpdate,
) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	update := bson.M{"$set": mod}

	if res, err := c.UpdateMany(ctx, ctxstore.WithTenantID(ctx, filter), update); err != nil {
		return errors.Wrap(err, "failed to update auth set")
	} else if res.MatchedCount == 0 {
		return store.ErrAuthSetNotFound
	}

	return nil
}

func (db *DataStoreMongo) UpdateAuthSetById(
	ctx context.Context,
	id string,
	mod model.AuthSetUpdate,
) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)
	res, err := c.UpdateOne(ctx, ctxstore.WithTenantID(ctx, bson.M{"_id": id}), bson.M{"$set": mod})
	if err != nil {
		return errors.Wrap(err, "failed to update auth set")
	}
	if res.MatchedCount == 0 {
		return store.ErrAuthSetNotFound
	}

	return nil
}

func (db *DataStoreMongo) DeleteAuthSetsForDevice(ctx context.Context, devid string) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)
	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}

	ci, err := c.DeleteMany(
		ctx,
		model.AuthSet{
			DeviceId: devid,
			TenantID: tenantId,
		},
	)

	if err != nil {
		return errors.Wrap(err, "failed to remove authentication sets for device")
	}

	if ci.DeletedCount == 0 {
		return store.ErrAuthSetNotFound
	}

	return nil
}

func (db *DataStoreMongo) DeleteAuthSetForDevice(
	ctx context.Context,
	devId string,
	authId string,
) error {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}
	filter := model.AuthSet{Id: authId, DeviceId: devId, TenantID: tenantId}
	result, err := c.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "failed to remove authentication set for device")
	} else if result.DeletedCount < 1 {
		return store.ErrAuthSetNotFound
	}

	return nil
}

func (db *DataStoreMongo) WithMultitenant() *DataStoreMongo {
	db.multitenant = true
	return db
}

func (db *DataStoreMongo) WithAutomigrate() store.DataStore {
	db.automigrate = true
	return db
}

func (db *DataStoreMongo) PutLimit(ctx context.Context, lim model.Limit) error {
	if lim.Name == "" {
		return errors.New("empty limit name")
	}

	c := db.client.Database(DbName).Collection(DbLimitsColl)
	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}

	lim.TenantID = tenantId
	query := ctxstore.WithTenantID(ctx, bson.M{dbFieldName: lim.Name})

	updateOptions := mopts.Update()
	updateOptions.SetUpsert(true)
	if _, err := c.UpdateOne(
		ctx, query, bson.M{"$set": lim}, updateOptions); err != nil {
		return errors.Wrap(err, "failed to set or update limit")
	}

	return nil
}

func (db *DataStoreMongo) DeleteLimit(ctx context.Context, lim string) error {
	if lim == "" {
		return errors.New("empty limit name")
	}

	c := db.client.Database(DbName).Collection(DbLimitsColl)

	query := ctxstore.WithTenantID(ctx, bson.M{dbFieldName: lim})

	if _, err := c.DeleteOne(ctx, query); err != nil {
		return errors.Wrap(err, "failed to delete limit")
	}

	return nil
}

func (db *DataStoreMongo) GetLimit(ctx context.Context, name string) (*model.Limit, error) {
	c := db.client.Database(DbName).Collection(DbLimitsColl)

	var lim model.Limit

	err := c.FindOne(ctx, ctxstore.WithTenantID(ctx, bson.M{dbFieldName: name})).Decode(&lim)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrLimitNotFound
		} else {
			return nil, errors.Wrap(err, "failed to fetch limit")
		}
	}

	return &lim, nil
}

func (db *DataStoreMongo) GetDevCountByStatus(ctx context.Context, status string) (int, error) {
	var (
		fltr     = bson.D{}
		devsColl = db.client.
				Database(DbName).
				Collection(DbDevicesColl)
	)

	if status != "" {
		fltr = bson.D{{Key: "status", Value: status}}
	}
	count, err := devsColl.CountDocuments(ctx, ctxstore.WithTenantID(ctx, fltr))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (db *DataStoreMongo) GetDeviceStatus(ctx context.Context, devId string) (string, error) {
	id := identity.FromContext(ctx)
	tenantId := ""
	if id != nil {
		tenantId = id.Tenant
	}
	return getDeviceStatusDB(db, DbName, tenantId, ctx, devId)
}

func getDeviceStatus(statuses map[string]int) (string, error) {
	if statuses[model.DevStatusAccepted] > 1 || statuses[model.DevStatusPreauth] > 1 {
		return "", store.ErrDevStatusBroken
	}

	if statuses[model.DevStatusAccepted] == 1 {
		return model.DevStatusAccepted, nil
	}

	if statuses[model.DevStatusPreauth] == 1 {
		return model.DevStatusPreauth, nil
	}

	if statuses[model.DevStatusPending] > 0 {
		return model.DevStatusPending, nil
	}

	if statuses[model.DevStatusRejected] > 0 {
		return model.DevStatusRejected, nil
	}

	return "", store.ErrDevStatusBroken
}

func getDeviceStatusDB(
	db *DataStoreMongo,
	dbName string,
	tenantId string,
	ctx context.Context,
	devId string,
) (string, error) {
	var statuses = map[string]int{}

	c := db.client.Database(dbName).Collection(DbAuthSetColl)

	filter := model.AuthSet{
		DeviceId: devId,
	}
	if tenantId != "" {
		filter.TenantID = tenantId
	}

	match := bson.D{
		{Key: "$match", Value: filter},
	}
	group := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.M{"$sum": 1}}},
		},
	}

	pipeline := []bson.D{
		match,
		group,
	}
	var result []struct {
		Status string `bson:"_id"`
		Value  int    `bson:"count"`
	}
	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return "", err
	}
	if err := cursor.All(ctx, &result); err != nil {
		if err == mongo.ErrNoDocuments {
			return "", store.ErrAuthSetNotFound
		}
		return "", err
	}

	if len(result) == 0 {
		return "", store.ErrAuthSetNotFound
	}

	for _, res := range result {
		statuses[res.Status] = res.Value
	}

	status, err := getDeviceStatus(statuses)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (db *DataStoreMongo) ListTenantsIds(
	ctx context.Context,
) ([]string, error) {
	collDevs := db.client.
		Database(DbName).
		Collection(DbDevicesColl)

	results, err := collDevs.Distinct(ctx, dbFieldTenantID, bson.D{})
	if err != nil {
		return []string{}, nil
	}
	if len(results) < 1 {
		return []string{}, mongo.ErrNoDocuments
	}
	ids := make([]string, len(results))
	for i, id := range results {
		ids[i] = id.(string)
	}
	return ids, nil
}

func (db *DataStoreMongo) ListAllDevices(
	ctx context.Context,
	fields ...string,
) iter.Seq2[*model.Device, error] {
	collDevs := db.client.
		Database(DbName).
		Collection(DbDevicesColl,
			mopts.Collection().
				SetReadPreference(readpref.SecondaryPreferred()))

	findOpts := mopts.Find()
	if len(fields) > 0 {
		projection := make(bson.D, 0, len(fields))
		for _, key := range fields {
			projection = append(projection, bson.E{Key: key, Value: 1})
		}
		findOpts.SetProjection(projection)
	}

	cur, err := collDevs.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		return func(yield func(*model.Device, error) bool) {
			yield(nil, err)
		}
	}
	return func(yield func(*model.Device, error) bool) {
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var dev model.Device
			err := cur.Decode(&dev)
			if err != nil {
				if !yield(nil, err) {
					return
				}
			} else {
				if !yield(&dev, nil) {
					return
				}
			}
		}
	}
}
