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

package mongo

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	dconfig "github.com/mendersoftware/mender-server/services/workflows/config"
	"github.com/mendersoftware/mender-server/services/workflows/model"
)

func TestInsertWorkflows(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	workflow := model.Workflow{
		Name:          "TestInsertWorkflows",
		Description:   "Description",
		SchemaVersion: 1,
		Version:       1,
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
		InputParameters: []string{
			"key1",
			"key2",
		},
	}

	ctx := context.Background()
	count, _ := testDataStore.InsertWorkflows(ctx, workflow, workflow)
	assert.Equal(t, 2, count)

	workflowFromDb, err := testDataStore.GetWorkflowByName(ctx, workflow.Name, strconv.Itoa(workflow.Version))
	assert.Nil(t, err)
	assert.Equal(t, workflow.Name, workflowFromDb.Name)
	assert.Equal(t, workflow.Description, workflowFromDb.Description)
	assert.Equal(t, workflow.SchemaVersion, workflowFromDb.SchemaVersion)
	assert.Equal(t, workflow.Version, workflowFromDb.Version)
	assert.Equal(t, workflow.InputParameters, workflowFromDb.InputParameters)

	workflows := testDataStore.GetWorkflows(ctx)
	assert.Len(t, workflows, 1)
	assert.Equal(t, workflow.Name, workflows[0].Name)
	assert.Equal(t, workflow.Description, workflows[0].Description)
	assert.Equal(t, workflow.SchemaVersion, workflows[0].SchemaVersion)
	assert.Equal(t, workflow.Version, workflows[0].Version)
	assert.Equal(t, workflow.InputParameters, workflows[0].InputParameters)

	count, err = testDataStore.InsertWorkflows(ctx, workflow)
	assert.NotNil(t, err)
	assert.Equal(t, 1, count)

	testDataStore.workflows = make(map[string]*model.Workflow)
	count, err = testDataStore.InsertWorkflows(ctx, workflow)
	assert.NotNil(t, err)
	assert.Equal(t, 1, count)

	workflowFromDb, err = testDataStore.GetWorkflowByName(ctx, workflow.Name, strconv.Itoa(workflow.Version))
	assert.Nil(t, err)
	assert.Equal(t, workflow.Name, workflowFromDb.Name)
	assert.Equal(t, workflow.Description, workflowFromDb.Description)
	assert.Equal(t, workflow.SchemaVersion, workflowFromDb.SchemaVersion)
	assert.Equal(t, workflow.Version, workflowFromDb.Version)
	assert.Equal(t, workflow.InputParameters, workflowFromDb.InputParameters)
}

func TestLoadWorkflows(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	data := []byte(`
{
	"name": "decommission_device_from_filesystem",
	"description": "Removes device info from all services.",
	"version": 4,
	"tasks": [
		{
			"name": "delete_device_inventory",
			"type": "http",
			"http": {
				"uri": "http://mender-inventory:8080/api/management/v1/inventory/devices/${workflow.input.device_id}",
				"method": "DELETE",
				"body": "Payload",
				"headers": {
					"X-MEN-RequestID": "${workflow.input.request_id}",
					"Authorization": "${workflow.input.authorization}"
				},
				"connectionTimeOut": 1000,
				"readTimeOut": 1000
			}
		}
	],
	"inputParameters": [
		"device_id",
		"request_id",
		"authorization"
	],
	"schemaVersion": 1
}`)

	dir, _ := ioutil.TempDir("", "example")
	defer os.RemoveAll(dir) // clean up

	config.Config.Set(dconfig.SettingWorkflowsPath, dir)

	tmpfn := filepath.Join(dir, "test.json")
	ioutil.WriteFile(tmpfn, data, 0666)

	ctx := context.Background()
	err := testDataStore.LoadWorkflows(ctx, log.FromContext(ctx))
	assert.Nil(t, err)

	workflowFromDb, err := testDataStore.GetWorkflowByName(ctx, "decommission_device_from_filesystem", "4")
	assert.Nil(t, err)
	assert.Equal(t, "decommission_device_from_filesystem", workflowFromDb.Name)
	assert.Equal(t, "Removes device info from all services.", workflowFromDb.Description)
	assert.Equal(t, 1, workflowFromDb.SchemaVersion)
	assert.Equal(t, 4, workflowFromDb.Version)
}

