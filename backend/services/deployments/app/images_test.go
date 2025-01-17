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

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-artifact/areader"
	"github.com/mendersoftware/mender-artifact/artifact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/identity"

	workflows_mocks "github.com/mendersoftware/mender-server/services/deployments/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	h "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
)

type BogusReader struct{}

func (r *BogusReader) Read(b []byte) (int, error) {
	return len(b), nil
}

func TestGenerateImageError(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)

	testCases := []struct {
		multipartGenerateImage *model.MultipartGenerateImageMsg
		expectedError          error
	}{
		{
			multipartGenerateImage: nil,
			expectedError:          ErrModelMultipartUploadMsgMalformed,
		},
	}

	ctx := context.Background()
	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			artifactID, err := d.GenerateImage(ctx, tc.multipartGenerateImage)

			assert.Equal(t, artifactID, "")
			assert.Error(t, err)
			assert.EqualError(t, err, tc.expectedError.Error())
		})
	}
}

func TestGenerateImageArtifactIsNotUnique(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(false, nil)

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            nil,
	}

	ctx := context.Background()
	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, ErrModelArtifactNotUnique.Error())

	db.AssertExpectations(t)
}

func TestGenerateImageErrorWhileCheckingIfArtifactIsNotUnique(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(false, errors.New("error"))

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            nil,
	}

	ctx := context.Background()
	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "Fail to check if artifact is unique: error")

	db.AssertExpectations(t)
}

func TestGenerateImageErrorWhileUploading(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)
	ctx := context.Background()

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(errors.New("error while uploading"))

	db.On("GetStorageSettings",
		ctx,
	).Return(nil, nil)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "error while uploading")

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
}

func TestGenerateImageErrorS3GetRequest(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)
	ctx := context.Background()

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(nil)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	db.On("GetStorageSettings",
		ctx,
	).Return(nil, nil)

	fs.On("GetRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(nil, errors.New("error get request")).
		On("DeleteObject",
			h.ContextMatcher(),
			mock.AnythingOfType("string")).
		Return(errors.New("error!"))

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "error get request")

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
}

func TestGenerateImageErrorS3DeleteRequest(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)
	ctx := context.Background()

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(nil)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	db.On("GetStorageSettings",
		ctx,
	).Return(nil, nil)

	fs.On("GetRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "GET",
	}, nil)

	fs.On("DeleteRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(nil, errors.New("error delete request")).
		On("DeleteObject",
			h.ContextMatcher(),
			mock.AnythingOfType("string")).
		Return(errors.New("error!"))

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "error delete request")

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
}

func TestGenerateImageErrorWhileStartingWorkflow(t *testing.T) {
	generateErr := errors.New("failed to start workflow: generate_artifact")
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)
	ctx := context.Background()

	fs.On("GetRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "GET",
	}, nil)

	fs.On("DeleteRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "DELETE",
	}, nil)

	workflowsClient := &workflows_mocks.Client{}
	workflowsClient.On("StartGenerateArtifact",
		h.ContextMatcher(),
		mock.AnythingOfType("*model.MultipartGenerateImageMsg"),
	).Return(generateErr)
	d.SetWorkflowsClient(workflowsClient)

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(nil)

	fs.On("DeleteObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
	).Return(nil)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	db.On("GetStorageSettings",
		ctx,
	).Return(nil, nil)

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, generateErr.Error())

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
	workflowsClient.AssertExpectations(t)
}

func TestGenerateImageErrorWhileStartingWorkflowAndFailsWhenCleaningUp(t *testing.T) {
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)
	ctx := context.Background()

	workflowsClient := &workflows_mocks.Client{}
	d.SetWorkflowsClient(workflowsClient)

	workflowsClient.On("StartGenerateArtifact",
		h.ContextMatcher(),
		mock.AnythingOfType("*model.MultipartGenerateImageMsg"),
	).Return(errors.New("failed to start workflow: generate_artifact"))

	fs.On("GetRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "GET",
	}, nil)

	fs.On("DeleteRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "DELETE",
	}, nil)

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(nil)

	fs.On("DeleteObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
	).Return(errors.New("unable to remove the file"))

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	db.On("GetStorageSettings",
		ctx,
	).Return(nil, nil)

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.Equal(t, artifactID, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "unable to remove the file: failed to start workflow: generate_artifact")

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
	workflowsClient.AssertExpectations(t)
}

func TestGenerateImageSuccessful(t *testing.T) {
	ctx := context.Background()
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "args",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	workflowsClient := &workflows_mocks.Client{}
	d.SetWorkflowsClient(workflowsClient)

	workflowsClient.On("StartGenerateArtifact",
		h.ContextMatcher(),
		multipartGenerateImage,
	).Return(nil)

	fs.On("GetRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "GET",
	}, nil)

	fs.On("DeleteRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "DELETE",
	}, nil)

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(nil)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	db.On("GetStorageSettings",
		ctx,
	).Return(nil, nil)

	artifactID, err := d.GenerateImage(ctx, multipartGenerateImage)

	assert.NotEqual(t, artifactID, "")
	assert.Nil(t, err)

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
	workflowsClient.AssertExpectations(t)
}

