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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-artifact/areader"
	"github.com/mendersoftware/mender-artifact/handlers"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *DeploymentsManagementV1Suite) TestShowArtifactSize() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := "qemux86-64"
	artifactName := "deployments-phase-testing-" + uuid.NewString()

	devs, err := s.makeAcceptedDevices(ctx, 10)
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
	require.NotNil(artifacts[0].Size)
	artifactSize := *artifacts[0].Size

	deviceIDs := make([]string, len(devs))
	for i, d := range devs {
		deviceIDs[i] = d.ID
	}
	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "phased-deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      deviceIDs,
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)
	depID := path.Base(r.Header.Get("Location"))

	dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
	require.NoError(err)
	require.NotNil(dep.Statistics)
	require.NotNil(dep.Statistics.TotalSize)
	assert.EqualValues(0, *dep.Statistics.TotalSize)

	// no artifact for the first device
	code, err := s.tryUpdate(ctx, devs[0], "bugs-bunny", "foo")
	require.NoError(err)
	assert.Equal(http.StatusNoContent, code)

	// second device already has the artifact installed
	code, err = s.tryUpdate(ctx, devs[1], artifactName, deviceType)
	require.NoError(err)
	assert.Equal(http.StatusNoContent, code)

	// the rest perform a normal update
	for _, dev := range devs[2:] {
		code, err := s.tryUpdate(ctx, dev, "bugs-bunny", deviceType)
		require.NoError(err)
		assert.Equal(http.StatusOK, code)
	}

	dep, _, err = s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
	require.NoError(err)
	require.NotNil(dep.Statistics.TotalSize)
	assert.EqualValues(int64(len(devs)-2)*int64(artifactSize), *dep.Statistics.TotalSize)

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
	assert.Nil(dd0.Image)

	for _, dev := range devs[1:] {
		dd, ok := byID[dev.ID]
		require.True(ok)
		require.NotNil(dd.Image)
		require.NotNil(dd.Image.Size)
		assert.EqualValues(artifactSize, *dd.Image.Size)
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
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	p := uuid.NewString()
	descBase := "abc-" + p

	type testArtifact struct {
		name        string
		description string
		deviceType  string
	}
	artifacts := []testArtifact{
		{name: p + "-deployments-phase-testing", description: descBase, deviceType: "qemux86-64-" + p},
		{name: p + "-test-artifact", description: descBase, deviceType: "arm1-" + p},
		{name: p + "-test-artifact-0", description: descBase + "-0", deviceType: "arm1-" + p + "-0"},
		{name: p + "-test-artifact-1", description: descBase + "-0-1", deviceType: "arm1-" + p + "-1"},
	}
	for _, a := range artifacts {
		// clear CreateArtifact's placeholder {"foo":"bar"} depends: the
		// generated ArtifactV2 model types artifact_depends values as
		// []string, but a scalar depends value like that fails to unmarshal
		// from the real API response.
		f, err := common.CreateArtifact(a.name, s.T(),
			common.WithCompatibleDevices([]string{a.deviceType}),
			common.WithDependsProvides(map[string]string{}, map[string]string{}),
		)
		require.NoError(err)
		_, err = s.uploadArtifact(ctx, f, a.description)
		require.NoError(err)
	}

	apiV2 := s.APIClient.DeploymentsV2ManagementAPIAPI
	names := func(vs []client.ArtifactV2) []string {
		out := make([]string, len(vs))
		for i, v := range vs {
			out[i] = v.Name
		}
		return out
	}

	s.Run("Success/ListAll", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Name([]string{p + "*"}).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[0].name, artifacts[1].name, artifacts[2].name, artifacts[3].name}, names(got))
	})

	s.Run("Success/ExactName", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Name([]string{artifacts[1].name}).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[1].name}, names(got))
	})

	s.Run("Success/MultipleExactNames", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{artifacts[1].name, artifacts[2].name}).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[1].name, artifacts[2].name}, names(got))
	})

	s.Run("Success/NamePrefix", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{artifacts[1].name + "-*"}).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[2].name, artifacts[3].name}, names(got))
	})

	s.Run("Failure/NamePrefixAndExactName", func() {
		require, assert := s.Require(), s.Assert()
		_, r, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{artifacts[0].name, artifacts[1].name + "-*"}).Execute()
		require.Error(err)
		assert.Equal(http.StatusBadRequest, r.StatusCode)
	})

	s.Run("Failure/MultiplePrefixes", func() {
		require, assert := s.Require(), s.Assert()
		_, r, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{p + "*", artifacts[1].name + "-*"}).Execute()
		require.Error(err)
		assert.Equal(http.StatusBadRequest, r.StatusCode)
	})

	s.Run("Success/ExactDescription", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Description(descBase).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[0].name, artifacts[1].name}, names(got))
	})

	s.Run("Success/DescriptionPrefix", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).Description(descBase + "-*").Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[2].name, artifacts[3].name}, names(got))
	})

	s.Run("Success/ExactDeviceType", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).DeviceType(artifacts[1].deviceType).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[1].name}, names(got))
	})

	s.Run("Success/DeviceTypePrefix", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			DeviceType(artifacts[1].deviceType + "-*").Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[2].name, artifacts[3].name}, names(got))
	})

	s.Run("Success/Pagination", func() {
		require, assert := s.Require(), s.Assert()
		got, _, err := apiV2.DeploymentsV2ListArtifactsWithPagination(ctx).
			Name([]string{p + "*"}).Page(2).PerPage(3).Execute()
		require.NoError(err)
		assert.Equal([]string{artifacts[3].name}, names(got))
	})
}

