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

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	mock_nats "github.com/mendersoftware/mender-server/services/workflows/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store"
	"github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		Name string

		DataStoreErr    error
		NatsIsConnected bool

		HTTPStatus int
		HTTPBody   map[string]interface{}
	}{{
		Name:            "ok",
		NatsIsConnected: true,
		HTTPStatus:      http.StatusNoContent,
	}, {
		Name:         "error, MongoDB not reachable",
		DataStoreErr: errors.New("connection refused"),
		HTTPStatus:   http.StatusServiceUnavailable,
		HTTPBody: map[string]interface{}{
			"error": "error reaching MongoDB: connection refused",
		},
	}, {
		Name:            "error, nats not connected",
		NatsIsConnected: false,
		HTTPStatus:      http.StatusServiceUnavailable,
		HTTPBody: map[string]interface{}{
			"error": "not connected to nats",
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			dataStore := mock.NewDataStore()
			defer dataStore.AssertExpectations(t)
			dataStore.On("Ping", mocklib.Anything).
				Return(tc.DataStoreErr)

			nats := &mock_nats.Client{}
			defer nats.AssertExpectations(t)
			if tc.DataStoreErr == nil {
				nats.On("IsConnected").Return(tc.NatsIsConnected)
			}

			router := NewRouter(dataStore, nats)

			req, err := http.NewRequest(http.MethodGet, "http://localhost"+APIURLHealth, nil)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)
			if tc.HTTPStatus >= 400 {
				var bodyJSON map[string]interface{}
				decoder := json.NewDecoder(w.Body)
				err := decoder.Decode(&bodyJSON)
				if assert.NoError(t, err) {
					assert.Equal(t, tc.HTTPBody, bodyJSON)
				}
			} else {
				assert.Nil(t, w.Body.Bytes())
			}
		})
	}
}

func TestWorkflowNotFound(t *testing.T) {
	const workflowName = "test"

	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflowName,
		mocklib.AnythingOfType("string"),
	).Return(nil, store.ErrWorkflowNotFound)

	payload := `{
      "key": "value"
	}`

	url := strings.Replace(APIURLWorkflow, ":name", workflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkflowFoundButMissingParameters(t *testing.T) {
	const workflowName = "test"

	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		InputParameters: []string{"param1", "param2", "param3"},
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflowName,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	payload := `{
      "key": "value"
	}`

	url := strings.Replace(APIURLWorkflow, ":name", workflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Missing input parameters: [param1 param2 param3]",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestWorkflowFoundAndStartedWithInvalidParameters(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	payload := ``

	const workflowName = "test"
	url := strings.Replace(APIURLWorkflow, ":name", workflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Unable to parse the input parameters: EOF",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestWorkflowFoundAndStartedWithParameters(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflow := &model.Workflow{
		Name: "test",
	}
	mockedJob := &model.Job{
		ID:           "123456",
		WorkflowName: workflow.Name,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "value", job.InputParameters[0].Value)

			return true
		}),
	).Return(nil)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(mockedJob, nil)

	payload := `{
      "key": "value"
	}`

	url := strings.Replace(APIURLWorkflow, ":name", workflow.Name, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	value, ok := response["id"]
	assert.True(t, ok)

	w = httptest.NewRecorder()
	url = strings.Replace(strings.Replace(APIURLWorkflowID, ":name", workflow.Name, 1), ":id", value, 1)
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkflowFailToPublish(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflow := &model.Workflow{
		Name: "test",
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "value", job.InputParameters[0].Value)

			return true
		}),
	).Return(errors.New("failure"))

	payload := `{
      "key": "value"
	}`

	url := strings.Replace(APIURLWorkflow, ":name", workflow.Name, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWorkflowFoundAndStartedWithVersion(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflowVersion := "16"
	workflow := &model.Workflow{
		Name:    "test",
		Version: 16,
	}
	mockedJob := &model.Job{
		ID:              "1234567890",
		WorkflowName:    "test",
		WorkflowVersion: workflowVersion,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, workflowVersion, job.WorkflowVersion)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "value", job.InputParameters[0].Value)

			return true
		}),
	).Return(nil)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(mockedJob, nil)

	payload := `{
      "key": "value"
	}`

	url := strings.Replace(APIURLWorkflow, ":name", mockedJob.WorkflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	req.Header.Add(HeaderWorkflowMinVersion, workflowVersion)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	value, ok := response["id"]
	assert.True(t, ok)

	w = httptest.NewRecorder()
	url = strings.Replace(strings.Replace(APIURLWorkflowID, ":name", mockedJob.WorkflowName, 1), ":id", value, 1)
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var job model.Job
	body = w.Body.Bytes()
	err = json.Unmarshal(body, &job)

	assert.Nil(t, err)
	assert.Equal(t, mockedJob.WorkflowName, job.WorkflowName)
	assert.Equal(t, int32(0), job.Status)
	assert.Equal(t, workflowVersion, job.WorkflowVersion)
}