func TestGenerateImageSuccessfulWithTenant(t *testing.T) {
	ctx := context.Background()
	db := mocks.DataStore{}
	fs := &fs_mocks.ObjectStorage{}
	d := NewDeployments(&db, fs, 0, false)

	multipartGenerateImage := &model.MultipartGenerateImageMsg{
		Name:                  "name",
		Description:           "description",
		DeviceTypesCompatible: []string{"Beagle Bone"},
		Type:                  "single_file",
		Args:                  "args",
		FileReader:            bytes.NewReader([]byte("123456790")),
	}

	workflowsClient := &workflows_mocks.Client{}
	d.SetWorkflowsClient(workflowsClient)
	workflowsClient.On("StartGenerateArtifact",
		h.ContextMatcher(), multipartGenerateImage,
	).Return(nil)

	fs.On("GetRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "GET",
	}, nil)

	fs.On("DeleteRequest",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Duration"),
		false,
	).Return(&model.Link{
		Uri: "DELETE",
	}, nil)

	fs.On("PutObject",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*bytes.Reader"),
	).Return(nil)

	db.On("IsArtifactUnique",
		h.ContextMatcher(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]string"),
	).Return(true, nil)

	identityObject := &identity.Identity{Tenant: "tenant_id"}
	ctxWithIdentity := identity.WithContext(ctx, identityObject)

	db.On("GetStorageSettings",
		ctxWithIdentity,
	).Return(nil, nil)

	artifactID, err := d.GenerateImage(ctxWithIdentity, multipartGenerateImage)

	assert.NotEqual(t, artifactID, "")
	assert.Nil(t, err)

	db.AssertExpectations(t)
	fs.AssertExpectations(t)
	workflowsClient.AssertExpectations(t)
}

func TestGenerateConfigurationImage(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		DeviceType   string
		DeploymentID string

		StoreError error
		Deployment *model.Deployment

		Error error
	}{{
		Name: "ok",

		DeviceType: "strawberryPlanck",
		DeploymentID: uuid.NewSHA1(
			uuid.NameSpaceOID,
			[]byte("deployment"),
		).String(),
		Deployment: &model.Deployment{
			Id: uuid.NewSHA1(
				uuid.NameSpaceOID,
				[]byte("deployment"),
			).String(),
			Type:          model.DeploymentTypeConfiguration,
			Configuration: []byte("{\"foo\":\"bar\"}"),
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "spicyDeployment",
				ArtifactName: "spicyPi",
			},
		},
	}, {
		Name: "error, internal DataStore error",

		DeviceType: "strawberryPlanck",
		DeploymentID: uuid.NewSHA1(
			uuid.NameSpaceOID,
			[]byte("deployment"),
		).String(),
		StoreError: errors.New("internal error"),
		Error:      errors.New("internal error"),
	}, {
		Name: "error, deployment not found",

		DeviceType: "strawberryPlanck",
		DeploymentID: uuid.NewSHA1(
			uuid.NameSpaceOID,
			[]byte("deployment"),
		).String(),
		Error: ErrModelDeploymentNotFound,
	}, {
		Name: "error, invalid JSON metadata",

		DeviceType: "strawberryPlanck",
		DeploymentID: uuid.NewSHA1(
			uuid.NameSpaceOID,
			[]byte("deployment"),
		).String(),
		Deployment: &model.Deployment{
			Id: uuid.NewSHA1(
				uuid.NameSpaceOID,
				[]byte("deployment"),
			).String(),
			Type:          model.DeploymentTypeConfiguration,
			Configuration: []byte("gotcha"),
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "spicyDeployment",
				ArtifactName: "spicyPi",
			},
		},
		Error: errors.New(
			"malformed configuration in deployment: " +
				"invalid character 'g' looking for beginning of value",
		),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := new(mocks.DataStore)
			defer ds.AssertExpectations(t)
			ds.On("FindDeploymentByID", ctx, tc.DeploymentID).
				Return(tc.Deployment, tc.StoreError)
			d := NewDeployments(ds, nil, 0, false)
			artieFact, err := d.GenerateConfigurationImage(
				ctx, tc.DeviceType, tc.DeploymentID,
			)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				artieReader := areader.NewReader(artieFact)
				err := artieReader.ReadArtifactHeaders()
				if !assert.NoError(t, err, "Generated artifact is invalid") {
					t.FailNow()
				}
				assert.Equal(t,
					[]artifact.UpdateType{{Type: &ArtifactConfigureType}},
					artieReader.GetUpdates(),
				)
				provides, _ := artieReader.MergeArtifactProvides()
				if assert.Contains(t,
					provides,
					ArtifactConfigureProvides,
				) {
					assert.Equal(t,
						tc.Deployment.ArtifactName,
						provides[ArtifactConfigureProvides],
					)
				}
				depends, _ := artieReader.MergeArtifactDepends()
				if assert.Contains(t, depends, "device_type") {
					deviceTypes := []interface{}{tc.DeviceType}
					assert.Equal(t,
						deviceTypes,
						depends["device_type"],
					)
				}
				handlers := artieReader.GetHandlers()
				if assert.Len(t, handlers, 1) && assert.Contains(t, handlers, 0) {
					handler := handlers[0]
					metadata, _ := handler.GetUpdateMetaData()
					var actual map[string]interface{}
					//nolint:errcheck
					json.Unmarshal(tc.Deployment.Configuration, &actual)
					assert.Equal(t, metadata, actual)
				}
			}
		})
	}
}
