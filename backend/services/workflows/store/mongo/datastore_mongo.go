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
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	dconfig "github.com/mendersoftware/mender-server/services/workflows/config"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store"
)

const (
	// JobsCollectionName refers to the collection of finished or
	// jobs in progress.
	JobsCollectionName = "jobs"

	// WorkflowCollectionName refers to the collection of stored workflows
	WorkflowCollectionName = "workflows"
)

var (
	ErrNoSuchWorkflowByVersion = errors.New("Workflow of the given version not found")
)

// SetupDataStore returns the mongo data store and optionally runs migrations
func SetupDataStore(automigrate bool) (*DataStoreMongo, error) {
	ctx := context.Background()
	dbClient, err := NewClient(ctx, config.Config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect to db: %v", err))
	}
	err = doMigrations(ctx, dbClient, automigrate)
	if err != nil {
		return nil, err
	}
	dataStore := NewDataStoreWithClient(dbClient, config.Config)
	return dataStore, nil
}

func doMigrations(ctx context.Context, client *mongo.Client,
	automigrate bool) error {
	db := config.Config.GetString(dconfig.SettingDbName)
	err := Migrate(ctx, db, DbVersion, client, automigrate)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to run migrations: %v", err))
	}

	return nil
}

func disconnectClient(parentCtx context.Context, client *mongo.Client) {
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	_ = client.Disconnect(ctx)
}

// NewClient returns a mongo client
func NewClient(_ context.Context, c config.Reader) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	mongoURL := c.GetString(dconfig.SettingMongo)
	if !strings.Contains(mongoURL, "://") {
		return nil, errors.Errorf("Invalid mongoURL %q: missing schema.",
			mongoURL)
	}
	clientOptions.ApplyURI(mongoURL).
		SetRegistry(newRegistry())

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

	// Set 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
	client *mongo.Client
	// dbName contains the name of the workflow database.
	dbName string
	// workflows holds a local cache of workflows - a worker should NEVER
	// access this cache directly, but through
	// DataStoreMongo.GetWorkflowByName.
	workflows map[string]*model.Workflow
}

// NewDataStoreWithClient initializes a DataStore object
func NewDataStoreWithClient(client *mongo.Client, c config.Reader) *DataStoreMongo {
	dbName := c.GetString(dconfig.SettingDbName)
	ctx := context.Background()

	// Maybe initialize workflows
	var findResults []*model.Workflow
	workflows := make(map[string]*model.Workflow)
	database := client.Database(dbName)
	collWflows := database.Collection(WorkflowCollectionName)
	cur, err := collWflows.Find(ctx, bson.M{})
	if err == nil {
		if err = cur.All(ctx, &findResults); err == nil {
			for _, workflow := range findResults {
				workflows[workflow.Name+"."+strconv.Itoa(workflow.Version)] = workflow
			}
		}
	}

	return &DataStoreMongo{
		client:    client,
		dbName:    dbName,
		workflows: workflows,
	}
}