func (s *DeploymentsManagementV1Suite) TestDeleteReleases() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := "qemux86-64"
	artifactName := "deployments-phase-testing-" + uuid.NewString()

	devs, err := s.makeAcceptedDevices(ctx, 1)
	require.NoError(err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

	deviceIDs := []string{devs[0].ID}
	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "phased-deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      deviceIDs,
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)

	// can't delete the release while it's part of an active deployment
	r, err = s.APIClient.DeploymentsV2ManagementAPIAPI.DeleteReleases(ctx).
		Name(artifactName).Execute()
	require.Error(err)
	assert.Equal(http.StatusConflict, r.StatusCode)

	// finish the deployment
	code, err := s.tryUpdate(ctx, devs[0], "bugs-bunny", "foo")
	require.NoError(err)
	assert.Equal(http.StatusNoContent, code)

	// now the release can be deleted
	r, err = s.APIClient.DeploymentsV2ManagementAPIAPI.DeleteReleases(ctx).
		Name(artifactName).Execute()
	require.NoError(err)
	assert.Equal(http.StatusNoContent, r.StatusCode)
}

func (s *DeploymentsManagementV1Suite) TestUploadArtifact() {
	s.Run("Success/SelectionAlreadyInstalled", s.testUploadArtifactSelectionAlreadyInstalled)
	s.Run("Success/DependsProvidesValid", s.testUploadArtifactDependsProvidesValid)
	s.Run("Success/ProvidesDependsIgnoredInOpenSource", s.testUploadArtifactProvidesDependsIgnoredInOpenSource)
}

func (s *DeploymentsManagementV1Suite) testUploadArtifactSelectionAlreadyInstalled() {
	require := require.New(s.T())
	assert := assert.New(s.T())
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

	dev, err := s.setupArtifactSelection(ctx, artifactName, []*os.File{a})
	require.NoError(err)

	_, r, err := dev.CheckUpdate(ctx, s.APIClient, artifactName, deviceType)
	require.NoError(err)
	assert.Equal(http.StatusNoContent, r.StatusCode)
}

func (s *DeploymentsManagementV1Suite) testUploadArtifactDependsProvidesValid() {
	require := require.New(s.T())
	assert := assert.New(s.T())
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

	// the generated ArtifactV1 model types artifact_depends values as
	// []string, but the real API returns a scalar string for custom
	// (non-device_type) depends keys, which fails to unmarshal through the
	// generated ShowArtifact client method. Fetch and decode the artifact
	// ourselves instead.
	got, err := s.showArtifactRaw(ctx, id)
	require.NoError(err)

	assert.Equal("description", got.Description)
	assert.Equal(artifactName, got.Name)
	assert.Equal("mender", got.Info.Format)
	assert.EqualValues(3, got.Info.Version)
	assert.False(got.Signed)
	assert.Len(got.Updates, 1)
	assert.Greater(got.Size, int64(0))
	assert.NotEmpty(got.Id)
	assert.False(got.Modified.IsZero())

	assert.ElementsMatch([]string{deviceType}, toStringSlice(got.ArtifactDepends["device_type"]))
	assert.Equal("value1", got.ArtifactDepends["key1"])
	assert.Equal("value2", got.ArtifactDepends["key2"])

	assert.Equal(artifactName, got.ArtifactProvides["artifact_name"])
	assert.Equal("value3", got.ArtifactProvides["key3"])
	assert.Equal("value4", got.ArtifactProvides["key4"])
	assert.Equal("value5", got.ArtifactProvides["key5"])
	assert.Equal(artifactName, got.ArtifactProvides["rootfs-image.dummy.version"])
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

	dev, err := s.setupArtifactSelection(ctx, artifactName, []*os.File{small, big})
	require.NoError(err)

	instr, r, err := dev.CheckUpdate(ctx, s.APIClient, "old-artifact", deviceType)
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)

	uri := instr.Artifact.Source.GetUri()
	require.NotEmpty(uri)

	resp, err := s.downloadHTTPClient().Get(uri)
	require.NoError(err)
	defer resp.Body.Close()
	require.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(err)

	// if provides/depends weren't ignored on open source, the matching,
	// larger artifact would have been selected; instead the smallest of
	// all matching artifacts is picked.
	assertArtifactPayloadSize(s.T(), body, 256)
}

