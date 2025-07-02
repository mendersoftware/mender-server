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

package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	expiredCtx, cancel := context.WithDeadline(
		context.TODO(), time.Now().Add(-1*time.Second))
	defer cancel()
	defaultCtx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	testCases := []struct {
		Name string

		Ctx context.Context

		// Workflows response
		ResponseCode int
		ResponseBody interface{}

		Error error
	}{{
		Name: "ok",

		Ctx:          defaultCtx,
		ResponseCode: http.StatusOK,
	}, {
		Name: "error, expired deadline",

		Ctx:   expiredCtx,
		Error: errors.New(context.DeadlineExceeded.Error()),
	}, {
		Name: "error, workflows unhealthy",

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: rest.Error{
			Err:       "internal error",
			RequestID: "test",
		},

		Error: errors.New("internal error"),
	}, {
		Name: "error, bad response",

		Ctx: context.TODO(),

		ResponseCode: http.StatusServiceUnavailable,
		ResponseBody: "foobar",

		Error: errors.New("health check HTTP error: 503 Service Unavailable"),
	}}

	responses := make(chan http.Response, 1)
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		rsp := <-responses
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(serveHTTP))
	client := NewClient().(*client)
	client.baseURL = srv.URL
	defer srv.Close()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			if tc.ResponseCode > 0 {
				rsp := http.Response{
					StatusCode: tc.ResponseCode,
				}
				if tc.ResponseBody != nil {
					b, _ := json.Marshal(tc.ResponseBody)
					rsp.Body = ioutil.NopCloser(bytes.NewReader(b))
				}
				responses <- rsp
			}

			err := client.CheckHealth(tc.Ctx)

			if tc.Error != nil {
				assert.Contains(t, err.Error(), tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateArtifactFails(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))

	workflowsClient := NewClient().(*client)
	workflowsClient.baseURL = srv.URL

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		TenantID:              "tenant_id",
		ArtifactID:            "artifact_id",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	ctx := context.Background()
	err := workflowsClient.StartGenerateArtifact(ctx, multipartGenerateImage)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to start workflow: generate_artifact")
}

func TestGenerateArtifactSuccessful(t *testing.T) {
	reqChan := make(chan *model.MultipartGenerateImageMsg, 1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		var multipartGenerateImage model.MultipartGenerateImageMsg
		defer w.WriteHeader(http.StatusCreated)
		b, err := ioutil.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		err = json.Unmarshal(b, &multipartGenerateImage)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		select {
		case reqChan <- &multipartGenerateImage:
		default:
			t.FailNow()
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(handler))

	workflowsClient := NewClient().(*client)
	workflowsClient.baseURL = srv.URL

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "args",
		TenantID:              "tenant_id",
		ArtifactID:            "artifact_id",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	ctx := context.Background()
	err := workflowsClient.StartGenerateArtifact(ctx, multipartGenerateImage)
	assert.Nil(t, err)
	select {
	case multipartGenerateImage = <-reqChan:

	default:
		panic("[PROG ERR] Did not receive any response from httptest handler")
	}
	assert.NoError(t, err)
	assert.Equal(t, "name", multipartGenerateImage.Name)
	assert.Equal(t, "description", multipartGenerateImage.Description)
	assert.Len(t, multipartGenerateImage.DeviceTypesCompatible, 1)
	assert.Equal(t, "Beagle Bone", multipartGenerateImage.DeviceTypesCompatible[0])
	assert.Equal(t, "single_file", multipartGenerateImage.Type)
	assert.Equal(t, "args", multipartGenerateImage.Args)
	assert.Equal(t, "tenant_id", multipartGenerateImage.TenantID)
	assert.Equal(t, "artifact_id", multipartGenerateImage.ArtifactID)
}

func mockServerReindex(t *testing.T, tenant, device, reqid string, code int) (*httptest.Server, error) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if code != http.StatusOK {
			w.WriteHeader(code)
			return
		}
		defer r.Body.Close()

		request := ReindexWorkflow{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		assert.NoError(t, err)

		assert.Equal(t, reqid, request.RequestID)
		assert.Equal(t, tenant, request.TenantID)
		assert.Equal(t, device, request.DeviceID)
		assert.Equal(t, ServiceDeployments, request.Service)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}

	srv := httptest.NewServer(http.HandlerFunc(h))
	return srv, nil
}

func TestReindexReporting(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		tenant string
		device string
		reqid  string

		code int

		err error
	}{
		{
			name:   "ok",
			tenant: "tenant1",
			device: "device2",
			reqid:  "reqid1",

			code: http.StatusOK,
		},
		{
			name:   "404",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			code: http.StatusNotFound,
			err:  errors.New(`workflows: workflow "reindex_reporting" not defined`),
		},
		{
			name:   "500",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			code: http.StatusInternalServerError,
			err:  errors.New(`workflows: unexpected HTTP status from workflows service: 500 Internal Server Error`),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv, err := mockServerReindex(t, tc.tenant, tc.device, tc.reqid, tc.code)
			assert.NoError(t, err)

			defer srv.Close()

			ctx := context.Background()
			ctx = requestid.WithContext(ctx, tc.reqid)
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx,
					&identity.Identity{
						Tenant: tc.tenant,
					})
			}

			client := NewClient().(*client)
			client.baseURL = srv.URL

			err = client.StartReindexReporting(ctx, tc.device)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockServerReindexDeployment(t *testing.T, tenant, device, deployment, id, reqid string,
	code int) (*httptest.Server, error) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if code != http.StatusOK {
			w.WriteHeader(code)
			return
		}
		defer r.Body.Close()

		request := ReindexDeploymentWorkflow{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		assert.NoError(t, err)

		assert.Equal(t, reqid, request.RequestID)
		assert.Equal(t, tenant, request.TenantID)
		assert.Equal(t, device, request.DeviceID)
		assert.Equal(t, deployment, request.DeploymentID)
		assert.Equal(t, id, request.ID)
		assert.Equal(t, ServiceDeployments, request.Service)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}

	srv := httptest.NewServer(http.HandlerFunc(h))
	return srv, nil
}

