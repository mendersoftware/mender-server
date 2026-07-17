//nolint:all // This is all test code
package opensource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-artifact/areader"
	"github.com/mendersoftware/mender-artifact/handlers"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	depmodel "github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/require"
)

func (s *DeploymentsManagementV1Suite) TestShowArtifactSize() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := common.DefaultDeviceType
	artifactName := "deployments-artifact-" + uuid.NewString()

	// one incompatible device, one already-installed, two that install (two,
	// not one, to verify the deployment-size math).
	devs, err := s.makeAcceptedDevices(ctx, 4)
	require.NoError(err)

	// clear CreateArtifact's placeholder {"foo":"bar"} depends: the generated
	// ArtifactV1 model types artifact_depends values as []string, but a scalar
	// depends value like that fails to unmarshal from the real API response.
	artifact, err := common.CreateArtifact(artifactName, s.T(),
		common.WithCompatibleDevices([]string{deviceType}),
		common.WithDependsProvides(map[string]string{}, map[string]string{}),
	)
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

	artifacts, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListArtifacts(ctx).
		Name(artifactName).Execute()
	require.NoError(err)
	require.Len(artifacts, 1)
	artifactSize := artifacts[0].GetSize()

	deviceIDs := make([]string, len(devs))
	for i, d := range devs {
		deviceIDs[i] = d.ID
	}
	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      deviceIDs,
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)
	loc, err := r.Location()
	require.NoError(err)
	depID := path.Base(loc.Path)

	dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
	require.NoError(err)
	require.NotNil(dep.Statistics)
	require.EqualValues(0, dep.Statistics.GetTotalSize())

	// no artifact for the first device (incompatible device type)
	_, r, err = s.APIClient.DeploymentsDeviceAPIAPI.
		CheckUpdate(common.JWTAuthContext(ctx, devs[0].Token)).
		ArtifactName("bugs-bunny").DeviceType("foo").Execute()
	require.NoError(err)
	require.Equal(http.StatusNoContent, r.StatusCode)

	// second device already has the artifact installed
	_, r, err = s.APIClient.DeploymentsDeviceAPIAPI.
		CheckUpdate(common.JWTAuthContext(ctx, devs[1].Token)).
		ArtifactName(artifactName).DeviceType(deviceType).Execute()
	require.NoError(err)
	require.Equal(http.StatusNoContent, r.StatusCode)

	// the rest are offered the artifact and report it successfully installed.
	for _, dev := range devs[2:] {
		devCtx := common.JWTAuthContext(ctx, dev.Token)
		instr, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(devCtx).ArtifactName("bugs-bunny").DeviceType(deviceType).Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
		_, err = s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, instr.GetId()).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusSuccessStr}).Execute()
		require.NoError(err)
	}

	dep, _, err = s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
	require.NoError(err)
	require.EqualValues(int32(len(devs)-2)*artifactSize, dep.Statistics.GetTotalSize())

	ddevs, _, err := s.APIClient.DeploymentsManagementAPIAPI.
		ListAllDevicesInDeployment(ctx, depID).Execute()
	require.NoError(err)
	byID := make(map[string]client.DeviceWithImage, len(ddevs))
	for _, d := range ddevs {
		byID[d.Id] = d
	}

	// first device got no artifact, so no image
	dd0, ok := byID[devs[0].ID]
	require.True(ok)
	require.Nil(dd0.Image)

	for _, dev := range devs[1:] {
		dd, ok := byID[dev.ID]
		require.True(ok)
		require.NotNil(dd.Image)
		require.EqualValues(artifactSize, dd.Image.GetSize())
	}
}

