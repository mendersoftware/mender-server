//nolint:all // This is all test code
package opensource

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	depmodel "github.com/mendersoftware/mender-server/services/deployments/model"
	authmodel "github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DeploymentsManagementV1Suite covers the open-source deployment, artifact and
// create-artifact management tests. The suite runs against a shared,
// never-reset environment together with every other suite, so every test uses
// uuid-suffixed names for the deployments/artifacts/groups it creates and
// scopes its listing assertions by exact name rather than relying on exact,
// unscoped totals.
type DeploymentsManagementV1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
}

func (i *BackendIntegrationSuite) TestDeploymentsManagementV1() {
	suite.Run(i.T(), &DeploymentsManagementV1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (s *DeploymentsManagementV1Suite) SetupTest() {
	require := require.New(s.T())

	ctx := common.BasicAuthContext(s.T().Context(), s.User)
	token, r, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.NotEmpty(token)
	require.Equal(http.StatusOK, r.StatusCode)
	s.JWT = token
}

func (s *DeploymentsManagementV1Suite) TestUploadArtifact() {
	s.Run("Success/MissingTypeInfo", func() {
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
	})
}

func (s *DeploymentsManagementV1Suite) TestRegularDeployment() {
	// A deployment can target an explicit device list or every accepted device
	// (AllDevices). Both drive the same offer/install/already-installed
	// lifecycle, so they share one body parameterised on how the deployment
	// selects its devices.
	s.Run("Success/ExplicitDevices", func() { s.regularDeployment(false) })
	s.Run("Success/AllDevices", func() { s.regularDeployment(true) })
}

func (s *DeploymentsManagementV1Suite) regularDeployment(allDevices bool) {
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := common.DefaultDeviceType
	artifactName := "deployments-artifact-" + uuid.NewString()

	// For an all-devices deployment, stats.pending covers every previously
	// accepted device in the shared, never-reset environment, so record the
	// count right before creating the deployment.
	var acceptedBefore int32
	if allDevices {
		count, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementCountDevices(ctx).Status(authmodel.DevStatusAccepted).Execute()
		require.NoError(s.T(), err)
		acceptedBefore = count.GetCount()
	}

	devs, err := s.makeAcceptedDevices(ctx, 5)
	require.NoError(s.T(), err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(s.T(), err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(s.T(), err)

	newDeployment := client.NewDeployment{
		Name:         "deployment-" + uuid.NewString(),
		ArtifactName: artifactName,
	}
	expected := client.Statistics{Pending: int32(len(devs)), AdditionalProperties: map[string]interface{}{}}
	if allDevices {
		newDeployment.AllDevices = new(true)
		expected = client.Statistics{Pending: acceptedBefore + int32(len(devs)), AdditionalProperties: map[string]interface{}{}}
	} else {
		deviceIDs := make([]string, len(devs))
		for i, d := range devs {
			deviceIDs[i] = d.ID
		}
		newDeployment.Devices = deviceIDs
	}

	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(newDeployment).Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, r.StatusCode)
	depID := path.Base(r.Header.Get("Location"))

	stats, _, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(ctx, depID).Execute()
	require.NoError(s.T(), err)
	require.EqualValues(s.T(), &expected, stats)

	// every device is offered the new artifact ("bugs-bunny" stands in for
	// whatever's currently installed) and reports it successfully installed.
	for _, dev := range devs {
		devCtx := common.JWTAuthContext(ctx, dev.Token)
		instr, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(devCtx).ArtifactName("bugs-bunny").DeviceType(deviceType).Execute()
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusOK, r.StatusCode)
		_, err = s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, instr.GetId()).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusSuccessStr}).
			Execute()
		require.NoError(s.T(), err)
	}

	// the deployment has finished, so nothing more is offered.
	for _, dev := range devs {
		_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, dev.Token)).
			ArtifactName(artifactName).DeviceType(deviceType).Execute()
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusNoContent, r.StatusCode)
	}

	// a second deployment of the same artifact: the devices already have it
	// installed, so nothing is offered.
	secondDeployment := client.NewDeployment{
		Name:         "already-installed-deployment-" + uuid.NewString(),
		ArtifactName: artifactName,
	}
	if allDevices {
		secondDeployment.AllDevices = new(true)
	} else {
		secondDeployment.Devices = newDeployment.Devices
	}
	r, err = s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(secondDeployment).Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, r.StatusCode)

	for _, dev := range devs {
		_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, dev.Token)).
			ArtifactName(artifactName).DeviceType(deviceType).Execute()
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusNoContent, r.StatusCode)
	}
}