func TestWorkflowFoundAndStartedWithNonStringParameter(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflow := &model.Workflow{
		Name: "test",
	}
	mockedJob := &model.Job{
		ID:           "123456",
		WorkflowName: workflow.Name,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "2", job.InputParameters[0].Value)

			return true
		}),
	).Return(nil)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(mockedJob, nil)

	payload := `{
      "key": 2
	}`

	url := strings.Replace(APIURLWorkflow, ":name", workflow.Name, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	value, ok := response["id"]
	assert.True(t, ok)

	w = httptest.NewRecorder()
	url = strings.Replace(strings.Replace(APIURLWorkflowID, ":name", workflow.Name, 1), ":id", value, 1)
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkflowFoundAndStartedWithListOfStringsParameter(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflow := &model.Workflow{
		Name: "test",
	}
	mockedJob := &model.Job{
		ID:           "123456",
		WorkflowName: workflow.Name,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "2,3,4", job.InputParameters[0].Value)

			return true
		}),
	).Return(nil)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(mockedJob, nil)

	payload := `{
		"key": ["2", 3, 4.0]
	}`

	url := strings.Replace(APIURLWorkflow, ":name", workflow.Name, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	value, ok := response["id"]
	assert.True(t, ok)

	w = httptest.NewRecorder()
	url = strings.Replace(strings.Replace(APIURLWorkflowID, ":name", workflow.Name, 1), ":id", value, 1)
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBatchWorkflowNotFound(t *testing.T) {
	const workflowName = "test"

	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflowName,
		mocklib.AnythingOfType("string"),
	).Return(nil, store.ErrWorkflowNotFound)

	payload := `[{
      "key": "value"
	}]`

	url := strings.Replace(APIURLWorkflowBatch, ":name", workflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBatchWorkflowFoundButMissingParameters(t *testing.T) {
	const workflowName = "test"

	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		InputParameters: []string{"param1", "param2", "param3"},
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflowName,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	payload := `[{
      "key": "value"
	}]`

	url := strings.Replace(APIURLWorkflowBatch, ":name", workflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Missing input parameters: [param1 param2 param3]",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestBatchWorkflowFoundAndStartedWithInvalidParameters(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	payload := ``

	const workflowName = "test"
	url := strings.Replace(APIURLWorkflowBatch, ":name", workflowName, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Unable to parse the input parameters: EOF",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestBatchWorkflowFoundAndStartedWithParameters(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflow := &model.Workflow{
		Name: "test",
	}
	mockedJob := &model.Job{
		ID:           "123456",
		WorkflowName: workflow.Name,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "value", job.InputParameters[0].Value)

			return true
		}),
	).Return(nil)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(mockedJob, nil)

	payload := `[{
      "key": "value"
	}]`

	url := strings.Replace(APIURLWorkflowBatch, ":name", workflow.Name, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response []map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	value, ok := response[0]["id"]
	assert.True(t, ok)

	w = httptest.NewRecorder()
	url = strings.Replace(strings.Replace(APIURLWorkflowID, ":name", workflow.Name, 1), ":id", value, 1)
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBatchWorkflowFailToPublish(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	nats := &mock_nats.Client{}
	defer nats.AssertExpectations(t)

	nats.On("StreamName").Return("stream")

	workflow := &model.Workflow{
		Name: "test",
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	nats.On("JetStreamPublish",
		"stream.default",
		mocklib.MatchedBy(func(data []byte) bool {
			job := &model.Job{}
			err := json.Unmarshal(data, job)
			assert.NoError(t, err)
			assert.Equal(t, workflow.Name, job.WorkflowName)
			assert.Equal(t, "key", job.InputParameters[0].Name)
			assert.Equal(t, "value", job.InputParameters[0].Value)

			return true
		}),
	).Return(errors.New("failure"))

	payload := `[{
      "key": "value"
	}]`

	url := strings.Replace(APIURLWorkflowBatch, ":name", workflow.Name, 1)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response []map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	value, _ := response[0]["error"]
	assert.NotEmpty(t, value)
}

func TestWorkflowByNameAndIDNotFound(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	const (
		workflowName = "test"
		jobID        = "dummy"
	)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflowName,
		jobID,
	).Return(nil, nil)

	url := strings.Replace(strings.Replace(APIURLWorkflowID, ":name", workflowName, 1), ":id", jobID, 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkflowByNameAndIDError(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	const (
		workflowName = "test"
		jobID        = "dummy"
	)

	dataStore.On("GetJobByNameAndID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflowName,
		jobID,
	).Return(nil, errors.New("error"))

	url := strings.Replace(strings.Replace(APIURLWorkflowID, ":name", workflowName, 1), ":id", jobID, 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRegisterWorkflowBadRequestIncompletePayload(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	payload := ``
	req, err := http.NewRequest(http.MethodPost, APIURLWorkflows, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Error parsing JSON form: unexpected end of JSON input",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestRegisterWorkflowBadRequestMissingName(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	payload := `{}`
	req, err := http.NewRequest(http.MethodPost, APIURLWorkflows, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Workflow missing name",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestRegisterWorkflowAlreadyExists(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	dataStore.On("InsertWorkflows",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		mocklib.AnythingOfType("[]model.Workflow"),
	).Return(nil, store.ErrWorkflowAlreadyExists)

	payload := `{"name": "workflow"}`
	req, err := http.NewRequest(http.MethodPost, APIURLWorkflows, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	value, ok := response["error"]

	assert.Nil(t, err)
	assert.True(t, ok)

	expectedBody := gin.H{
		"error": "Workflow already exists",
	}
	assert.Equal(t, expectedBody["error"], value)
}

func TestRegisterWorkflow(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	dataStore.On("InsertWorkflows",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		mocklib.MatchedBy(
			func(workflows []model.Workflow) bool {
				assert.Len(t, workflows, 1)

				workflow := workflows[0]
				assert.Equal(t, "test_workflow", workflow.Name)
				assert.Equal(t, "Test workflow", workflow.Description)
				assert.Equal(t, 4, workflow.Version)
				assert.Equal(t, 1, workflow.SchemaVersion)

				assert.Len(t, workflow.Tasks, 1)

				task := workflow.Tasks[0]
				assert.Equal(t, "test_http_call", task.Name)
				assert.Equal(t, model.TaskTypeHTTP, task.Type)
				assert.Equal(t, "http://localhost:8000", task.HTTP.URI)

				return true
			}),
	).Return(1, nil)

	payload := `{
		"name": "test_workflow",
		"description": "Test workflow",
		"version": 4,
		"tasks": [
			{
				"name": "test_http_call",
				"type": "http",
				"http": {
					"uri": "http://localhost:8000",
					"method": "POST",
					"body": "{\"device_id\": \"${workflow.input.device_id}\"}",
					"connectionTimeOut": 1000,
					"readTimeOut": 1000
				}
			}
		],
		"inputParameters": [
			"device_id"
		],
		"schemaVersion": 1
	}`

	req, err := http.NewRequest(http.MethodPost, APIURLWorkflows, strings.NewReader(payload))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetWorkflows(t *testing.T) {
	task := model.Task{
		Name: "test_http_call",
		Type: model.TaskTypeHTTP,
		HTTP: &model.HTTPTask{
			URI:               "http://localhost:8000",
			Method:            http.MethodPost,
			Body:              "{\"device_id\": \"${workflow.input.device_id}\"}",
			ConnectionTimeOut: 1000,
			ReadTimeOut:       1000,
		},
	}
	workflow := model.Workflow{
		Name:        "test_workflow",
		Description: "Test workflow",
		Version:     4,
		Tasks:       []model.Task{task},
		InputParameters: []string{
			"device_id",
		},
		SchemaVersion: 1,
	}

	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	dataStore.On("GetWorkflows",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
	).Return([]model.Workflow{workflow}, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, APIURLWorkflows, nil)
	assert.NoError(t, err)

	router := NewRouter(dataStore, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var workflows []model.Workflow
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &workflows)

	assert.Nil(t, err)
	assert.Len(t, workflows, 1)
	assert.Equal(t, "test_workflow", workflows[0].Name)
	assert.Equal(t, "Test workflow", workflows[0].Description)
	assert.Equal(t, 4, workflows[0].Version)
	assert.Equal(t, 1, workflows[0].SchemaVersion)
}
