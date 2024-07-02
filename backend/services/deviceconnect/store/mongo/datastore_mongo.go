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
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	mdoc "github.com/mendersoftware/mender-server/pkg/mongo/doc"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store/v2"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	dconfig "github.com/mendersoftware/mender-server/services/deviceconnect/config"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
	"github.com/mendersoftware/mender-server/services/deviceconnect/store"
	"github.com/mendersoftware/mender-server/services/deviceconnect/utils"
)

var (
	clock                        utils.Clock = utils.RealClock{}
	recordingReadBufferSize                  = 1024
	ErrUnknownControlMessageType             = errors.New("unknown control message type")
	ErrRecordingDataInconsistent             = errors.New("recording data corrupt")
)

const (
	// DevicesCollectionName refers to the name of the collection of stored devices
	DevicesCollectionName = "devices"

	// SessionsCollectionName refers to the name of the collection of sessions
	SessionsCollectionName = "sessions"

	// RecordingsCollectionName name of the collection of session recordings
	RecordingsCollectionName = "recordings"

	// ControlCollectionName name of the collection of session control data
	ControlCollectionName = "control"

	dbFieldID        = "_id"
	dbFieldVersion   = "version"
	dbFieldSessionID = "session_id"
	dbFieldDeviceID  = "device_id"
	dbFieldStatus    = "status"
	dbFieldCreatedTs = "created_ts"
	dbFieldUpdatedTs = "updated_ts"
)

// SetupDataStore returns the mongo data store and optionally runs migrations
func SetupDataStore(automigrate bool) (store.DataStore, error) {
	ctx := context.Background()
	dbClient, err := NewClient(ctx, config.Config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect to db: %v", err))
	}
	err = doMigrations(ctx, dbClient, automigrate)
	if err != nil {
		return nil, err
	}
	dataStore := NewDataStoreWithClient(dbClient,
		time.Second*time.Duration(config.Config.GetInt(dconfig.SettingRecordingExpireSec)))
	return dataStore, nil
}

func doMigrations(ctx context.Context, client *mongo.Client,
	automigrate bool) error {
	db := config.Config.GetString(dconfig.SettingDbName)
	dbs, err := migrate.GetTenantDbs(ctx, client, mstore.IsTenantDb(db))
	if err != nil {
		return errors.Wrap(err, "failed go retrieve tenant DBs")
	}
	if len(dbs) == 0 {
		dbs = []string{DbName}
	}

	for _, d := range dbs {
		err := Migrate(ctx, d, DbVersion, client, automigrate)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to run migrations: %v", err))
		}
	}
	return nil
}

// NewClient returns a mongo client
func NewClient(ctx context.Context, c config.Reader) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	mongoURL := c.GetString(dconfig.SettingMongo)
	if !strings.Contains(mongoURL, "://") {
		return nil, errors.Errorf("Invalid mongoURL %q: missing schema.",
			mongoURL)
	}
	clientOptions.ApplyURI(mongoURL)

	username := c.GetString(dconfig.SettingDbUsername)
	if username != "" {
		credentials := mopts.Credential{
			Username: c.GetString(dconfig.SettingDbUsername),
		}
		password := c.GetString(dconfig.SettingDbPassword)
		if password != "" {
			credentials.Password = password
			credentials.PasswordSet = true
		}
		clientOptions.SetAuth(credentials)
	}

	if c.GetBool(dconfig.SettingDbSSL) {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = c.GetBool(dconfig.SettingDbSSLSkipVerify)
		clientOptions.SetTLSConfig(tlsConfig)
	}

	// Set writeconcern to acknowlage after write has propagated to the
	// mongod instance and commited to the file system journal.
	var wc *writeconcern.WriteConcern
	journal := true
	wc = &writeconcern.WriteConcern{
		W:       1,
		Journal: &journal,
	}
	clientOptions.SetWriteConcern(wc)

	// Set 10s timeout
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to mongo server")
	}

	// Validate connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "Error reaching mongo server")
	}

	return client, nil
}

// DataStoreMongo is the data storage service
type DataStoreMongo struct {
	// client holds the reference to the client used to communicate with the
	// mongodb server.
	client          *mongo.Client
	recordingExpire time.Duration
}

