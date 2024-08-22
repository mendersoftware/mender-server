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

package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/model"
)

// DataStore is a mocked data storage service
type DataStore struct {
	mock.Mock
}

// NewDataStore initializes a DataStore mock object
func NewDataStore() *DataStore {
	return &DataStore{}
}

func (db *DataStore) Ping(ctx context.Context) error {
	ret := db.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// LoadWorkflows from filesystem if the workflowsPath setting is provided
func (db *DataStore) LoadWorkflows(ctx context.Context, l *log.Logger) error {
	ret := db.Called(ctx, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *log.Logger) error); ok {
		r0 = rf(ctx, l)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// InsertWorkflows inserts one or multiple workflows
func (db *DataStore) InsertWorkflows(ctx context.Context,
	workflows ...model.Workflow) (int, error) {
	ret := db.Called(ctx, workflows)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, []model.Workflow) int); ok {
		r0 = rf(ctx, workflows)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(int)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []model.Workflow) error); ok {
		r1 = rf(ctx, workflows)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// GetWorkflowByName returns a workflow by name
func (db *DataStore) GetWorkflowByName(ctx context.Context,
	workflowName string, version string) (*model.Workflow, error) {
	ret := db.Called(ctx, workflowName, version)

	var r0 *model.Workflow
	if rf, ok := ret.
		Get(0).(func(context.Context, string, string) *model.Workflow); ok {
		r0 = rf(ctx, workflowName, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Workflow)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, workflowName, version)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// GetWorkflows returns the list of workflows
func (db *DataStore) GetWorkflows(ctx context.Context) []model.Workflow {
	ret := db.Called(ctx)

	var r0 []model.Workflow
	if rf, ok := ret.Get(0).(func() []model.Workflow); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Workflow)
		}
	}

	return r0
}

// UpsertJob inserts the job in the queue
func (db *DataStore) UpsertJob(ctx context.Context, job *model.Job) (*model.Job, error) {
	ret := db.Called(ctx, job)

	var r0 *model.Job
	if rf, ok := ret.
		Get(0).(func(context.Context, *model.Job) *model.Job); ok {
		r0 = rf(ctx, job)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.Job) error); ok {
		r1 = rf(ctx, job)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateJobAddResult add a task execution result to a job status
func (db *DataStore) UpdateJobAddResult(ctx context.Context,
	job *model.Job, result *model.TaskResult) error {
	ret := db.Called(ctx, job, result)

	var r0 error
	if rf, ok := ret.Get(0).(func(
		context.Context, *model.Job, *model.TaskResult) error); ok {
		r0 = rf(ctx, job, result)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateJobStatus set the task execution status for a job status
func (db *DataStore) UpdateJobStatus(ctx context.Context, job *model.Job,
	status int32) error {
	ret := db.Called(ctx, job, status)

	var r0 error
	if rf, ok := ret.Get(0).(func(
		context.Context, *model.Job, int32) error); ok {
		r0 = rf(ctx, job, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetJobByNameAndID get the task execution status for a job status bu Name and ID
func (db *DataStore) GetJobByNameAndID(ctx context.Context,
	name, ID string) (*model.Job, error) {
	ret := db.Called(ctx, name, ID)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(
		context.Context, string, string) *model.Job); ok {
		r0 = rf(ctx, name, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(
		context.Context, string, string) error); ok {
		r1 = rf(ctx, name, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetJobByID get a job by ID
func (db *DataStore) GetJobByID(ctx context.Context,
	ID string) (*model.Job, error) {
	ret := db.Called(ctx, ID)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(
		context.Context, string) *model.Job); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(
		context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (db *DataStore) GetAllJobs(
	ctx context.Context,
	page int64,
	perPage int64,
) ([]model.Job, int64, error) {
	ret := db.Called(ctx, page, perPage)

	var r0 []model.Job
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) []model.Job); ok {
		r0 = rf(ctx, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Job)
		}
	}

	var r1 int64
	if rf, ok := ret.Get(1).(func(
		context.Context, int64, int64) int64); ok {
		r1 = rf(ctx, page, perPage)
	} else {
		r1 = ret.Get(1).(int64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(
		context.Context, int64, int64) error); ok {
		r2 = rf(ctx, page, perPage)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