func (s *DeploymentsManagementV1Suite) TestListingDeployments() {
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := common.DefaultDeviceType
	artifactName := "deployments-artifact-" + uuid.NewString()
	// CreateDeployment doesn't validate that the target devices exist, so
	// throwaway ids avoid creating real devices just to list deployments.
	namePrefix := "listdeployment-" + uuid.NewString() + "-"
	deviceIDs := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(s.T(), err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(s.T(), err)

	names := []string{namePrefix + "1", namePrefix + "2", namePrefix + "3", namePrefix + "4", namePrefix + "5"}
	idByName := map[string]string{}
	for _, name := range names {
		r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
			NewDeployment(client.NewDeployment{
				Name:         name,
				ArtifactName: artifactName,
				Devices:      deviceIDs,
			}).Execute()
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusCreated, r.StatusCode)
		depID := path.Base(r.Header.Get("Location"))
		require.NotEmpty(s.T(), depID)
		idByName[name] = depID
	}

	apiV2 := s.APIClient.DeploymentsV2ManagementAPIAPI

	s.Run("Success/FilterByExactName", func() {
		require := require.New(s.T())
		// each deployment is retrievable by its exact name, scoping the listing
		// to exactly this run's deployments.
		for _, name := range names {
			deps, _, err := apiV2.DeploymentsV2ListDeployments(ctx).Name(name).Execute()
			require.NoError(err)
			require.Len(deps, 1)
			require.Equal(idByName[name], deps[0].Id)
		}
	})

	s.Run("Success/FilterByCreatedTimestamp", func() {
		require := require.New(s.T())
		// created_before/created_after are inclusive, second-grained filters.
		// Scope by exact name (not by the timestamp) and derive the bound from
		// the server-assigned created timestamp, so this is independent of other
		// suites' deployments and of host-vs-server clock skew.
		name := names[0]
		deps, _, err := apiV2.DeploymentsV2ListDeployments(ctx).Name(name).Execute()
		require.NoError(err)
		require.Len(deps, 1)
		created := int32(deps[0].Created.Unix())

		// a window bracketing the created second (±1s) returns it
		deps, _, err = apiV2.DeploymentsV2ListDeployments(ctx).Name(name).CreatedAfter(created - 1).Execute()
		require.NoError(err)
		require.Len(deps, 1)
		deps, _, err = apiV2.DeploymentsV2ListDeployments(ctx).Name(name).CreatedBefore(created + 1).Execute()
		require.NoError(err)
		require.Len(deps, 1)

		// a window entirely after / before the created second excludes it
		deps, _, err = apiV2.DeploymentsV2ListDeployments(ctx).Name(name).CreatedAfter(created + 1).Execute()
		require.NoError(err)
		require.Len(deps, 0)
		deps, _, err = apiV2.DeploymentsV2ListDeployments(ctx).Name(name).CreatedBefore(created - 1).Execute()
		require.NoError(err)
		require.Len(deps, 0)
	})

	s.Run("Success/Pagination", func() {
		require := require.New(s.T())
		// paging params, scoped to a single named deployment
		name := names[0]
		deps, _, err := apiV2.DeploymentsV2ListDeployments(ctx).Name(name).PerPage(1).Page(1).Execute()
		require.NoError(err)
		require.Len(deps, 1)
		// page beyond the last returns zero results
		deps, _, err = apiV2.DeploymentsV2ListDeployments(ctx).Name(name).PerPage(1).Page(2).Execute()
		require.NoError(err)
		require.Len(deps, 0)
	})

	s.Run("Failure/PerPageOverMax", func() {
		require := require.New(s.T())
		// The generated client types per_page/page as int32, so use values that
		// are valid int32s but still rejected by the API (per_page over its
		// configured max).
		_, r, err := apiV2.DeploymentsV2ListDeployments(ctx).PerPage(501).Execute()
		require.Error(err)
		require.Equal(http.StatusBadRequest, r.StatusCode)
	})

	s.Run("Failure/PageBelowOne", func() {
		require := require.New(s.T())
		_, r, err := apiV2.DeploymentsV2ListDeployments(ctx).Page(0).Execute()
		require.Error(err)
		require.Equal(http.StatusBadRequest, r.StatusCode)
	})
}