// NewDataStoreWithClient initializes a DataStore object
func NewDataStoreWithClient(client *mongo.Client, expire time.Duration) store.DataStore {
	return &DataStoreMongo{
		client:          client,
		recordingExpire: expire,
	}
}

// Ping verifies the connection to the database
func (db *DataStoreMongo) Ping(ctx context.Context) error {
	res := db.client.Database(DbName).RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

// ProvisionDevice provisions a new device
func (db *DataStoreMongo) ProvisionDevice(ctx context.Context, tenantID, deviceID string) error {
	coll := db.client.Database(DbName).Collection(DevicesCollectionName)

	now := clock.Now().UTC()

	updateOpts := &mopts.UpdateOptions{}
	updateOpts.SetUpsert(true)
	_, err := coll.UpdateOne(ctx,
		bson.M{dbFieldID: deviceID, mstore.FieldTenantID: tenantID},
		bson.M{
			"$setOnInsert": bson.M{
				dbFieldStatus:        model.DeviceStatusUnknown,
				dbFieldCreatedTs:     &now,
				dbFieldUpdatedTs:     &now,
				mstore.FieldTenantID: tenantID,
			},
		},
		updateOpts,
	)
	return err
}

// DeleteDevice deletes a device
func (db *DataStoreMongo) DeleteDevice(ctx context.Context, tenantID, deviceID string) error {
	coll := db.client.Database(DbName).Collection(DevicesCollectionName)

	_, err := coll.DeleteOne(ctx, bson.M{dbFieldID: deviceID, mstore.FieldTenantID: tenantID})
	return err
}

// GetDevice returns a device
func (db *DataStoreMongo) GetDevice(
	ctx context.Context,
	tenantID string,
	deviceID string,
) (*model.Device, error) {
	coll := db.client.Database(DbName).Collection(DevicesCollectionName)

	cur := coll.FindOne(ctx, bson.M{dbFieldID: deviceID, mstore.FieldTenantID: tenantID})

	device := &model.Device{}
	if err := cur.Decode(&device); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return device, nil
}

func (db *DataStoreMongo) SetDeviceConnected(
	ctx context.Context,
	tenantID, deviceID string,
) (int64, error) {
	coll := db.client.Database(DbName).Collection(DevicesCollectionName)

	updateOpts := mopts.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(mopts.After).
		SetProjection(bson.M{"version": 1})

	now := clock.Now().UTC()

	var version struct {
		Version int64 `bson:"version"`
	}

	err := coll.FindOneAndUpdate(ctx,
		bson.M{dbFieldID: deviceID, mstore.FieldTenantID: tenantID},
		bson.M{
			"$set": bson.M{
				dbFieldStatus:    model.DeviceStatusConnected,
				dbFieldUpdatedTs: &now,
			},
			"$inc": bson.M{"version": 1},
			"$setOnInsert": bson.M{
				dbFieldCreatedTs:     &now,
				mstore.FieldTenantID: tenantID,
			},
		},
		updateOpts,
	).Decode(&version)

	return version.Version, err
}
func (db *DataStoreMongo) SetDeviceDisconnected(
	ctx context.Context,
	tenantID, deviceID string,
	version int64,
) error {
	coll := db.client.Database(DbName).Collection(DevicesCollectionName)

	now := clock.Now().UTC()

	_, err := coll.UpdateOne(ctx,
		bson.M{
			dbFieldID:            deviceID,
			mstore.FieldTenantID: tenantID,
			dbFieldVersion:       version,
		},
		bson.M{
			"$set": bson.M{
				dbFieldStatus:    model.DeviceStatusDisconnected,
				dbFieldUpdatedTs: &now,
			},
			"$setOnInsert": bson.M{
				dbFieldCreatedTs:     &now,
				mstore.FieldTenantID: tenantID,
			},
		},
	)
	return err
}

// AllocateSession allocates a new session.
func (db *DataStoreMongo) AllocateSession(ctx context.Context, sess *model.Session) error {

	if err := sess.Validate(); err != nil {
		return errors.Wrap(err, "store: cannot allocate invalid Session")
	}

	coll := db.client.Database(DbName).Collection(SessionsCollectionName)
	tenantElem := bson.E{Key: mstore.FieldTenantID, Value: sess.TenantID}
	_, err := coll.InsertOne(ctx, mdoc.DocumentFromStruct(*sess, tenantElem))
	if err != nil {
		return errors.Wrap(err, "store: failed to allocate session")
	}

	return nil
}

// DeleteSession deletes a session
func (db *DataStoreMongo) DeleteSession(
	ctx context.Context, sessionID string,
) (*model.Session, error) {
	collSess := db.client.Database(DbName).
		Collection(SessionsCollectionName)

	sess := new(model.Session)
	err := collSess.FindOneAndDelete(
		ctx, mstore.WithTenantID(ctx, bson.D{{Key: dbFieldID, Value: sessionID}}),
	).Decode(sess)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrSessionNotFound
		} else {
			return nil, err
		}
	}
	if idty := identity.FromContext(ctx); idty != nil {
		sess.TenantID = idty.Tenant
	}
	return sess, nil
}