func (s *DeploymentsManagementV1Suite) TestListArtifactsV2() {
	// Every artifact created here is uuid-prefixed and the listing steps
	// filter by that prefix (name=<prefix>*), so the assertions hold on the
	// shared, never-reset environment.
	//
	// The four artifacts are given descriptions "abc", "abc", "abc-0" and
	// "abc-0-1" (each built from the previous, not from a fixed base), so the
	// description-prefix assertions ("abc-*") match the last two.
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	p := uuid.NewString()
	descBase := "abc-" + p

	type testArtifact struct {
		name        string
		description string
		deviceType  string
	}
	artifacts := []testArtifact{
		{name: p + "-deployments-artifact", description: descBase, deviceType: common.DefaultDeviceType + "-" + p},
		{name: p + "-test-artifact", description: descBase, deviceType: "arm1-" + p},
		{name: p + "-test-artifact-0", description: descBase + "-0", deviceType: "arm1-" + p + "-0"},
		{name: p + "-test-artifact-1", description: descBase + "-0-1", deviceType: "arm1-" + p + "-1"},
	}
	for _, a := range artifacts {
		f, err := common.CreateArtifact(a.name, s.T(),
			common.WithCompatibleDevices([]string{a.deviceType}),
			common.WithDependsProvides(map[string]string{}, map[string]string{}),
		)
		require.NoError(s.T(), err)
		_, err = s.uploadArtifact(ctx, f, a.description)
		require.NoError(s.T(), err)
	}

	apiV2 := s.APIClient.DeploymentsV2ManagementAPIAPI

	s.Run("Success/ListAll", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Name([]string{p + "*"}).Execute()
		require.NoError(err)
		require.Len(got, 4)
		for _, name := range []string{artifacts[0].name, artifacts[1].name, artifacts[2].name, artifacts[3].name} {
			require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return name == a.GetName() }))
		}
	})

	s.Run("Success/ExactName", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Name([]string{artifacts[1].name}).Execute()
		require.NoError(err)
		require.Len(got, 1)
		require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return artifacts[1].name == a.GetName() }))
	})

	s.Run("Success/MultipleExactNames", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{artifacts[1].name, artifacts[2].name}).Execute()
		require.NoError(err)
		require.Len(got, 2)
		for _, name := range []string{artifacts[1].name, artifacts[2].name} {
			require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return name == a.GetName() }))
		}
	})

	s.Run("Success/NamePrefix", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{artifacts[1].name + "-*"}).Execute()
		require.NoError(err)
		require.Len(got, 2)
		for _, name := range []string{artifacts[2].name, artifacts[3].name} {
			require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return name == a.GetName() }))
		}
	})

	s.Run("Failure/NamePrefixAndExactName", func() {
		require := require.New(s.T())
		_, r, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{artifacts[0].name, artifacts[1].name + "-*"}).Execute()
		require.Error(err)
		require.Equal(http.StatusBadRequest, r.StatusCode)
	})

	s.Run("Failure/MultiplePrefixes", func() {
		require := require.New(s.T())
		_, r, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{p + "*", artifacts[1].name + "-*"}).Execute()
		require.Error(err)
		require.Equal(http.StatusBadRequest, r.StatusCode)
	})

	s.Run("Success/ExactDescription", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Description(descBase).Execute()
		require.NoError(err)
		require.Len(got, 2)
		for _, name := range []string{artifacts[0].name, artifacts[1].name} {
			require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return name == a.GetName() }))
		}
	})

	s.Run("Success/DescriptionPrefix", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Description(descBase + "-*").Execute()
		require.NoError(err)
		require.Len(got, 2)
		for _, name := range []string{artifacts[2].name, artifacts[3].name} {
			require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return name == a.GetName() }))
		}
	})

	s.Run("Success/ExactDeviceType", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).DeviceType(artifacts[1].deviceType).Execute()
		require.NoError(err)
		require.Len(got, 1)
		require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return artifacts[1].name == a.GetName() }))
	})

	s.Run("Success/DeviceTypePrefix", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			DeviceType(artifacts[1].deviceType + "-*").Execute()
		require.NoError(err)
		require.Len(got, 2)
		for _, name := range []string{artifacts[2].name, artifacts[3].name} {
			require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return name == a.GetName() }))
		}
	})

	s.Run("Success/Pagination", func() {
		require := require.New(s.T())
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{p + "*"}).Page(2).PerPage(3).Execute()
		require.NoError(err)
		require.Len(got, 1)
		require.True(slices.ContainsFunc(got, func(a client.ArtifactV2) bool { return artifacts[3].name == a.GetName() }))
	})
}

