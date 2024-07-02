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

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/model"
)

// Error messages
var (
	ErrWorkflowNotFound      = errors.New("Workflow not found")
	ErrWorkflowMissingName   = errors.New("Workflow missing name")
	ErrWorkflowAlreadyExists = errors.New("Workflow already exists")
)

// DataStore interface for DataStore services
type DataStore interface {
	Ping(ctx context.Context) error
	InsertWorkflows(ctx context.Context, workflow ...model.Workflow) (int, error)
	GetWorkflowByName(
		ctx context.Context,
		workflowName string,
		version string,
	) (*model.Workflow, error)
	GetWorkflows(ctx context.Context) []model.Workflow
	LoadWorkflows(ctx context.Context, l *log.Logger) error
	UpsertJob(ctx context.Context, job *model.Job) (*model.Job, error)
	GetAllJobs(ctx context.Context, page int64, perPage int64) ([]model.Job, int64, error)
	UpdateJobAddResult(ctx context.Context, job *model.Job, result *model.TaskResult) error
	UpdateJobStatus(ctx context.Context, job *model.Job, status int32) error
	GetJobByNameAndID(ctx context.Context, name string, ID string) (*model.Job, error)
	GetJobByID(ctx context.Context, ID string) (*model.Job, error)
}