// GetSession returns a session
func (db *DataStoreMongo) GetSession(
	ctx context.Context,
	sessionID string,
) (*model.Session, error) {
	collSess := db.client.
		Database(DbName).
		Collection(SessionsCollectionName)

	session := &model.Session{}
	err := collSess.
		FindOne(ctx, mstore.WithTenantID(ctx, bson.M{dbFieldID: sessionID})).
		Decode(session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrSessionNotFound
		}
		return nil, err
	}
	idty := identity.FromContext(ctx)
	if idty != nil {
		session.TenantID = idty.Tenant
	}

	return session, nil
}

func sendControlMessage(control app.Control, sessionID string, w io.Writer) (int, error) {
	messageType := ""
	var data []byte
	properties := make(map[string]interface{})

	switch control.Type {
	case app.DelayMessage:
		messageType = model.DelayMessageName
		properties[model.DelayMessageValueField] = control.DelayMs
	case app.ResizeMessage:
		messageType = shell.MessageTypeResizeShell
		properties[model.ResizeMessageTermHeightField] = control.TerminalHeight
		properties[model.ResizeMessageTermWidthField] = control.TerminalWidth
	default:
		return 0, ErrUnknownControlMessageType
	}

	msg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:      ws.ProtoTypeShell,
			MsgType:    messageType,
			SessionID:  sessionID,
			Properties: properties,
		},
		Body: data,
	}
	messagePacked, err := msgpack.Marshal(&msg)
	if err != nil {
		return 0, err
	} else {
		return w.Write(messagePacked)
	}
}