func (s *DeploymentsManagementV1Suite) TestDeleteReleases() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := common.DefaultDeviceType
	artifactName := "deployments-artifact-" + uuid.NewString()

	devs, err := s.makeAcceptedDevices(ctx, 1)
	require.NoError(err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

	deviceIDs := []string{devs[0].ID}
	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      deviceIDs,
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)

	// can't delete the release while it's part of an active deployment
	r, err = s.APIClient.DeploymentsV2ManagementAPIAPI.DeleteReleases(ctx).
		Name(artifactName).Execute()
	require.Error(err)
	require.Equal(http.StatusConflict, r.StatusCode)

	// finish the deployment: the device is offered nothing (incompatible
	// device type), moving the single-device deployment to finished.
	_, r, err = s.APIClient.DeploymentsDeviceAPIAPI.
		CheckUpdate(common.JWTAuthContext(ctx, devs[0].Token)).
		ArtifactName("bugs-bunny").DeviceType("foo").Execute()
	require.NoError(err)
	require.Equal(http.StatusNoContent, r.StatusCode)

	// now the release can be deleted
	r, err = s.APIClient.DeploymentsV2ManagementAPIAPI.DeleteReleases(ctx).
		Name(artifactName).Execute()
	require.NoError(err)
	require.Equal(http.StatusNoContent, r.StatusCode)
}

func (s *DeploymentsManagementV1Suite) TestUploadArtifact() {
	s.Run("Success/MissingTypeInfo", s.testUploadArtifactMissingTypeInfo)
	s.Run("Success/SelectionAlreadyInstalled", s.testUploadArtifactSelectionAlreadyInstalled)
	s.Run("Success/DependsProvidesValid", s.testUploadArtifactDependsProvidesValid)
	s.Run("Success/ProvidesDependsIgnoredInOpenSource", s.testUploadArtifactProvidesDependsIgnoredInOpenSource)
}

func (s *DeploymentsManagementV1Suite) testUploadArtifactMissingTypeInfo() {
	// Ensuring server backwards compatibility with artifacts created with
	// older versions of mender-artifact. Ref:
	// 	 https://github.com/mendersoftware/mender-artifact/commit/c25764218c6e48677ab0b5b1736ccaf531c62e42
	var (
		ctx     = common.JWTAuthContext(s.T().Context(), s.JWT)
		require = require.New(s.T())
	)
	file, err := os.Open("../data/missing-type-info.mender")
	require.NoError(err)
	r, err := s.APIClient.DeploymentsManagementAPIAPI.
		UploadArtifact(ctx).
		Artifact(file).
		Execute()
	require.NoError(err)
	require.NotNil(r)
	require.Equal(http.StatusCreated, r.StatusCode)
}

func (s *DeploymentsManagementV1Suite) testUploadArtifactSelectionAlreadyInstalled() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	artifactName := "test-" + uuid.NewString()
	deviceType := "arm1-" + uuid.NewString()

	a, err := common.CreateArtifact(artifactName, s.T(),
		common.WithCompatibleDevices([]string{deviceType}),
		common.WithDependsProvides(
			map[string]string{"rootfs-image.checksum": "another-checksum"},
			map[string]string{"rootfs-image.checksum": "provided"},
		),
	)
	require.NoError(err)

	_, err = s.uploadArtifact(ctx, a, "description")
	require.NoError(err)
	dev, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      []string{dev.ID},
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)

	_, r, err = s.APIClient.DeploymentsDeviceAPIAPI.
		CheckUpdate(common.JWTAuthContext(ctx, dev.Token)).
		ArtifactName(artifactName).DeviceType(deviceType).Execute()
	require.NoError(err)
	require.Equal(http.StatusNoContent, r.StatusCode)
}

