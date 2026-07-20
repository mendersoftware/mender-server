//nolint:all // This is all test code
package opensource

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DeploymentsManagementV1Suite ports (a subset of) the deployments tests from
// mender-server/backend/tests/integration/test_deployments.py,
// test_artifact.py and test_create_artifact.py. Enterprise-only branches
// (useExistingTenant()/tenant-scoped fixtures) are intentionally not ported.
//
// The suite runs against a shared, never-reset environment together with
// every other suite, so every test below uses uuid-suffixed names for
// deployments/artifacts/groups it creates and scopes its listing assertions
// (by exact name, or by a created_after/created_before time window) instead
// of relying on exact, unscoped totals the way the original pytest suite
// could (each pytest module got a freshly seeded database).
type DeploymentsManagementV1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string

	// createdDevices tracks every device this suite creates, so
	// TearDownSuite can decommission them again -- see the same field in
	// DevauthManagementV2Suite for why this matters on a shared environment.
	createdDevices []string
	// createdArtifacts/createdDeployments are best-effort cleanup lists;
	// deployments/artifacts don't affect other suites' exact-count
	// assertions the way devices do, so failures to delete them here are
	// not fatal.
	createdArtifacts   []string
	createdDeployments []string
}

func (i *BackendIntegrationSuite) TestDeploymentsManagementV1() {
	suite.Run(i.T(), &DeploymentsManagementV1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (s *DeploymentsManagementV1Suite) SetupSuite() {
	require := require.New(s.T())

	ctx := common.BasicAuthContext(s.T().Context(), s.User)
	token, r, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.Equal(http.StatusOK, r.StatusCode)
	require.NotEmpty(token)
	s.JWT = token
}

// TearDownSuite decommissions every device this suite created (so it
// doesn't leak into other suites' exact device/inventory count assertions)
// and best-effort cleans up deployments/artifacts.
func (s *DeploymentsManagementV1Suite) TearDownSuite() {
	if s.JWT == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx = common.JWTAuthContext(ctx, s.JWT)
	for _, id := range s.createdDeployments {
		_, _ = s.APIClient.DeploymentsManagementAPIAPI.AbortDeployment(ctx, id).
			AbortDeploymentRequest(client.AbortDeploymentRequest{Status: "aborted"}).
			Execute()
	}
	for _, id := range s.createdArtifacts {
		_, _ = s.APIClient.DeploymentsManagementAPIAPI.DeleteArtifact(ctx, id).Execute()
	}
	for _, id := range s.createdDevices {
		_, _ = s.deleteDevice(ctx, id)
	}
}

func (s *DeploymentsManagementV1Suite) trackDevice(id string) {
	if id != "" {
		s.createdDevices = append(s.createdDevices, id)
	}
}

func (s *DeploymentsManagementV1Suite) trackArtifact(id string) {
	if id != "" {
		s.createdArtifacts = append(s.createdArtifacts, id)
	}
}

func (s *DeploymentsManagementV1Suite) trackDeployment(id string) {
	if id != "" {
		s.createdDeployments = append(s.createdDeployments, id)
	}
}

func (s *DeploymentsManagementV1Suite) deleteDevice(ctx context.Context, id string) (int, error) {
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementDecommissionDevice(ctx, id).Execute()
	if r == nil {
		return 0, err
	}
	return r.StatusCode, err
}

// ---------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------

// makeAcceptedDevices creates n onboarded (accepted) devices, tracking them
// for cleanup.
func (s *DeploymentsManagementV1Suite) makeAcceptedDevices(ctx context.Context, n int) ([]*common.Device, error) {
	devs := make([]*common.Device, 0, n)
	for range n {
		dev, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
		if err != nil {
			return nil, err
		}
		s.trackDevice(dev.ID)
		devs = append(devs, dev)
	}
	return devs, nil
}

// uploadArtifact uploads the given artifact file under the management API
// and returns its id (extracted from the Location header), tracking it for
// (best-effort) cleanup.
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
	s.trackArtifact(id)
	return id, nil
}

// tryUpdate mimics test_deployments.py::try_update: it asks for the next
// update for the device and, if one is offered, immediately reports it as
// successfully installed. It returns the CheckUpdate status code.
func (s *DeploymentsManagementV1Suite) tryUpdate(
	ctx context.Context, dev *common.Device, artifactName, deviceType string,
) (int, error) {
	instr, r, err := dev.CheckUpdate(ctx, s.APIClient, artifactName, deviceType)
	if r == nil {
		return 0, err
	}
	if r.StatusCode == http.StatusOK {
		_, err = s.APIClient.DeploymentsDeviceAPIAPI.
			UpdateDeploymentStatus(common.JWTAuthContext(ctx, dev.Token), instr.GetId()).
			DeploymentStatus(client.DeploymentStatus{Status: "success"}).
			Execute()
		if err != nil {
			return r.StatusCode, err
		}
	}
	return r.StatusCode, nil
}

func (s *DeploymentsManagementV1Suite) getStats(ctx context.Context, depID string) (*client.Statistics, error) {
	stats, _, err := s.APIClient.DeploymentsManagementAPIAPI.
		DeploymentStatusStatistics(ctx, depID).Execute()
	return stats, err
}

// verifyStats asserts that every field of stats not mentioned in expected is
// zero, and that the mentioned ones match (mirrors verify_stats in
// test_deployments.py).
func (s *DeploymentsManagementV1Suite) verifyStats(stats *client.Statistics, expected map[string]int32) {
	assert := assert.New(s.T())
	fields := map[string]int32{
		"success":                 stats.Success,
		"pending":                 stats.Pending,
		"downloading":             stats.Downloading,
		"rebooting":               stats.Rebooting,
		"installing":              stats.Installing,
		"failure":                 stats.Failure,
		"noartifact":              stats.Noartifact,
		"already-installed":       stats.AlreadyInstalled,
		"aborted":                 stats.Aborted,
		"pause_before_installing": stats.PauseBeforeInstalling,
		"pause_before_rebooting":  stats.PauseBeforeRebooting,
		"pause_before_committing": stats.PauseBeforeCommitting,
	}
	for k, v := range fields {
		if exp, ok := expected[k]; ok {
			assert.Equalf(exp, v, "stat %q", k)
		} else {
			assert.Equalf(int32(0), v, "stat %q", k)
		}
	}
}

func (s *DeploymentsManagementV1Suite) getDeployment(ctx context.Context, id string) (*client.DeploymentV1, error) {
	dep, _, err := s.APIClient.DeploymentsManagementAPIAPI.ShowDeployment(ctx, id).Execute()
	return dep, err
}

// deviceDeploymentStatus looks up a single device's status/substate within
// a deployment, as reported by GET /deployments/{id}/devices. A device
// that is not part of the deployment yields (nil, nil): the Python
// original's status_verify iterates the device list and simply asserts
// nothing when the device is absent, and the not-part-of-deployment
// status-update case relies on that.
func (s *DeploymentsManagementV1Suite) deviceDeploymentStatus(
	ctx context.Context, deploymentID, deviceID string,
) (*client.DeviceWithImage, error) {
	devices, _, err := s.APIClient.DeploymentsManagementAPIAPI.
		ListAllDevicesInDeployment(ctx, deploymentID).Execute()
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		if d.Id == deviceID {
			return &d, nil
		}
	}
	return nil, nil
}