func (s *DeploymentsManagementV1Suite) TestDeploymentStatusUpdate() {
	// The same status lifecycle runs against an explicit device list and
	// against a device group; only how the deployment selects its devices
	// differs.
	s.Run("Success/ExplicitDevices", func() { s.runDeploymentStatusUpdate("") })
	s.Run("Success/Group", func() { s.runDeploymentStatusUpdate("g0-" + uuid.NewString()) })
}

// runDeploymentStatusUpdate drives a full deployment status lifecycle for one
// device, optionally targeting a device group instead of an explicit device
// list.
func (s *DeploymentsManagementV1Suite) runDeploymentStatusUpdate(deployToGroup string) {
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := common.DefaultDeviceType
	artifactName := "deployments-artifact-" + uuid.NewString()

	devs, err := s.makeAcceptedDevices(ctx, 5)
	require.NoError(s.T(), err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(s.T(), err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(s.T(), err)

	if deployToGroup != "" {
		invm := s.APIClient.DeviceInventoryManagementAPIAPI
		for _, dev := range devs[:len(devs)-1] {
			_, err := invm.AssignGroup(ctx, dev.ID).
				Group(client.Group{Group: deployToGroup}).Execute()
			require.NoError(s.T(), err)
		}
	}

	var r *http.Response
	depName := "deployment-" + uuid.NewString()
	if deployToGroup != "" {
		r, err = s.APIClient.DeploymentsManagementAPIAPI.
			CreateDeploymentForAGroupOfDevices(ctx, deployToGroup).
			NewDeploymentForGroup(client.NewDeploymentForGroup{
				Name:         depName,
				ArtifactName: artifactName,
			}).Execute()
	} else {
		deviceIDs := make([]string, len(devs)-1)
		for i, d := range devs[:len(devs)-1] {
			deviceIDs[i] = d.ID
		}
		r, err = s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
			NewDeployment(client.NewDeployment{
				Name:         depName,
				ArtifactName: artifactName,
				Devices:      deviceIDs,
			}).Execute()
	}
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, r.StatusCode)
	depID := path.Base(r.Header.Get("Location"))

	inProgress := string(depmodel.DeploymentStatusInProgress)

	dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), string(depmodel.DeploymentStatusPending), dep.Status)

	s.Run("Success/NextUpdateAvailable", func() {
		require := require.New(s.T())
		// devs[0]: next update available
		_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, devs[0].Token)).
			ArtifactName("bugs-bunny").DeviceType(deviceType).Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
	})

	s.Run("Success/AlreadyInstalled", func() {
		require := require.New(s.T())
		// devs[1]: already has the artifact installed
		_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, devs[1].Token)).
			ArtifactName(artifactName).DeviceType(deviceType).Execute()
		require.NoError(err)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[1].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusAlreadyInstStr, string(dd.Status))

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(inProgress, dep.Status)
	})

	s.Run("Success/IncompatibleDeviceType", func() {
		require := require.New(s.T())
		// devs[2]: incompatible device type
		_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, devs[2].Token)).
			ArtifactName("bugs-bunny").DeviceType("foo").Execute()
		require.NoError(err)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[2].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusNoArtifactStr, string(dd.Status))

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(inProgress, dep.Status)
	})

	s.Run("Failure/NotPartOfDeployment", func() {
		require := require.New(s.T())
		// devs[4] is not part of the deployment, so the status update is rejected
		// and the deployment stays in progress.
		devCtx := common.JWTAuthContext(ctx, devs[4].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusInstallingStr}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusNotFound, r.StatusCode)

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(inProgress, dep.Status)
	})

	s.Run("Failure/WrongStatus", func() {
		require := require.New(s.T())
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: "foo"}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusBadRequest, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusPendingStr, string(dd.Status))

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(inProgress, dep.Status)
	})

	s.Run("Success/PendingToDownloading", func() {
		require := require.New(s.T())
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusDownloadingStr}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusDownloadingStr, string(dd.Status))

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(inProgress, dep.Status)
	})

	s.Run("Success/DownloadingToPauseBeforeInstalling", func() {
		require := require.New(s.T())
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusPauseBeforeInstallStr}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusPauseBeforeInstallStr, string(dd.Status))

		stats, _, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(ctx, depID).Execute()
		require.NoError(err)
		require.EqualValues(&client.Statistics{
			AlreadyInstalled:      1,
			Noartifact:            1,
			Pending:               1,
			PauseBeforeInstalling: 1,
			AdditionalProperties:  map[string]interface{}{},
		}, stats)
	})

	s.Run("Success/PauseToInstallingWithSubstate", func() {
		require := require.New(s.T())
		// pause_before_installing -> installing, substate "" -> "foo"
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{
				Status:   depmodel.DeviceDeploymentStatusInstallingStr,
				Substate: new("foo"),
			}).Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusInstallingStr, string(dd.Status))
		require.Equal("foo", dd.GetSubstate())
	})

	s.Run("Success/InstallingToDownloading", func() {
		require := require.New(s.T())
		// installing -> downloading (any valid status transition is allowed until finished)
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusDownloadingStr}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusDownloadingStr, string(dd.Status))
		// substate is preserved across the transition
		require.Equal("foo", dd.GetSubstate())
	})

	s.Run("Success/DownloadingToPauseBeforeCommitting", func() {
		require := require.New(s.T())
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{
				Status:   depmodel.DeviceDeploymentStatusPauseBeforeCommitStr,
				Substate: new("foo"),
			}).Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		stats, _, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(ctx, depID).Execute()
		require.NoError(err)
		require.EqualValues(&client.Statistics{
			AlreadyInstalled:      1,
			Noartifact:            1,
			Pending:               1,
			PauseBeforeCommitting: 1,
			AdditionalProperties:  map[string]interface{}{},
		}, stats)
	})

	s.Run("Success/PauseCommittingToPauseRebooting", func() {
		require := require.New(s.T())
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{
				Status:   depmodel.DeviceDeploymentStatusPauseBeforeRebootStr,
				Substate: new("foo"),
			}).Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		stats, _, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(ctx, depID).Execute()
		require.NoError(err)
		require.EqualValues(&client.Statistics{
			AlreadyInstalled:     1,
			Noartifact:           1,
			Pending:              1,
			PauseBeforeRebooting: 1,
			AdditionalProperties: map[string]interface{}{},
		}, stats)
	})

	s.Run("Success/PauseRebootingToRebooting", func() {
		require := require.New(s.T())
		// pause_before_rebooting -> rebooting, substate "foo" -> "bar"
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{
				Status:   depmodel.DeviceDeploymentStatusRebootingStr,
				Substate: new("bar"),
			}).Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusRebootingStr, string(dd.Status))
		require.Equal("bar", dd.GetSubstate())
	})

	s.Run("Success/MidCycleRepoll", func() {
		require := require.New(s.T())
		// CheckUpdate again with the original artifact/device-type params still
		// returns the same deployment while it's inprogress.
		instrAgain, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, devs[0].Token)).
			ArtifactName("bugs-bunny").DeviceType(deviceType).Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
		require.Equal(depID, instrAgain.GetId())
	})

	s.Run("Success/RebootingToSuccess", func() {
		require := require.New(s.T())
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusSuccessStr}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusNoContent, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusSuccessStr, string(dd.Status))
		require.Equal("bar", dd.GetSubstate())
	})

	s.Run("Failure/UpdateAfterFinished", func() {
		require := require.New(s.T())
		// devs[0] already reported success, so a further status update is rejected
		devCtx := common.JWTAuthContext(ctx, devs[0].Token)
		r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusPendingStr}).
			Execute()
		require.NotNil(r)
		require.Equal(http.StatusBadRequest, r.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[0].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusSuccessStr, string(dd.Status))
	})

	s.Run("Success/LastDeviceFinishesDeployment", func() {
		require := require.New(s.T())
		// devs[3]: next update available
		_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			CheckUpdate(common.JWTAuthContext(ctx, devs[3].Token)).
			ArtifactName("bugs-bunny").DeviceType(deviceType).Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)

		// devs[3]: pending -> failure, which finishes the deployment
		devCtx := common.JWTAuthContext(ctx, devs[3].Token)
		r2, _ := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(devCtx, depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusFailureStr}).
			Execute()
		require.NotNil(r2)
		require.Equal(http.StatusNoContent, r2.StatusCode)

		devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.ListAllDevicesInDeployment(ctx, depID).Execute()
		require.NoError(err)
		var dd *client.DeviceWithImage
		for i := range devices {
			if devices[i].Id == devs[3].ID {
				dd = &devices[i]
				break
			}
		}
		require.NotNil(dd)
		require.Equal(depmodel.DeviceDeploymentStatusFailureStr, string(dd.Status))

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(string(depmodel.DeploymentStatusFinished), dep.Status)
	})

	s.Run("Success/PostFinishStatusOverride", func() {
		require := require.New(s.T())
		// reporting a new terminal status for a device whose deployment has
		// already finished is still accepted; the deployment status stays
		// "finished" but the stats reflect the override.
		statsBefore, _, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(ctx, depID).Execute()
		require.NoError(err)

		r, err := s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(common.JWTAuthContext(ctx, devs[0].Token), depID).
			DeploymentStatus(client.DeploymentStatus{Status: depmodel.DeviceDeploymentStatusFailureStr}).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusNoContent, r.StatusCode)

		statsAfter, _, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentStatusStatistics(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(statsBefore.Success-1, statsAfter.Success)
		require.Equal(statsBefore.Failure+1, statsAfter.Failure)

		dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, depID).Execute()
		require.NoError(err)
		require.Equal(string(depmodel.DeploymentStatusFinished), dep.Status)
	})
}

// makeAcceptedDevices creates n onboarded (accepted) devices.
func (s *DeploymentsManagementV1Suite) makeAcceptedDevices(ctx context.Context, n int) ([]*common.Device, error) {
	devs := make([]*common.Device, 0, n)
	for range n {
		dev, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
		if err != nil {
			return nil, err
		}
		devs = append(devs, dev)
	}
	return devs, nil
}

// uploadArtifact uploads the given artifact file under the management API
// and returns its id (extracted from the Location header).
func (s *DeploymentsManagementV1Suite) uploadArtifact(
	ctx context.Context, artifact *os.File, description string,
) (string, error) {
	defer artifact.Close()

	fi, err := artifact.Stat()
	if err != nil {
		return "", err
	}
	r, err := s.APIClient.DeploymentsManagementAPIAPI.UploadArtifact(ctx).
		Artifact(artifact).
		Size(int32(fi.Size())).
		Description(description).
		Execute()
	if err != nil {
		return "", err
	}
	if r.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected upload artifact status: %d", r.StatusCode)
	}
	id := path.Base(r.Header.Get("Location"))
	return id, nil
}
