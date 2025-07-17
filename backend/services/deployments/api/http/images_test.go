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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mt "github.com/mendersoftware/mender-server/pkg/testing"

	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	"github.com/mendersoftware/mender-server/services/deployments/app"
	app_mocks "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	dmodel "github.com/mendersoftware/mender-server/services/deployments/model"
	store_mocks "github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	store_mongo "github.com/mendersoftware/mender-server/services/deployments/store/mongo"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
	deployments_testing "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
	h "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
)

func TestPostArtifacts(t *testing.T) {
	var testConflictError = model.NewConflictError(
		store_mongo.ErrConflictingDepends,
	).WithMetadata(map[string]interface{}{
		"conflict": map[string]interface{}{"device_type": "arm6"},
	})

	imageBody := []byte("123456790")

	testCases := []struct {
		requestBodyObject      []h.Part
		requestContentType     string
		responseCode           int
		responseBody           string
		appCreateImage         bool
		appCreateImageResponse string
		appCreateImageError    error
	}{
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "",
			responseCode:       http.StatusBadRequest,
			responseBody:       "request Content-Type isn't multipart/form-data",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "application/x-www-form-urlencoded",
			responseCode:       http.StatusBadRequest,
			responseBody:       "request Content-Type isn't multipart/form-data",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       ErrArtifactFileMissing.Error(),
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "artifact_id",
					FieldValue: "wrong_uuidv4",
				},
				{
					FieldName:   "artifact",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "artifact_id is not a valid UUID",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "id",
					FieldValue: "5e2fbcf6a6a7eca56cbc9476",
				},
				{
					FieldName:  "artifact_id",
					FieldValue: "24436884-a710-4d20-aec4-82c89fbfe29e",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:   "artifact",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:     "multipart/form-data",
			responseCode:           http.StatusCreated,
			responseBody:           "",
			appCreateImage:         true,
			appCreateImageResponse: "24436884-a710-4d20-aec4-82c89fbfe29e",
			appCreateImageError:    nil,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "id",
					FieldValue: "5e2fbcf6a6a7eca56cbc9476",
				},
				{
					FieldName:  "artifact_id",
					FieldValue: "24436884-a710-4d20-aec4-82c89fbfe29e",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:   "artifact",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusConflict,
			// no slashes will be present in the real response - must be added
			// because we're comparing to body, _ := recorded.DecodedBody(), which does does funny formatting
			responseBody:           store_mongo.ErrConflictingDepends.Error(),
			appCreateImage:         true,
			appCreateImageResponse: "24436884-a710-4d20-aec4-82c89fbfe29e",
			appCreateImageError:    testConflictError,
		},
	}

	store := &store_mocks.DataStore{}
	restView := new(view.RESTView)

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			app := &app_mocks.App{}

			if tc.appCreateImage {
				app.On("CreateImage",
					h.ContextMatcher(),
					mock.MatchedBy(func(msg *model.MultipartUploadMsg) bool {
						assert.Equal(t, msg.ArtifactID, tc.requestBodyObject[1].FieldValue)
						assert.Equal(t, msg.MetaConstructor.Description, tc.requestBodyObject[2].FieldValue)

						return true
					}),
				).Return(tc.appCreateImageResponse, tc.appCreateImageError)
			}

			d := NewDeploymentsApiHandlers(store, restView, app)
			router := setUpTestRouter()
			router.POST("/api/0.0.1/artifacts", d.NewImage)
			req := h.MakeMultipartRequest("POST", "http://localhost/api/0.0.1/artifacts",
				tc.requestContentType, tc.requestBodyObject)
			req.Header.Set("Authorization", HTTPHeaderAuthorizationBearer+" TOKEN")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.responseCode, w.Code)
			if tc.responseBody == "" {
				assert.Empty(t, w.Body.String())
			} else {
				assert.Contains(t, w.Body.String(), tc.responseBody)
			}

			if tc.appCreateImage {
				app.AssertExpectations(t)
			}
		})
	}

}