func (s *DeploymentsManagementV1Suite) testUploadArtifactDependsProvidesValid() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	artifactName := "test-" + uuid.NewString()
	deviceType := "arm1-" + uuid.NewString()

	a, err := common.CreateArtifact(artifactName, s.T(),
		common.WithModuleImage(handlers.NewModuleImage("dummy")),
		common.WithCompatibleDevices([]string{deviceType}),
		common.WithDependsProvides(
			map[string]string{"key1": "value1", "key2": "value2"},
			map[string]string{
				"key3": "value3", "key4": "value4", "key5": "value5",
				// the mender-artifact CLI adds this default software-version
				// provides entry; the Go library writer does not, so add it
				// explicitly to keep the backend-side assertions identical.
				"rootfs-image.dummy.version": artifactName,
			},
		),
	)
	require.NoError(err)

	id, err := s.uploadArtifact(ctx, a, "description")
	require.NoError(err)

	// TODO: possible spec bug (tracked separately). The deployments API spec
	// types artifact_depends values as []string, but the server returns a
	// scalar string for custom (non-device_type) depends keys, so the generated
	// ShowArtifact client method fails to unmarshal them. Until that's resolved,
	// fetch and decode the artifact ourselves.
	got, err := s.showArtifactRaw(ctx, id)
	require.NoError(err)

	require.Equal("description", got.Description)
	require.Equal(artifactName, got.Name)
	require.Equal("mender", got.Info.Format)
	require.EqualValues(3, got.Info.Version)
	require.False(got.Signed)
	require.Len(got.Updates, 1)
	require.Greater(got.Size, int64(0))
	require.NotEmpty(got.Id)
	require.False(got.Modified.IsZero())

	require.ElementsMatch([]string{deviceType}, toStringSlice(got.ArtifactDepends["device_type"]))
	require.Equal("value1", got.ArtifactDepends["key1"])
	require.Equal("value2", got.ArtifactDepends["key2"])

	require.Equal(artifactName, got.ArtifactProvides["artifact_name"])
	require.Equal("value3", got.ArtifactProvides["key3"])
	require.Equal("value4", got.ArtifactProvides["key4"])
	require.Equal("value5", got.ArtifactProvides["key5"])
	require.Equal(artifactName, got.ArtifactProvides["rootfs-image.dummy.version"])
}

func (s *DeploymentsManagementV1Suite) testUploadArtifactProvidesDependsIgnoredInOpenSource() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	artifactName := "test-" + uuid.NewString()
	deviceType := "arm1-" + uuid.NewString()

	small, err := common.CreateArtifact(artifactName, s.T(),
		common.WithCompatibleDevices([]string{deviceType}),
		common.WithPayloadFile(s.T(), 256),
	)
	require.NoError(err)
	big, err := common.CreateArtifact(artifactName+"-big", s.T(),
		common.WithCompatibleDevices([]string{deviceType}),
		common.WithDependsProvides(
			map[string]string{"foo": "fooval", "bar": "barval"}, nil,
		),
		common.WithPayloadFile(s.T(), 1024),
	)
	require.NoError(err)

	for _, f := range []*os.File{small, big} {
		_, err = s.uploadArtifact(ctx, f, "description")
		require.NoError(err)
	}
	dev, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      []string{dev.ID},
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)

	instr, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
		CheckUpdate(common.JWTAuthContext(ctx, dev.Token)).
		ArtifactName("bugs-bunny").DeviceType(deviceType).Execute()
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)

	uri := instr.Artifact.Source.GetUri()
	require.NotEmpty(uri)

	resp, err := s.APIClient.GetConfig().HTTPClient.Get(uri)
	require.NoError(err)
	defer resp.Body.Close()
	require.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(err)

	// if provides/depends weren't ignored on open source, the matching, larger
	// artifact would have been selected; instead the smallest of all matching
	// artifacts is picked, so verify the downloaded payload is the small one.
	ar := areader.NewReader(bytes.NewReader(body))
	require.NoError(ar.RegisterHandler(handlers.NewModuleImage("foo")))
	require.NoError(ar.ReadArtifact())
	// RegisterHandler's argument is only used as a template: the reader builds
	// its own installer instance (via NewInstance) to hold the parsed data, so
	// the populated files must be read back from GetHandlers(), not from the
	// handler passed to RegisterHandler.
	installer, ok := ar.GetHandlers()[0]
	require.True(ok)
	files := installer.GetUpdateFiles()
	require.Len(files, 1)
	require.EqualValues(256, files[0].Size)
}