func TestInsertWorkflowsMissingName(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	workflow := model.Workflow{
		Name:          "",
		Description:   "Description",
		SchemaVersion: 1,
		Version:       1,
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
		InputParameters: []string{
			"key1",
			"key2",
		},
	}

	ctx := context.Background()
	count, err := testDataStore.InsertWorkflows(ctx, workflow)
	assert.NotNil(t, err)
	assert.Equal(t, 0, count)
}

func TestGetWorkflowByName(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	workflowFromDb, err := testDataStore.GetWorkflowByName(ctx, "TestGetWorkflowByName", "1")
	assert.NotNil(t, err)
	assert.Nil(t, workflowFromDb)

	testWorkflowname := "TestGetWorkflowByName1"
	testWorkflow := &model.Workflow{
		Name:        testWorkflowname,
		Description: "The workflows especially for TestGetWorkflowByName",
		Version:     2,
	}

	testDataStore.workflows[testWorkflowname] = testWorkflow
	workflowFromDb, err = testDataStore.GetWorkflowByName(ctx, testWorkflowname, "10")
	assert.NotNil(t, err)
	assert.Nil(t, workflowFromDb)

	testDataStore.workflows[testWorkflowname] = testWorkflow
	workflowFromDb, err = testDataStore.GetWorkflowByName(ctx, testWorkflowname, "2")
	assert.Nil(t, err)
	assert.NotNil(t, workflowFromDb)
	assert.Equal(t, workflowFromDb, testWorkflow)

	testWorkflowname = testWorkflowname + "2"
	testWorkflow.Name = testWorkflowname
	testWorkflow.Version = 10
	c, err := testDataStore.InsertWorkflows(ctx, []model.Workflow{*testWorkflow}...)
	assert.Equal(t, c, 1)
	assert.Nil(t, err)
	testDataStore.workflows[testWorkflowname] = testWorkflow
	workflowFromDb, err = testDataStore.GetWorkflowByName(ctx, testWorkflowname, "10")
	assert.Nil(t, err)
	assert.Equal(t, workflowFromDb, testWorkflow)
}

func TestUpsertJob(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	job := &model.Job{
		WorkflowName: "test",
		Status:       model.StatusPending,
		InputParameters: []model.InputParameter{
			{
				Name:  "key1",
				Value: "value1",
			},
			{
				Name:  "key2",
				Value: "value2",
			},
		},
	}

	inserted, err := testDataStore.UpsertJob(ctx, job)
	assert.Nil(t, err)
	assert.NotNil(t, inserted)
	assert.Equal(t, job.WorkflowName, inserted.WorkflowName)
	assert.Equal(t, job.InputParameters, inserted.InputParameters)
	assert.Equal(t, model.StatusPending, inserted.Status)
}

func TestUpdateJobStatus(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	workflow := model.Workflow{
		Name:          "TestUpdateJobStatus",
		Description:   "Description",
		SchemaVersion: 1,
		Version:       1,
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
		InputParameters: []string{
			"key1",
			"key2",
		},
	}

	ctx := context.Background()
	count, _ := testDataStore.InsertWorkflows(ctx, workflow)
	assert.Equal(t, 1, count)

	job := &model.Job{
		WorkflowName: workflow.Name,
		InputParameters: []model.InputParameter{
			{
				Name:  "key1",
				Value: "value1",
			},
			{
				Name:  "key2",
				Value: "value2",
			},
		},
	}

	inserted, err := testDataStore.UpsertJob(ctx, job)
	assert.Nil(t, err)

	err = testDataStore.UpdateJobStatus(ctx, inserted, model.StatusFailure)
	assert.Nil(t, err)

	jobFromDb, err := testDataStore.GetJobByNameAndID(ctx, workflow.Name, inserted.ID)
	assert.Nil(t, err)
	assert.Equal(t, model.StatusFailure, jobFromDb.Status)
}