// statusUpdateAndVerify mirrors StatusVerifier.status_update_and_verify: it
// updates a device's deployment status (optionally with a substate) and then
// checks the resulting device-deployment and deployment status.
func (s *DeploymentsManagementV1Suite) statusUpdateAndVerify(
	ctx context.Context,
	dev *common.Device,
	deploymentID string,
	status, substate string,
	expectStatusCode int,
	expectDeviceStatus, expectSubstate, expectDeploymentStatus string,
) {
	require := require.New(s.T())
	devCtx := common.JWTAuthContext(ctx, dev.Token)

	body := client.DeploymentStatus{Status: status}
	if substate != "" {
		body.Substate = client.PtrString(substate)
	}
	r, _ := s.APIClient.DeploymentsDeviceAPIAPI.
		UpdateDeploymentStatus(devCtx, deploymentID).
		DeploymentStatus(body).
		Execute()
	require.NotNil(r)
	require.Equal(expectStatusCode, r.StatusCode)

	s.statusVerify(ctx, deploymentID, dev.ID, expectDeviceStatus, expectSubstate, expectDeploymentStatus)
}

func (s *DeploymentsManagementV1Suite) statusVerify(
	ctx context.Context,
	deploymentID, deviceID, expectDeviceStatus, expectSubstate, expectDeploymentStatus string,
) {
	require := require.New(s.T())
	assert := assert.New(s.T())

	if expectDeviceStatus != "" {
		dd, err := s.deviceDeploymentStatus(ctx, deploymentID, deviceID)
		require.NoError(err)
		if dd != nil {
			assert.Equal(expectDeviceStatus, string(dd.Status))
			if expectSubstate != "" {
				require.NotNil(dd.Substate)
				assert.Equal(expectSubstate, *dd.Substate)
			}
		}
	}

	if expectDeploymentStatus != "" {
		dep, err := s.getDeployment(ctx, deploymentID)
		require.NoError(err)
		assert.Equal(expectDeploymentStatus, dep.Status)
	}
}