func (db *DataStoreMongo) Ping(ctx context.Context) error {
	res := db.client.Database(db.dbName).
		RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

// LoadWorkflows from filesystem if the workflowsPath setting is provided
func (db *DataStoreMongo) LoadWorkflows(ctx context.Context, l *log.Logger) error {
	workflowsPath := config.Config.GetString(dconfig.SettingWorkflowsPath)
	if workflowsPath != "" {
		workflows := model.GetWorkflowsFromPath(workflowsPath)
		l.Infof("LoadWorkflows: loading %d workflows from %s.", len(workflows), workflowsPath)
		for _, workflow := range workflows {
			l.Infof("LoadWorkflows: loading %s v%d.", workflow.Name, workflow.Version)
			count, err := db.InsertWorkflows(ctx, *workflow)
			if count != 1 {
				l.Infof("LoadWorkflows: not loaded: %s v%d.", workflow.Name, workflow.Version)
			}
			if err != nil {
				l.Infof(
					"LoadWorkflows: error loading: %s v%d: %s.",
					workflow.Name,
					workflow.Version,
					err.Error(),
				)
			}
		}
	} else {
		l.Info("LoadWorkflows: empty workflowsPath, not loading workflows")
	}
	return nil
}

// InsertWorkflows inserts a workflow to the database and cache and returns the number of
// inserted elements or an error for the first error generated.
func (db *DataStoreMongo) InsertWorkflows(
	ctx context.Context,
	workflows ...model.Workflow,
) (int, error) {
	database := db.client.Database(db.dbName)
	collWflows := database.Collection(WorkflowCollectionName)
	for i, workflow := range workflows {
		if workflow.Name == "" {
			return i, store.ErrWorkflowMissingName
		}
		workflowDb, _ := db.GetWorkflowByName(ctx, workflow.Name, strconv.Itoa(workflow.Version))
		if workflowDb != nil && workflowDb.Version >= workflow.Version {
			return i + 1, store.ErrWorkflowAlreadyExists
		}
		if workflowDb == nil || workflowDb.Version < workflow.Version {
			upsert := true
			opt := &mopts.UpdateOptions{
				Upsert: &upsert,
			}
			query := bson.M{"_id": workflow.Name}
			update := bson.M{"$set": workflow}
			if _, err := collWflows.UpdateOne(ctx, query, update, opt); err != nil {
				return i + 1, err
			}
		}
		db.workflows[workflow.Name] = &workflow
	}
	return len(workflows), nil
}

// GetWorkflowByName gets the workflow with the given name - either from the
// cache, or searches the database if the workflow is not cached.
func (db *DataStoreMongo) GetWorkflowByName(
	ctx context.Context,
	workflowName string,
	version string,
) (*model.Workflow, error) {
	workflow, ok := db.workflows[workflowName]
	l := log.FromContext(ctx)

	versionNumber, e := strconv.Atoi(version)
	if e != nil {
		versionNumber = 0
	}

	if !ok || workflow.Version < versionNumber {
		var result model.Workflow
		database := db.client.Database(db.dbName)
		collWflows := database.Collection(WorkflowCollectionName)
		err := collWflows.FindOne(ctx, bson.M{"_id": workflowName}).
			Decode(&result)
		if err != nil {
			return nil, err
		}
		if result.Version < versionNumber {
			l.Errorf("workflow found but with version too low: %s v%s", workflowName, version)
			return nil, ErrNoSuchWorkflowByVersion
		}
		db.workflows[result.Name] = &result
		return &result, err
	} else {
		l.Debugf("cache hit: %s v%s", workflowName, version)
	}
	return workflow, nil
}

// GetWorkflows gets all workflows from the cache as a list
// (should only be used by the server process)
func (db *DataStoreMongo) GetWorkflows(ctx context.Context) []model.Workflow {
	workflows := make([]model.Workflow, len(db.workflows))
	var i int
	for _, workflow := range db.workflows {
		workflows[i] = *workflow
		i++
	}

	return workflows
}

// UpsertJob inserts the job in the queue
func (db *DataStoreMongo) UpsertJob(
	ctx context.Context, job *model.Job) (*model.Job, error) {
	if job.ID == "" {
		job.ID = primitive.NewObjectID().Hex()
	}
	query := bson.M{
		"_id": job.ID,
	}
	update := bson.M{
		"$set": job,
	}
	findUpdateOptions := &mopts.FindOneAndUpdateOptions{}
	findUpdateOptions.SetReturnDocument(mopts.After)
	findUpdateOptions.SetUpsert(true)

	database := db.client.Database(db.dbName)
	collJobs := database.Collection(JobsCollectionName)

	err := collJobs.FindOneAndUpdate(ctx, query, update, findUpdateOptions).Decode(job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// UpdateJobAddResult add a task execution result to a job status
func (db *DataStoreMongo) UpdateJobAddResult(ctx context.Context,
	job *model.Job, result *model.TaskResult) error {
	options := &mopts.UpdateOptions{}
	options.SetUpsert(true)

	update := bson.M{
		"$addToSet": bson.M{
			"results": result,
		},
		"$setOnInsert": bson.M{
			"workflow_name":    job.WorkflowName,
			"input_parameters": job.InputParameters,
			"status":           job.Status,
			"insert_time":      job.InsertTime,
			"version":          job.WorkflowVersion,
		},
	}

	collection := db.client.Database(db.dbName).
		Collection(JobsCollectionName)
	_, err := collection.UpdateOne(ctx, bson.M{"_id": job.ID}, update, options)
	if err != nil {
		return err
	}

	return nil
}

// UpdateJobStatus set the task execution status for a job status
func (db *DataStoreMongo) UpdateJobStatus(
	ctx context.Context, job *model.Job, status int32) error {
	if model.StatusToString(status) == "unknown" {
		return model.ErrInvalidStatus
	}

	options := &mopts.UpdateOptions{}
	options.SetUpsert(true)

	collection := db.client.Database(db.dbName).
		Collection(JobsCollectionName)
	_, err := collection.UpdateOne(ctx, bson.M{
		"_id": job.ID,
	}, bson.M{
		"$set": bson.M{
			"status": status,
		},
		"$setOnInsert": bson.M{
			"workflow_name":    job.WorkflowName,
			"input_parameters": job.InputParameters,
			"results":          job.Results,
			"insert_time":      job.InsertTime,
			"version":          job.WorkflowVersion,
		},
	}, options)
	if err != nil {
		return err
	}

	return nil
}

// GetJobByNameAndID get the task execution status for a job
// by workflow name and ID
func (db *DataStoreMongo) GetJobByNameAndID(
	ctx context.Context, name string, ID string) (*model.Job, error) {
	collection := db.client.Database(db.dbName).
		Collection(JobsCollectionName)
	cur := collection.FindOne(ctx, bson.M{
		"_id":           ID,
		"workflow_name": name,
	})
	var job model.Job
	err := cur.Decode(&job)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

// GetJobByID get the task execution status for a job by ID
func (db *DataStoreMongo) GetJobByID(
	ctx context.Context, ID string) (*model.Job, error) {
	collection := db.client.Database(db.dbName).
		Collection(JobsCollectionName)
	cur := collection.FindOne(ctx, bson.M{
		"_id": ID,
	})
	var job model.Job
	err := cur.Decode(&job)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (db *DataStoreMongo) GetAllJobs(
	ctx context.Context, page int64, perPage int64) ([]model.Job, int64, error) {
	collection := db.client.Database(db.dbName).
		Collection(JobsCollectionName)
	findOptions := &mopts.FindOptions{}
	findOptions.SetSkip(int64((page - 1) * perPage))
	findOptions.SetLimit(int64(perPage))
	sortField := bson.M{}
	sortField["insert_time"] = -1
	findOptions.SetSort(sortField)
	cur, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return []model.Job{}, 0, err
	}

	var jobs []model.Job
	err = cur.All(ctx, &jobs)
	if err == mongo.ErrNoDocuments {
		return []model.Job{}, 0, nil
	} else if err != nil {
		return []model.Job{}, 0, err
	}

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return []model.Job{}, 0, err
	}
	return jobs, count, nil
}

// Close disconnects the client
func (db *DataStoreMongo) Close() {
	ctx := context.Background()
	disconnectClient(ctx, db.client)
}