func TestPostArtifactsInternal(t *testing.T) {
	imageBody := []byte("123456790")
	var testConflictError = model.NewConflictError(
		store_mongo.ErrConflictingDepends,
	).WithMetadata(map[string]interface{}{
		"conflict": map[string]interface{}{"device_type": "arm6"},
	})

	testCases := []struct {
		requestBodyObject      []h.Part
		requestContentType     string
		responseCode           int
		responseBody           string
		appCreateImage         bool
		appCreateImageResponse string
		appCreateImageError    error
	}{
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "",
			responseCode:       http.StatusBadRequest,
			responseBody:       "request Content-Type isn't multipart/form-data",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "application/x-www-form-urlencoded",
			responseCode:       http.StatusBadRequest,
			responseBody:       "request Content-Type isn't multipart/form-data",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       ErrArtifactFileMissing.Error(),
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "artifact_id",
					FieldValue: "wrong_uuidv4",
				},
				{
					FieldName:   "artifact",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "artifact_id is not a valid UUID",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "id",
					FieldValue: "5e2fbcf6a6a7eca56cbc9476",
				},
				{
					FieldName:  "artifact_id",
					FieldValue: "24436884-a710-4d20-aec4-82c89fbfe29e",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:   "artifact",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:     "multipart/form-data",
			responseCode:           http.StatusCreated,
			responseBody:           "",
			appCreateImage:         true,
			appCreateImageResponse: "24436884-a710-4d20-aec4-82c89fbfe29e",
			appCreateImageError:    nil,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "id",
					FieldValue: "5e2fbcf6a6a7eca56cbc9476",
				},
				{
					FieldName:  "artifact_id",
					FieldValue: "24436884-a710-4d20-aec4-82c89fbfe29e",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:   "artifact",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusConflict,
			// no slashes will be present in the real response - must be added
			// because we're comparing to body, _ := recorded.DecodedBody(), which does does funny formatting
			responseBody:           store_mongo.ErrConflictingDepends.Error(),
			appCreateImage:         true,
			appCreateImageResponse: "24436884-a710-4d20-aec4-82c89fbfe29e",
			appCreateImageError:    testConflictError,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: "999999999999999999999",
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:     "multipart/form-data",
			responseCode:           http.StatusBadRequest,
			responseBody:           "failed to read form value 'size'",
			appCreateImage:         false,
			appCreateImageResponse: "",
			appCreateImageError:    nil,
		},
	}

	store := &store_mocks.DataStore{}
	restView := new(view.RESTView)

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			app := &app_mocks.App{}

			if tc.appCreateImage {
				app.On("CreateImage",
					h.ContextMatcher(),
					mock.MatchedBy(func(msg *model.MultipartUploadMsg) bool {
						assert.Equal(t, msg.ArtifactID, tc.requestBodyObject[1].FieldValue)
						assert.Equal(t, msg.MetaConstructor.Description, tc.requestBodyObject[2].FieldValue)

						return true
					}),
				).Return(tc.appCreateImageResponse, tc.appCreateImageError)
			}

			d := NewDeploymentsApiHandlers(store, restView, app)
			router := setUpTestRouter()
			router.POST("/api/0.0.1/tenants/:tenant/artifacts", d.NewImageForTenantHandler)
			req := h.MakeMultipartRequest("POST", "http://localhost/api/0.0.1/tenants/default/artifacts",
				tc.requestContentType, tc.requestBodyObject)
			req.Header.Set("Authorization", HTTPHeaderAuthorizationBearer+" TOKEN")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.responseCode, w.Code, "unexpected HTTP status code")
			if tc.responseBody == "" {
				assert.Empty(t, w.Body.String(), "HTTP body not empty")
			} else {
				assert.Contains(t, w.Body.String(), tc.responseBody)
			}

			if tc.appCreateImage {
				app.AssertExpectations(t)
			}
		})
	}
}