func TestReindexDeploymentWorkflow(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		tenant     string
		device     string
		deployment string
		id         string
		reqid      string

		code int

		err error
	}{
		{
			name:       "ok",
			tenant:     "tenant1",
			device:     "device2",
			deployment: "deployment3",
			id:         "id4",
			reqid:      "reqid1",

			code: http.StatusOK,
		},
		{
			name:   "404",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			code: http.StatusNotFound,
			err:  errors.New(`workflows: workflow "reindex_reporting_deployment" not defined`),
		},
		{
			name:   "500",
			tenant: "tenant2",
			device: "device3",
			reqid:  "reqid2",

			code: http.StatusInternalServerError,
			err:  errors.New(`workflows: unexpected HTTP status from workflows service: 500 Internal Server Error`),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv, err := mockServerReindexDeployment(t, tc.tenant, tc.device, tc.deployment,
				tc.id, tc.reqid, tc.code)
			assert.NoError(t, err)

			defer srv.Close()

			ctx := context.Background()
			ctx = requestid.WithContext(ctx, tc.reqid)
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx,
					&identity.Identity{
						Tenant: tc.tenant,
					})
			}

			client := NewClient().(*client)
			client.baseURL = srv.URL

			err = client.StartReindexReportingDeployment(ctx, tc.device, tc.deployment, tc.id)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockServerReindexDeploymentBatch(
	t *testing.T,
	tenant string,
	deviceDeployments []DeviceDeploymentShortInfo,
	id, reqid string,
	code int,
) (*httptest.Server, error) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if code != http.StatusOK {
			w.WriteHeader(code)
			return
		}
		defer r.Body.Close()

		request := make([]ReindexDeploymentWorkflow, 100)

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		assert.NoError(t, err)
		assert.Len(t, request, len(deviceDeployments))

		for i, dd := range deviceDeployments {
			assert.Equal(t, reqid, request[i].RequestID)
			assert.Equal(t, tenant, request[i].TenantID)
			assert.Equal(t, dd.DeviceID, request[i].DeviceID)
			assert.Equal(t, dd.DeploymentID, request[i].DeploymentID)
			assert.Equal(t, dd.ID, request[i].ID)
			assert.Equal(t, ServiceDeployments, request[i].Service)
		}

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}

	srv := httptest.NewServer(http.HandlerFunc(h))
	return srv, nil
}

func TestReindexDeploymentBatchWorkflow(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		tenant            string
		deviceDeployments []DeviceDeploymentShortInfo
		id                string
		reqid             string

		code int

		err error
	}{
		{
			name:   "ok",
			tenant: "tenant1",
			deviceDeployments: []DeviceDeploymentShortInfo{
				{
					ID:           "id1",
					DeviceID:     "device1",
					DeploymentID: "deployment1",
				},
				{
					ID:           "id2",
					DeviceID:     "device2",
					DeploymentID: "deployment2",
				},
			},
			reqid: "reqid1",

			code: http.StatusOK,
		},
		{
			name:   "404",
			tenant: "tenant2",
			deviceDeployments: []DeviceDeploymentShortInfo{
				{
					ID:           "id1",
					DeviceID:     "device1",
					DeploymentID: "deployment1",
				},
				{
					ID:           "id2",
					DeviceID:     "device2",
					DeploymentID: "deployment2",
				},
			},
			reqid: "reqid2",

			code: http.StatusNotFound,
			err:  errors.New(`workflows: workflow "reindex_reporting_deployment" not defined`),
		},
		{
			name:   "500",
			tenant: "tenant2",
			deviceDeployments: []DeviceDeploymentShortInfo{
				{
					ID:           "id1",
					DeviceID:     "device1",
					DeploymentID: "deployment1",
				},
				{
					ID:           "id2",
					DeviceID:     "device2",
					DeploymentID: "deployment2",
				},
			},
			reqid: "reqid2",

			code: http.StatusInternalServerError,
			err:  errors.New(`workflows: unexpected HTTP status from workflows service: 500 Internal Server Error`),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv, err := mockServerReindexDeploymentBatch(t, tc.tenant, tc.deviceDeployments,
				tc.id, tc.reqid, tc.code)
			assert.NoError(t, err)

			defer srv.Close()

			ctx := context.Background()
			ctx = requestid.WithContext(ctx, tc.reqid)
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx,
					&identity.Identity{
						Tenant: tc.tenant,
					})
			}

			client := NewClient().(*client)
			client.baseURL = srv.URL

			err = client.StartReindexReportingDeploymentBatch(ctx, tc.deviceDeployments)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