// ---------------------------------------------------------------------
// _TestDeploymentsBase (test_deployments.py::TestDeploymentOpenSource)
// ---------------------------------------------------------------------

func (s *DeploymentsManagementV1Suite) TestRegularDeployment() {
	// ported from test_deployments.py::_TestDeploymentsBase::do_test_regular_deployment
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := "qemux86-64"
	artifactName := "deployments-phase-testing-" + uuid.NewString()

	devs, err := s.makeAcceptedDevices(ctx, 5)
	require.NoError(err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

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
	s.trackDeployment(depID)

	stats, err := s.getStats(ctx, depID)
	require.NoError(err)
	s.verifyStats(stats, map[string]int32{"pending": int32(len(devs))})

	for _, dev := range devs {
		// devices haven't installed anything yet ("bugs-bunny" stands in
		// for whatever's currently on the device), so the new artifact
		// should be offered.
		code, err := s.tryUpdate(ctx, dev, "bugs-bunny", deviceType)
		require.NoError(err)
		assert.Equal(http.StatusOK, code)
	}

	for _, dev := range devs {
		// deployment already finished
		code, err := s.tryUpdate(ctx, dev, artifactName, deviceType)
		require.NoError(err)
		assert.Equal(http.StatusNoContent, code)
	}

	r, err = s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "really-old-update-" + uuid.NewString(),
			ArtifactName: artifactName,
			Devices:      deviceIDs,
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)
	s.trackDeployment(path.Base(r.Header.Get("Location")))

	for _, dev := range devs {
		// already installed
		code, err := s.tryUpdate(ctx, dev, artifactName, deviceType)
		require.NoError(err)
		assert.Equal(http.StatusNoContent, code)
	}
}