func TestPostArtifactsGenerate(t *testing.T) {
	imageBody := []byte("123456790")

	testCases := []struct {
		requestBodyObject        []h.Part
		requestContentType       string
		responseCode             int
		responseBody             string
		appGenerateImage         bool
		appGenerateImageResponse string
		appGenerateImageError    error
	}{
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "",
			responseCode:       http.StatusBadRequest,
			responseBody:       "request Content-Type isn't multipart/form-data",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "application/x-www-form-urlencoded",
			responseCode:       http.StatusBadRequest,
			responseBody:       "request Content-Type isn't multipart/form-data",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "api: invalid form parameters:",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody: "api: invalid form parameters: " +
				"device_types_compatible: cannot be blank.",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "api: invalid form parameters: missing 'file' section",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody: "api: invalid form parameters: " +
				"type: cannot be blank.",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusCreated,
			responseBody:             "",
			appGenerateImage:         true,
			appGenerateImageResponse: "artifactID",
			appGenerateImageError:    nil,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name with spaces",
				},
				{
					FieldName:  "description",
					FieldValue: "description with spaces",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "arg1 arg2 arg3",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusCreated,
			responseBody:             "",
			appGenerateImage:         true,
			appGenerateImageResponse: "artifactID",
			appGenerateImageError:    nil,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusUnprocessableEntity,
			responseBody:             "Artifact not unique",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    app.ErrModelArtifactNotUnique,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusRequestEntityTooLarge,
			responseBody:             "Artifact file too large",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    ErrModelArtifactFileTooLarge,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusBadRequest,
			responseBody:             "Cannot parse artifact file",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    app.ErrModelParsingArtifactFailed,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusInternalServerError,
			responseBody:             "internal error",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    errors.New("generic error"),
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: "999999999999999999999",
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusBadRequest,
			responseBody:             "failed to read form value 'size'",
			appGenerateImage:         false,
			appGenerateImageResponse: "",
			appGenerateImageError:    nil,
		},
	}

	store := &store_mocks.DataStore{}
	restView := new(view.RESTView)

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			app := &app_mocks.App{}

			if tc.appGenerateImage {
				app.On("GenerateImage",
					h.ContextMatcher(),
					mock.MatchedBy(func(msg *model.MultipartGenerateImageMsg) bool {
						assert.Equal(t, msg.Name, tc.requestBodyObject[0].FieldValue)
						assert.Equal(t, msg.Description, tc.requestBodyObject[1].FieldValue)
						assert.Equal(t, msg.DeviceTypesCompatible, []string{tc.requestBodyObject[3].FieldValue})
						assert.Equal(t, msg.Type, tc.requestBodyObject[4].FieldValue)
						assert.Equal(t, msg.Args, tc.requestBodyObject[5].FieldValue)
						assert.Equal(t, msg.Token, "TOKEN")

						return true
					}),
				).Return(tc.appGenerateImageResponse, tc.appGenerateImageError)
			}

			d := NewDeploymentsApiHandlers(store, restView, app)
			router := setUpTestRouter()
			router.POST("/api/0.0.1/artifacts/generate", d.GenerateImage)
			req := h.MakeMultipartRequest("POST", "http://localhost/api/0.0.1/artifacts/generate",
				tc.requestContentType, tc.requestBodyObject)
			req.Header.Set("Authorization", HTTPHeaderAuthorizationBearer+" TOKEN")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.responseCode, w.Code, "unexpected HTTP code")
			if tc.responseBody == "" {
				assert.Empty(t, w.Body.String(), "HTTP body not empty")
			} else {
				assert.Contains(t, w.Body.String(), tc.responseBody)
			}

			if tc.appGenerateImage {
				app.AssertExpectations(t)
			}
		})
	}
}

func TestGetImages(t *testing.T) {
	testCases := map[string]struct {
		filter   *dmodel.ReleaseOrImageFilter
		images   []*model.Image
		appError error
		checker  mt.ResponseChecker
	}{
		"ok": {
			filter: &dmodel.ReleaseOrImageFilter{},
			images: []*dmodel.Image{
				{
					Id:   "1",
					Size: 1000,
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{
					{
						Id:   "1",
						Size: 1000,
					},
				},
			),
		},
		"ok, empty": {
			filter: &dmodel.ReleaseOrImageFilter{},
			images: []*dmodel.Image{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{},
			),
		},
		"ok, filter": {
			filter: &dmodel.ReleaseOrImageFilter{Name: "foo"},
			images: []*dmodel.Image{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{},
			),
		},
		"error: generic": {
			filter:   &dmodel.ReleaseOrImageFilter{},
			images:   []*dmodel.Image{},
			appError: errors.New("database error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				deployments_testing.RestError("internal error"),
			),
		},
	}

	for name := range testCases {
		tc := testCases[name]

		t.Run(name, func(t *testing.T) {
			restView := new(view.RESTView)
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			app.On("ListImages",
				deployments_testing.ContextMatcher(),
				tc.filter,
			).Return(tc.images, len(tc.images), tc.appError)

			c := NewDeploymentsApiHandlers(nil, restView, app)
			router := setUpTestRouter()
			router.GET("/api/management/v1/artifacts", c.GetImages)

			reqUrl := "http://1.2.3.4/api/management/v1/artifacts"

			if tc.filter != nil {
				reqUrl += "?name=" + tc.filter.Name
			}

			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   reqUrl,
			})

			recorded := restutil.RunRequest(t, router, req)

			mt.CheckHTTPResponse(t, tc.checker, recorded)
		})
	}
}