func TestUpdateJobStatusInvalid(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	workflow := model.Workflow{
		Name:          "TestUpdateJobStatusInvalid",
		Description:   "Description",
		SchemaVersion: 1,
		Version:       1,
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
		InputParameters: []string{
			"key1",
			"key2",
		},
	}

	ctx := context.Background()
	count, _ := testDataStore.InsertWorkflows(ctx, workflow)
	assert.Equal(t, 1, count)

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
		InputParameters: []model.InputParameter{
			{
				Name:  "key1",
				Value: "value1",
			},
			{
				Name:  "key2",
				Value: "value2",
			},
		},
	}

	inserted, err := testDataStore.UpsertJob(ctx, job)
	assert.Nil(t, err)

	err = testDataStore.UpdateJobStatus(ctx, inserted, 99999)
	assert.NotNil(t, err)

	jobFromDb, err := testDataStore.GetJobByNameAndID(ctx, workflow.Name, inserted.ID)
	assert.Nil(t, err)
	assert.Equal(t, model.StatusPending, jobFromDb.Status)
}

func TestUpdateJobAddResult(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	workflow := model.Workflow{
		Name:          "TestUpdateJobAddResult",
		Description:   "Description",
		SchemaVersion: 1,
		Version:       1,
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
		InputParameters: []string{
			"key1",
			"key2",
		},
	}

	ctx := context.Background()
	count, _ := testDataStore.InsertWorkflows(ctx, workflow)
	assert.Equal(t, 1, count)

	job := &model.Job{
		WorkflowName: workflow.Name,
		InputParameters: []model.InputParameter{
			{
				Name:  "key1",
				Value: "value1",
			},
			{
				Name:  "key2",
				Value: "value2",
			},
		},
	}

	inserted, err := testDataStore.UpsertJob(ctx, job)
	assert.Nil(t, err)

	result := &model.TaskResult{
		CLI: &model.TaskResultCLI{
			Command: []string{
				"command",
			},
			Output:   "output",
			Error:    "error",
			ExitCode: 0,
		},
	}
	err = testDataStore.UpdateJobAddResult(ctx, inserted, result)
	assert.Nil(t, err)

	jobFromDb, err := testDataStore.GetJobByNameAndID(ctx, workflow.Name, inserted.ID)
	assert.Nil(t, err)
	assert.Len(t, jobFromDb.Results, 1)
	assert.NotNil(t, jobFromDb.Results[0].CLI)
	assert.Equal(t, []string{"command"}, jobFromDb.Results[0].CLI.Command)
	assert.Equal(t, "output", jobFromDb.Results[0].CLI.Output)
	assert.Equal(t, "error", jobFromDb.Results[0].CLI.Error)
	assert.Equal(t, 0, jobFromDb.Results[0].CLI.ExitCode)
}

func TestGetAllJobs(t *testing.T) {
	flag.Parse()
	if testing.Short() {
		t.Skip()
	}

	workflow := model.Workflow{
		Name:          "TestGetAllJobs",
		Description:   "Description",
		SchemaVersion: 1,
		Version:       1,
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
		InputParameters: []string{
			"key1",
			"key2",
		},
	}

	ctx := context.Background()
	count, _ := testDataStore.InsertWorkflows(ctx, workflow)
	assert.Equal(t, 1, count)

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
		InputParameters: []model.InputParameter{
			{
				Name:  "key1",
				Value: "value1",
			},
			{
				Name:  "key2",
				Value: "value2",
			},
		},
	}

	database := testDataStore.client.Database(testDataStore.dbName)
	collJobs := database.Collection(JobsCollectionName)
	collJobs.DeleteMany(ctx, bson.M{})

	_, err := testDataStore.UpsertJob(ctx, job)
	assert.Nil(t, err)

	jobs, totalCount, err := testDataStore.GetAllJobs(ctx, 1, 4)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(jobs))
	assert.Equal(t, int64(1), totalCount)
	assert.Equal(t, job.WorkflowName, jobs[0].WorkflowName)
	assert.Equal(t, job.InputParameters, jobs[0].InputParameters)
	assert.Equal(t, model.StatusPending, jobs[0].Status)

	j, err := testDataStore.GetJobByID(ctx, job.ID)
	assert.Nil(t, err)
	assert.Equal(t, job.ID, j.ID)
	assert.Equal(t, job.WorkflowName, j.WorkflowName)
	assert.Equal(t, job.InputParameters, j.InputParameters)
	assert.Equal(t, model.StatusPending, j.Status)

	j, err = testDataStore.GetJobByID(ctx, "job.ID")
	assert.Nil(t, j)
	assert.Nil(t, err)
}

func TestPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	err := testDataStore.Ping(ctx)
	assert.NoError(t, err)
}