func (s *DeploymentsManagementV1Suite) TestCreateArtifact() {
	require := require.New(s.T())
	assert := assert.New(s.T())
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

	assert.Equal("description", got.GetDescription())
	assert.Equal(name, got.Name)
	require.NotNil(got.Info)
	assert.Equal("mender", got.Info.GetFormat())
	assert.EqualValues(3, got.Info.GetVersion())
	assert.False(got.GetSigned())
	assert.Len(got.Updates, 1)
	require.NotNil(got.Size)
	assert.Greater(*got.Size, int32(0))
	assert.NotEmpty(got.Id)
	assert.False(got.Modified.IsZero())

	// download the generated artifact and verify its structure with the
	// mender-artifact library, beyond what the API metadata alone confirms.
	link, _, err := s.APIClient.DeploymentsManagementAPIAPI.DownloadArtifact(ctx, id).Execute()
	require.NoError(err)
	resp, err := s.downloadHTTPClient().Get(link.Uri)
	require.NoError(err)
	defer resp.Body.Close()
	require.Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(err)

	ar := areader.NewReader(bytes.NewReader(body))
	require.NoError(ar.ReadArtifact())
	assert.Equal(name, ar.GetArtifactName())
	assert.Contains(ar.GetCompatibleDevices(), deviceType)
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

func (s *DeploymentsManagementV1Suite) downloadHTTPClient() *http.Client {
	return s.APIClient.GetConfig().HTTPClient
}

// setupArtifactSelection uploads the given artifacts, then creates an accepted
// device and a deployment targeting it with artifact_name artifactName.
func (s *DeploymentsManagementV1Suite) setupArtifactSelection(
	ctx context.Context, artifactName string, artifacts []*os.File,
) (*common.Device, error) {
	for _, a := range artifacts {
		if _, err := s.uploadArtifact(ctx, a, "description"); err != nil {
			return nil, err
		}
	}

	dev, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	if err != nil {
		return nil, err
	}

	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "deployment-1-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      []string{dev.ID},
		}).Execute()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected create deployment status: %d", r.StatusCode)
	}

	return dev, nil
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

// assertArtifactPayloadSize reads the downloaded artifact and asserts its
// single update payload file is exactly wantSize bytes.
func assertArtifactPayloadSize(t testing.TB, data []byte, wantSize int) {
	t.Helper()

	ar := areader.NewReader(bytes.NewReader(data))
	if err := ar.RegisterHandler(handlers.NewModuleImage("foo")); err != nil {
		t.Fatalf("failed to register handler: %v", err)
	}
	if err := ar.ReadArtifact(); err != nil {
		t.Fatalf("failed to read artifact: %v", err)
	}
	// RegisterHandler's argument is only used as a template: the reader builds
	// its own installer instance (via NewInstance) to hold the parsed data, so
	// the populated files must be read back from GetHandlers(), not from the
	// handler passed to RegisterHandler.
	installer, ok := ar.GetHandlers()[0]
	if !ok {
		t.Fatalf("no installer registered for update 0")
	}
	files := installer.GetUpdateFiles()
	if len(files) != 1 {
		t.Fatalf("expected exactly one update file, got %d", len(files))
	}
	if files[0].Size != int64(wantSize) {
		t.Fatalf("expected payload size %d, got %d", wantSize, files[0].Size)
	}
}