func (s *DeploymentsManagementV1Suite) TestRegularDeploymentAllDevices() {
	// ported from test_deployments.py::_TestDeploymentsBase::do_test_regular_deployment_all_devices
	//
	// Deviation: the original test asserts stats.pending == len(devs)
	// after creating an all_devices deployment, which only holds on a
	// freshly seeded database. On this shared environment, an
	// all_devices deployment targets every previously accepted device
	// from every other suite too, so we assert against the accepted
	// device count observed right before creating the deployment plus
	// the devices we just created, instead of len(devs) alone.
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := "qemux86-64"
	artifactName := "deployments-phase-testing-" + uuid.NewString()

	acceptedBefore, err := s.countAcceptedDevices(ctx)
	require.NoError(err)

	devs, err := s.makeAcceptedDevices(ctx, 5)
	require.NoError(err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

	r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "phased-deployment-" + uuid.NewString(),
			ArtifactName: artifactName,
			AllDevices:   client.PtrBool(true),
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)
	depID := path.Base(r.Header.Get("Location"))
	s.trackDeployment(depID)

	stats, err := s.getStats(ctx, depID)
	require.NoError(err)
	s.verifyStats(stats, map[string]int32{"pending": acceptedBefore + int32(len(devs))})

	for _, dev := range devs {
		// devices haven't installed anything yet ("bugs-bunny" stands in
		// for whatever's currently on the device), so the new artifact
		// should be offered.
		code, err := s.tryUpdate(ctx, dev, "bugs-bunny", deviceType)
		require.NoError(err)
		assert.Equal(http.StatusOK, code)
	}

	for _, dev := range devs {
		code, err := s.tryUpdate(ctx, dev, artifactName, deviceType)
		require.NoError(err)
		assert.Equal(http.StatusNoContent, code)
	}

	r, err = s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
		NewDeployment(client.NewDeployment{
			Name:         "really-old-update-" + uuid.NewString(),
			ArtifactName: artifactName,
			AllDevices:   client.PtrBool(true),
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)
	s.trackDeployment(path.Base(r.Header.Get("Location")))

	for _, dev := range devs {
		code, err := s.tryUpdate(ctx, dev, artifactName, deviceType)
		require.NoError(err)
		assert.Equal(http.StatusNoContent, code)
	}
}

func (s *DeploymentsManagementV1Suite) countAcceptedDevices(ctx context.Context) (int32, error) {
	count, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementCountDevices(ctx).Status("accepted").Execute()
	if err != nil {
		return 0, err
	}
	return count.GetCount(), nil
}

func (s *DeploymentsManagementV1Suite) TestListingDeployments() {
	// ported from test_deployments.py::_TestDeploymentsBase::do_test_listing_deployments
	//
	// Deviation: the deployments v2 list endpoint's generated Name(...)
	// filter only accepts a single name (a limitation of the client
	// codegen -- the API itself accepts the query parameter multiple
	// times), so instead of querying by the 5 deployment names at once,
	// this scopes "list all of them" style assertions to a
	// created_after/created_before window bracketing this test, the same
	// mechanism the original test used for its explicit
	// created_after/created_before assertions.
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := "qemux86-64"
	artifactName := "deployments-phase-testing-" + uuid.NewString()
	namePrefix := "phaseddeployment-" + uuid.NewString() + "-"

	devs, err := s.makeAcceptedDevices(ctx, 5)
	require.NoError(err)
	deviceIDs := make([]string, len(devs))
	for i, d := range devs {
		deviceIDs[i] = d.ID
	}

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

	// created_after/created_before appear to be inclusive, second-grained
	// bounds (see the assertions below). Since another test may have
	// created a deployment just moments ago (same wall-clock second),
	// wait for a fresh second boundary before recording "before", so it
	// doesn't sweep in deployments created immediately before this test
	// started.
	time.Sleep(1100 * time.Millisecond)
	before := int32(time.Now().Unix())

	createdAtByName := map[string]int32{}
	idsByName := map[string]string{}
	// The python original runs inside the compose network and shares the
	// server's clock; this test runs on the host, where the docker VM's
	// clock can drift (macOS). Measure the skew from the first
	// deployment's server-assigned created timestamp and shift all our
	// window bounds by it, so created_after/created_before comparisons
	// happen on the server's clock.
	var clockSkew int32
	for _, i := range []string{"1", "2", "3", "4", "5"} {
		name := namePrefix + i
		r, err := s.APIClient.DeploymentsManagementAPIAPI.DeploymentsCreateDeployment(ctx).
			NewDeployment(client.NewDeployment{
				Name:         name,
				ArtifactName: artifactName,
				Devices:      deviceIDs,
			}).Execute()
		require.NoError(err)
		require.Equal(http.StatusCreated, r.StatusCode)
		createdAtByName[i] = int32(time.Now().Unix())
		depID := path.Base(r.Header.Get("Location"))
		require.NotEmpty(depID)
		s.trackDeployment(depID)
		idsByName[name] = depID
		if i == "1" {
			dep, err := s.getDeployment(ctx, depID)
			require.NoError(err)
			clockSkew = int32(dep.Created.Unix()) - createdAtByName[i]
		}
		time.Sleep(4 * time.Second)
	}
	after := int32(time.Now().Unix()+1) + clockSkew
	before += clockSkew
	for k := range createdAtByName {
		createdAtByName[k] += clockSkew
	}

	apiV2 := s.APIClient.DeploymentsV2ManagementAPIAPI

	list := func(req client.ApiDeploymentsV2ListDeploymentsRequest) []client.DeploymentV2 {
		deps, _, err := req.Execute()
		require.NoError(err)
		return deps
	}
	scoped := func() client.ApiDeploymentsV2ListDeploymentsRequest {
		return apiV2.DeploymentsV2ListDeployments(ctx).
			CreatedAfter(before).CreatedBefore(after)
	}

	deps := list(scoped())
	assert.Len(deps, 5)

	// per_page/page pagination, scoped to our window
	for _, page := range []int32{1, 2, 3, 4, 5} {
		deps = list(scoped().PerPage(1).Page(page))
		assert.Len(deps, 1)
	}

	// created_after query parameter
	deps = list(apiV2.DeploymentsV2ListDeployments(ctx).
		CreatedAfter(createdAtByName["2"]).CreatedBefore(after))
	assert.Len(deps, 4)
	for _, j := range []string{"2", "3", "4", "5"} {
		expectedID := idsByName[namePrefix+j]
		found := false
		for _, d := range deps {
			if d.Id == expectedID {
				found = true
				break
			}
		}
		assert.True(found, "deployment %s not found", namePrefix+j)
	}

	// created_before query parameter
	deps = list(apiV2.DeploymentsV2ListDeployments(ctx).
		CreatedAfter(before).CreatedBefore(createdAtByName["4"]))
	assert.Len(deps, 3)

	// per_page query parameter
	deps = list(scoped().PerPage(1))
	assert.Len(deps, 1)

	// per_page and page query parameters
	deps = list(scoped().Page(2).PerPage(1))
	assert.Len(deps, 1)

	// per_page query parameter covering everything
	deps = list(scoped().PerPage(500))
	assert.Len(deps, 5)

	// pagination overshoot: page beyond last should return zero results
	deps = list(scoped().PerPage(10).Page(2))
	assert.Len(deps, 0)

	// invalid per_page/page: 400.
	//
	// Deviation: the original python test sends an out-of-int64-range
	// numeric string (18446744073709551617) to trigger a parse failure.
	// The generated client types per_page/page as int32, so that literal
	// can't even be represented; use values that are valid int32s but
	// still rejected by the API (per_page over its configured max, page
	// below 1) to exercise the same 400 path.
	_, r, err := apiV2.DeploymentsV2ListDeployments(ctx).PerPage(501).Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)

	_, r, err = apiV2.DeploymentsV2ListDeployments(ctx).Page(0).Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)
}

