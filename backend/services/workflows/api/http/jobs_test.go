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

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	mock_nats "github.com/mendersoftware/mender-server/services/workflows/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

func TestGetJobByID(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
	}
	mockedJob := &model.Job{
		ID:           "123456",
		WorkflowName: workflow.Name,
	}

	dataStore.On("GetJobByID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		mocklib.AnythingOfType("string"),
	).Return(mockedJob, nil)

	url := strings.Replace(APIURLJobsID, ":id", mockedJob.ID, 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	nats := &mock_nats.Client{}
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, mockedJob.ID, response["id"])
	assert.Equal(t, mockedJob.WorkflowName, response["workflowName"])
}

func TestGetJobByIDNotFound(t *testing.T) {
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
	}
	mockedJob := &model.Job{
		ID:           "123456",
		WorkflowName: workflow.Name,
	}

	dataStore.On("GetJobByID",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		mocklib.AnythingOfType("string"),
	).Return(nil, nil)

	url := strings.Replace(APIURLJobsID, ":id", mockedJob.ID, 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	nats := &mock_nats.Client{}
	router := NewRouter(dataStore, nats)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	body := w.Body.Bytes()
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "not found", response["error"])
}