func (s *DeploymentsManagementV1Suite) TestCreateArtifact() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	name := "artifact-" + uuid.NewString()
	deviceType := "device-type-" + uuid.NewString()

	f, err := os.Create(path.Join(s.T().TempDir(), "mender-create-artifact"))
	require.NoError(err)
	_, err = f.WriteString("#!/bin/bash\ntrue\n")
	require.NoError(err)
	_, err = f.Seek(0, io.SeekStart)
	require.NoError(err)

	_, err = s.APIClient.DeploymentsManagementAPIAPI.GenerateArtifact(ctx).
		Name(name).
		Description("description").
		Type_("single_file").
		DeviceTypesCompatible([]string{deviceType}).
		Args(`{"filename":"run.sh","dest_dir":"/tests"}`).
		File(f).
		Execute()
	require.NoError(err)

	var id string
	err = common.RetryUntil(ctx, 15*time.Second, time.Second, func() (bool, error) {
		release, _, err := s.APIClient.DeploymentsV2ManagementAPIAPI.
			GetReleaseWithGivenName(ctx, name).Execute()
		if err == nil && release != nil && len(release.Artifacts) > 0 {
			id = release.Artifacts[0].Id
			return true, nil
		}
		return false, nil
	})
	require.NoError(err, "artifact was not generated in time")

	got, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowArtifact(ctx, id).Execute()
	require.NoError(err)

	require.Equal("description", got.GetDescription())
	require.Equal(name, got.Name)
	require.NotNil(got.Info)
	require.Equal("mender", got.Info.GetFormat())
	require.EqualValues(3, got.Info.GetVersion())
	require.False(got.GetSigned())
	require.Len(got.Updates, 1)
	require.Greater(got.GetSize(), int32(0))
	require.NotEmpty(got.Id)
	require.False(got.Modified.IsZero())

	// download the generated artifact and verify its structure with the
	// mender-artifact library, beyond what the API metadata alone confirms.
	link, _, err := s.APIClient.DeploymentsManagementAPIAPI.DownloadArtifact(ctx, id).Execute()
	require.NoError(err)
	resp, err := s.APIClient.GetConfig().HTTPClient.Get(link.Uri)
	require.NoError(err)
	defer resp.Body.Close()
	require.Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(err)

	ar := areader.NewReader(bytes.NewReader(body))
	require.NoError(ar.ReadArtifact())
	require.Equal(name, ar.GetArtifactName())
	require.Contains(ar.GetCompatibleDevices(), deviceType)
}

// rawArtifact decodes the artifact management API's response directly,
// bypassing the generated ArtifactV1 model, which cannot represent a
// scalar-valued custom artifact_depends entry.
type rawArtifact struct {
	Id                    string   `json:"id"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	DeviceTypesCompatible []string `json:"device_types_compatible"`
	Info                  struct {
		Format  string `json:"format"`
		Version int32  `json:"version"`
	} `json:"info"`
	Signed           bool              `json:"signed"`
	Updates          []any             `json:"updates"`
	ArtifactProvides map[string]string `json:"artifact_provides"`
	ArtifactDepends  map[string]any    `json:"artifact_depends"`
	Size             int64             `json:"size"`
	Modified         time.Time         `json:"modified"`
}

func (s *DeploymentsManagementV1Suite) showArtifactRaw(ctx context.Context, id string) (*rawArtifact, error) {
	resp, err := common.RawRequest(ctx, s.APIClient, http.MethodGet,
		"/api/management/v1/deployments/artifacts/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected show artifact status: %d", resp.StatusCode)
	}

	var out rawArtifact
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// toStringSlice normalizes a JSON-decoded depends value (which may be a single
// string or a list of strings) into a []string.
func toStringSlice(v any) []string {
	switch val := v.(type) {
	case string:
		return []string{val}
	case []any:
		out := make([]string, len(val))
		for i, e := range val {
			out[i], _ = e.(string)
		}
		return out
	default:
		return nil
	}
}