// GetSession writes session recordings to given io.Writer
func (db *DataStoreMongo) WriteSessionRecords(ctx context.Context,
	sessionID string,
	w io.Writer) error {
	l := log.FromContext(ctx)
	collRecording := db.client.Database(DbName).
		Collection(RecordingsCollectionName)
	collControl := db.client.Database(DbName).
		Collection(ControlCollectionName)

	findOptions := mopts.Find()
	sortField := bson.M{
		"created_ts": 1,
	}
	findOptions.SetSort(sortField)
	recordingsCursor, err := collRecording.Find(ctx,
		mstore.WithTenantID(ctx, bson.M{
			dbFieldSessionID: sessionID,
		}),
		findOptions,
	)
	if err != nil {
		return err
	}
	defer recordingsCursor.Close(ctx)

	controlCursor, err := collControl.Find(ctx,
		mstore.WithTenantID(ctx, bson.M{
			dbFieldSessionID: sessionID,
		}),
		findOptions,
	)
	if err != nil {
		return err
	}
	defer controlCursor.Close(ctx)

	controlReader := NewControlMessageReader(ctx, controlCursor)
	recordingReader := NewRecordingReader(ctx, recordingsCursor)

	recordingWriter := NewRecordingWriter(sessionID, w)
	recordingBuffer := make([]byte, recordingReadBufferSize)
	recordingBytesSent := 0
	for {
		control := controlReader.Pop()
		if control == nil {
			l.Debug("WriteSessionRecords: no more control " +
				"messages, flushing the recording upstream.")
			//no more control messages, we send the whole recording
			n, err := io.Copy(recordingWriter, recordingReader)
			if err != nil && err != io.ErrShortWrite && n < 1 {
				l.Errorf("WriteSessionRecords: "+
					"error writing recording data, err: %+v n:%d",
					err, n)
			}
			if n == 0 {
				l.Errorf("WriteSessionRecords: "+
					"failed to write any recording data, err: %+v",
					err)
			} else {
				recordingBytesSent += int(n)
			}
			break
		} else {
			if recordingBytesSent > control.Offset {
				//this should never happen, we missed
				//the control message, data inconsistency
				l.Errorf("WriteSessionRecords: recordingBytesSent > control.Offset")
				err = ErrRecordingDataInconsistent
				break
			}

			bytesUntilControlMessage := control.Offset - recordingBytesSent
			l.Debugf("(1) WriteSessionRecords: control.Offset:%d"+
				" recordingBytesSent:%d "+
				" bytesUntilControlMessage:%d (recordingBuffer.len=%d) "+
				"reading up to %d bytes of recording and sending.",
				control.Offset,
				recordingBytesSent,
				bytesUntilControlMessage,
				len(recordingBuffer),
				control.Offset-recordingBytesSent)
			//it is possible that the recording is larger than one recordingBuffer,
			//we need to send until we have the control.Offset in the buffer
			for bytesUntilControlMessage > len(recordingBuffer) {
				n, e := recordingReader.Read(recordingBuffer)
				if n > 0 {
					_, err = sendRecordingMessage(recordingBuffer[:n],
						sessionID,
						w)
					if err != nil {
						l.Errorf("error sending recording data: %s",
							err.Error())
						break
					}
					recordingBytesSent += n
					bytesUntilControlMessage = control.Offset -
						recordingBytesSent
				}
				if e != nil || n == 0 {
					break
				}
			}
			if err != nil {
				break
			}

			bytesUntilControlMessage = control.Offset - recordingBytesSent
			l.Debugf("(2) WriteSessionRecords: control.Offset:%d"+
				" recordingBytesSent:%d "+
				" bytesUntilControlMessage:%d (recordingBuffer.len=%d) "+
				"reading up to %d bytes of recording and sending.",
				control.Offset,
				recordingBytesSent,
				bytesUntilControlMessage,
				len(recordingBuffer),
				bytesUntilControlMessage)
			//this means that the control offset is in the future
			//part of the recording buffer
			//we can send up to control.Offset-recordingBytesSent
			//bytes and then send the control message
			n, e := recordingReader.Read(recordingBuffer[:bytesUntilControlMessage])
			l.Debugf("recordingReader.Read(len=%d)=%d,%+v",
				bytesUntilControlMessage, n, e)
			if n > 0 {
				_, err = sendRecordingMessage(recordingBuffer[:n], sessionID, w)
				if err != nil {
					l.Errorf("error sending recording data: %s",
						err.Error())
					break
				}
				recordingBytesSent += n
			}
			l.Debugf("WriteSessionRecords: sending %+v.", *control)
			_, err = sendControlMessage(*control, sessionID, w)
			if err != nil {
				l.Errorf("error sending recording data: %s",
					err.Error())
				break
			}
		}
	}
	l.Infof("session playback: WriteSessionRecords: sent %d bytes.", recordingBytesSent)

	return err
}

// SetSession saves a session recording
func (db *DataStoreMongo) InsertSessionRecording(ctx context.Context,
	sessionID string,
	sessionBytes []byte) error {
	coll := db.client.Database(DbName).Collection(RecordingsCollectionName)

	now := clock.Now().UTC()
	recording := model.Recording{
		ID:        uuid.New(),
		SessionID: sessionID,
		Recording: sessionBytes,
		CreatedTs: now,
		ExpireTs:  now.Add(db.recordingExpire),
	}
	_, err := coll.InsertOne(ctx,
		mstore.WithTenantID(ctx, &recording),
	)
	return err
}

// Inserts control data recording
func (db *DataStoreMongo) InsertControlRecording(ctx context.Context,
	sessionID string,
	sessionBytes []byte) error {
	coll := db.client.Database(DbName).
		Collection(ControlCollectionName)

	now := clock.Now().UTC()
	recording := model.ControlData{
		ID:        uuid.New(),
		SessionID: sessionID,
		Control:   sessionBytes,
		CreatedTs: now,
		ExpireTs:  now.Add(db.recordingExpire),
	}
	_, err := coll.InsertOne(ctx,
		mstore.WithTenantID(ctx, &recording),
	)
	return err
}

// Close disconnects the client
func (db *DataStoreMongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := db.client.Disconnect(ctx)
	return err
}

//nolint:unused
func (db *DataStoreMongo) DropDatabase() error {
	ctx := context.Background()
	err := db.client.Database(DbName).Drop(ctx)
	return err
}