func TestListImages(t *testing.T) {
	testCases := map[string]struct {
		filter   *dmodel.ReleaseOrImageFilter
		images   []*model.Image
		appError error
		checker  mt.ResponseChecker
	}{
		"ok": {
			filter: &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			images: []*dmodel.Image{
				{
					Id:   "1",
					Size: 1000,
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{
					{
						Id:   "1",
						Size: 1000,
					},
				},
			),
		},
		"ok, empty": {
			filter: &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			images: []*dmodel.Image{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{},
			),
		},
		"ok, filter": {
			filter: &dmodel.ReleaseOrImageFilter{Name: "foo", Page: 1, PerPage: 20},
			images: []*dmodel.Image{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{},
			),
		},
		"error: generic": {
			filter:   &dmodel.ReleaseOrImageFilter{Page: 1, PerPage: 20},
			images:   []*dmodel.Image{},
			appError: errors.New("database error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				deployments_testing.RestError("internal error"),
			),
		},
	}

	for name := range testCases {
		tc := testCases[name]

		t.Run(name, func(t *testing.T) {
			restView := new(view.RESTView)
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			app.On("ListImages",
				deployments_testing.ContextMatcher(),
				tc.filter,
			).Return(tc.images, len(tc.images), tc.appError)

			c := NewDeploymentsApiHandlers(nil, restView, app)
			router := setUpTestRouter()
			router.GET("/api/management/v1/artifacts/list", c.ListImages)

			reqUrl := "http://1.2.3.4/api/management/v1/artifacts/list"

			if tc.filter != nil {
				reqUrl += "?name=" + tc.filter.Name
			}

			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   reqUrl,
			})

			recorded := restutil.RunRequest(t, router, req)

			mt.CheckHTTPResponse(t, tc.checker, recorded)
		})
	}
}

func TestListImagesV2(t *testing.T) {
	testCases := map[string]struct {
		filter   *dmodel.ImageFilter
		images   []*model.Image
		appError error
		checker  mt.ResponseChecker
	}{
		"ok": {
			filter: &dmodel.ImageFilter{Page: 1, PerPage: 20, Limit: 21},
			images: []*dmodel.Image{
				{
					Id:   "1",
					Size: 1000,
				},
			},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{
					{
						Id:   "1",
						Size: 1000,
					},
				},
			),
		},
		"ok, empty": {
			filter: &dmodel.ImageFilter{Page: 1, PerPage: 20, Limit: 21},
			images: []*dmodel.Image{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{},
			),
		},
		"ok, filter": {
			filter: &dmodel.ImageFilter{
				ExactNames: []string{"foo"},
				Page:       1,
				PerPage:    20,
				Limit:      21,
			},
			images: []*dmodel.Image{},
			checker: mt.NewJSONResponse(
				http.StatusOK,
				nil,
				[]*dmodel.Image{},
			),
		},
		"error: generic": {
			filter:   &dmodel.ImageFilter{Page: 1, PerPage: 20, Limit: 21},
			images:   []*dmodel.Image{},
			appError: errors.New("database error"),
			checker: mt.NewJSONResponse(
				http.StatusInternalServerError,
				nil,
				deployments_testing.RestError("internal error"),
			),
		},
	}

	for name := range testCases {
		tc := testCases[name]

		t.Run(name, func(t *testing.T) {
			restView := new(view.RESTView)
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			app.On("ListImagesV2",
				deployments_testing.ContextMatcher(),
				tc.filter,
			).Return(tc.images, tc.appError)

			c := NewDeploymentsApiHandlers(nil, restView, app)
			router := setUpTestRouter()
			router.GET("/api/management/v2/artifacts", c.ListImagesV2)

			reqUrl := "http://1.2.3.4/api/management/v2/artifacts"

			if tc.filter != nil && tc.filter.ExactNames != nil {
				reqUrl += "?"
				for i, name := range tc.filter.ExactNames {
					reqUrl += "name=" + name
					if i != len(tc.filter.ExactNames)-1 {
						reqUrl += "&"
					}
				}
			}

			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   reqUrl,
			})

			recorded := restutil.RunRequest(t, router, req)

			mt.CheckHTTPResponse(t, tc.checker, recorded)
		})
	}
}