// ---------------------------------------------------------------------
// _TestDeploymentsStatusUpdateBase (test_deployments.py::TestDeploymentsStatusUpdate,
// TestDeploymentsToGroupStatusUpdate)
// ---------------------------------------------------------------------

func (s *DeploymentsManagementV1Suite) TestDeploymentStatusUpdate() {
	// ported from test_deployments.py::_TestDeploymentsStatusUpdateBase::do_test_deployment_status_update
	s.runDeploymentStatusUpdate("")
}

func (s *DeploymentsManagementV1Suite) TestDeploymentStatusUpdateToGroup() {
	// ported from test_deployments.py::_TestDeploymentsStatusUpdateBase::do_test_deployment_status_update
	// (TestDeploymentsToGroupStatusUpdate variant, deploy_to_group="g0")
	s.runDeploymentStatusUpdate("g0-" + uuid.NewString())
}

func (s *DeploymentsManagementV1Suite) runDeploymentStatusUpdate(deployToGroup string) {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	deviceType := "qemux86-64"
	artifactName := "deployments-phase-testing-" + uuid.NewString()

	devs, err := s.makeAcceptedDevices(ctx, 5)
	require.NoError(err)

	artifact, err := common.CreateArtifact(artifactName, s.T(), common.WithCompatibleDevices([]string{deviceType}))
	require.NoError(err)
	_, err = s.uploadArtifact(ctx, artifact, "abc")
	require.NoError(err)

	if deployToGroup != "" {
		invm := s.APIClient.DeviceInventoryManagementAPIAPI
		for _, dev := range devs[:len(devs)-1] {
			_, err := invm.AssignGroup(ctx, dev.ID).
				Group(client.Group{Group: deployToGroup}).Execute()
			require.NoError(err)
		}
	}

	var r *http.Response
	depName := "phased-deployment-" + uuid.NewString()
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
	require.NoError(err)
	require.Equal(http.StatusCreated, r.StatusCode)
	depID := path.Base(r.Header.Get("Location"))
	s.trackDeployment(depID)

	s.statusVerify(ctx, depID, "", "", "", "pending")

	// devs[0]: next update available
	_, r2, err := devs[0].CheckUpdate(ctx, s.APIClient, "bugs-bunny", deviceType)
	require.NoError(err)
	require.Equal(http.StatusOK, r2.StatusCode)

	// devs[1]: already has the artifact installed
	_, r2, err = devs[1].CheckUpdate(ctx, s.APIClient, artifactName, deviceType)
	require.NoError(err)
	require.Equal(http.StatusNoContent, r2.StatusCode)
	s.statusVerify(ctx, depID, devs[1].ID, "already-installed", "", "inprogress")

	// devs[2]: incompatible device type
	_, r2, err = devs[2].CheckUpdate(ctx, s.APIClient, "bugs-bunny", "foo")
	require.NoError(err)
	require.Equal(http.StatusNoContent, r2.StatusCode)
	s.statusVerify(ctx, depID, devs[2].ID, "noartifact", "", "inprogress")

	// devs[4] is not part of the deployment
	s.statusUpdateAndVerify(ctx, devs[4], depID, "installing", "", http.StatusNotFound,
		"does-not-matter", "", "inprogress")

	// wrong status
	s.statusUpdateAndVerify(ctx, devs[0], depID, "foo", "", http.StatusBadRequest,
		"pending", "", "inprogress")

	// pending -> downloading
	s.statusUpdateAndVerify(ctx, devs[0], depID, "downloading", "", http.StatusNoContent,
		"downloading", "", "inprogress")

	// G5: downloading -> pause_before_installing (python test_device_deployments_full
	// L476-487): device status and deployment stats reflect the pause substate.
	s.statusUpdateAndVerify(ctx, devs[0], depID, "pause_before_installing", "", http.StatusNoContent,
		"pause_before_installing", "", "inprogress")
	stats, err := s.getStats(ctx, depID)
	require.NoError(err)
	s.verifyStats(stats, map[string]int32{"already-installed": 1, "noartifact": 1, "pending": 1, "pause_before_installing": 1})

	// pause_before_installing -> installing, substate "" -> "foo"
	s.statusUpdateAndVerify(ctx, devs[0], depID, "installing", "foo", http.StatusNoContent,
		"installing", "foo", "inprogress")

	// installing -> downloading (any valid status transition is allowed until finished)
	s.statusUpdateAndVerify(ctx, devs[0], depID, "downloading", "", http.StatusNoContent,
		"downloading", "foo", "inprogress")

	// G5: downloading -> pause_before_committing (python L476-487).
	s.statusUpdateAndVerify(ctx, devs[0], depID, "pause_before_committing", "foo", http.StatusNoContent,
		"pause_before_committing", "foo", "inprogress")
	stats, err = s.getStats(ctx, depID)
	require.NoError(err)
	s.verifyStats(stats, map[string]int32{"already-installed": 1, "noartifact": 1, "pending": 1, "pause_before_committing": 1})

	// G5: pause_before_committing -> pause_before_rebooting (python L476-487).
	s.statusUpdateAndVerify(ctx, devs[0], depID, "pause_before_rebooting", "foo", http.StatusNoContent,
		"pause_before_rebooting", "foo", "inprogress")
	stats, err = s.getStats(ctx, depID)
	require.NoError(err)
	s.verifyStats(stats, map[string]int32{"already-installed": 1, "noartifact": 1, "pending": 1, "pause_before_rebooting": 1})

	// pause_before_rebooting -> rebooting, substate "foo" -> "bar"
	s.statusUpdateAndVerify(ctx, devs[0], depID, "rebooting", "bar", http.StatusNoContent,
		"rebooting", "bar", "inprogress")

	// G6: mid-cycle re-poll identity (python L491-497): CheckUpdate again with
	// the original artifact/device-type params still returns the same
	// deployment while it's inprogress.
	instrAgain, r2b, err := devs[0].CheckUpdate(ctx, s.APIClient, "bugs-bunny", deviceType)
	require.NoError(err)
	require.Equal(http.StatusOK, r2b.StatusCode)
	require.Equal(depID, instrAgain.GetId())

	// rebooting -> success
	s.statusUpdateAndVerify(ctx, devs[0], depID, "success", "", http.StatusNoContent,
		"success", "bar", "inprogress")

	// devs[0] deployment already finished
	s.statusUpdateAndVerify(ctx, devs[0], depID, "pending", "", http.StatusBadRequest,
		"success", "bar", "inprogress")

	// devs[3]: next update available
	_, r2, err = devs[3].CheckUpdate(ctx, s.APIClient, "bugs-bunny", deviceType)
	require.NoError(err)
	require.Equal(http.StatusOK, r2.StatusCode)

	// devs[3]: pending -> failure, deployment: inprogress -> finished
	s.statusUpdateAndVerify(ctx, devs[3], depID, "failure", "", http.StatusNoContent,
		"failure", "", "finished")

	// G7: post-finish override (python L516-520): reporting a new terminal
	// status for a device whose deployment has already finished is still
	// accepted; the deployment status stays "finished" but the stats reflect
	// the override.
	assert := assert.New(s.T())
	statsBefore, err := s.getStats(ctx, depID)
	require.NoError(err)

	r3, err := s.APIClient.DeploymentsDeviceAPIAPI.
		UpdateDeploymentStatus(common.JWTAuthContext(ctx, devs[0].Token), depID).
		DeploymentStatus(client.DeploymentStatus{Status: "failure"}).
		Execute()
	require.NoError(err)
	require.Equal(http.StatusNoContent, r3.StatusCode)

	statsAfter, err := s.getStats(ctx, depID)
	require.NoError(err)
	assert.Equal(statsBefore.Success-1, statsAfter.Success)
	assert.Equal(statsBefore.Failure+1, statsAfter.Failure)

	dep, err := s.getDeployment(ctx, depID)
	require.NoError(err)
	assert.Equal("finished", dep.Status)
}
