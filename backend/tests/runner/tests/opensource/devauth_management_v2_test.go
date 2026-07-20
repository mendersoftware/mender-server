//nolint:all // This is all test code
package opensource

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/utils/types"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DevauthManagementV2Suite ports (a subset of) the OS device auth tests
// from mender-server/backend/tests/integration/test_devauth.py. Enterprise
// branches (useExistingTenant()) and TestDevAuthCli (covered by acceptance
// suites) are intentionally not ported.
type DevauthManagementV2Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string

	// createdDevices tracks every device this suite creates, so
	// TearDownSuite can decommission them again. The suite runs against a
	// shared environment (see docker_compose_environment.go), and other
	// suites (e.g. InventoryManagementV2) assert on exact, non-delta
	// device/inventory counts -- leaving devices behind would leak into
	// those.
	createdDevices []string
}

func (i *BackendIntegrationSuite) TestDevauthManagementV2() {
	suite.Run(i.T(), &DevauthManagementV2Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (s *DevauthManagementV2Suite) SetupSuite() {
	require := require.New(s.T())

	ctx := common.BasicAuthContext(s.T().Context(), s.User)
	token, r, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.Equal(http.StatusOK, r.StatusCode)
	require.NotEmpty(token)
	s.JWT = token
}

// TearDownSuite decommissions every device this suite created, so this
// suite doesn't leak devices into other suites that assert exact (not
// delta) device/inventory counts against the shared environment.
func (s *DevauthManagementV2Suite) TearDownSuite() {
	if s.JWT == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx = common.JWTAuthContext(ctx, s.JWT)
	for _, id := range s.createdDevices {
		_, _ = s.deleteDevice(ctx, id)
	}
}

func (s *DevauthManagementV2Suite) trackDevice(id string) {
	if id != "" {
		s.createdDevices = append(s.createdDevices, id)
	}
}

// ---------------------------------------------------------------------
// TestPreauth (test_devauth.py::TestPreauth)
// ---------------------------------------------------------------------

func (s *DevauthManagementV2Suite) TestPreauthOk() {
	// ported from test_devauth.py::TestPreauth::test_ok
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	count, err := s.countDevices(ctx, model.DevStatusPreauth)
	require.NoError(err)

	type devEntry struct {
		mac, sn string
		idData  string
		kp      *common.KeyPair
		id      string
	}

	devs := make([]*devEntry, 0, len(common.KeyKinds))
	for _, kind := range common.KeyKinds {
		kp, err := common.NewKeyPair(kind)
		require.NoError(err)
		mac, sn := randMacSn()
		devs = append(devs, &devEntry{mac: mac, sn: sn, idData: idDataJSON(mac, sn), kp: kp})
	}

	for _, d := range devs {
		id, r, err := s.preauthorize(ctx, d.mac, d.sn, d.kp.PublicKeyPEM(), false)
		require.NoError(err)
		require.Equal(http.StatusCreated, r.StatusCode)
		d.id = id
		s.trackDevice(id)
	}

	newCount, err := s.countDevices(ctx, model.DevStatusPreauth)
	require.NoError(err)
	assert.Equal(count+int32(len(devs)), newCount)

	for _, d := range devs {
		dev, err := s.getDevice(ctx, d.id)
		require.NoError(err)
		assert.Equal(model.DevStatusPreauth, dev.GetStatus())
		require.Len(dev.AuthSets, 1)
		aset := dev.AuthSets[0]
		assert.Equal(d.mac, fmt.Sprint(aset.IdentityData["mac"]))
		assert.Equal(d.sn, fmt.Sprint(aset.IdentityData["sn"]))
		assert.Equal(model.DevStatusPreauth, aset.GetStatus())

		// the actual device can obtain an auth token
		device := common.NewDeviceFromKeyPair(d.kp, d.idData)
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.True(ok)

		outdev, err := s.getDevice(ctx, d.id)
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, outdev.GetStatus())
		require.Len(outdev.AuthSets, 1)
		assert.Equal(model.DevStatusAccepted, outdev.AuthSets[0].GetStatus())
	}

	// send the preauth requests again with fresh keys: since the devices
	// show up with new keys they are immediately accepted with a new
	// authset, and the old one is rejected
	for _, d := range devs {
		kp, err := common.NewKeyPair(common.KeyKindRSA)
		require.NoError(err)
		d.kp = kp

		_, r, err := s.preauthorize(ctx, d.mac, d.sn, kp.PublicKeyPEM(), true)
		require.NoError(err)
		assert.Equal(http.StatusCreated, r.StatusCode)
	}

	for _, d := range devs {
		dev, err := s.getDevice(ctx, d.id)
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, dev.GetStatus())
		require.Len(dev.AuthSets, 2)

		device := common.NewDeviceFromKeyPair(d.kp, d.idData)
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.True(ok)

		outdev, err := s.getDevice(ctx, d.id)
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, outdev.GetStatus())
		require.Len(outdev.AuthSets, 2)

		foundAccepted := false
		for _, aset := range outdev.AuthSets {
			if aset.GetStatus() == model.DevStatusAccepted {
				foundAccepted = true
				assert.True(comparePubkeysPEM(aset.GetPubkey(), d.kp.PublicKeyPEM()))
				break
			}
		}
		assert.True(foundAccepted)
	}
}

func (s *DevauthManagementV2Suite) TestPreauthFailDuplicate() {
	// ported from test_devauth.py::TestPreauth::test_fail_duplicate
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	countBefore, err := s.countDevices(ctx, "")
	require.NoError(err)

	mac, sn := randMacSn()
	idData := idDataJSON(mac, sn)
	kp, err := common.NewKeyPair(common.KeyKindRSA)
	require.NoError(err)

	device := common.NewDeviceFromKeyPair(kp, idData)
	ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	assert.False(ok)

	newKP, err := common.NewKeyPair(common.KeyKindRSA)
	require.NoError(err)

	_, r, err := s.preauthorize(ctx, mac, sn, newKP.PublicKeyPEM(), false)
	require.Error(err)
	assert.Equal(http.StatusConflict, r.StatusCode)

	// device list is unmodified: the failed duplicate did not add a device
	countAfter, err := s.countDevices(ctx, "")
	require.NoError(err)
	assert.Equal(countBefore+1, countAfter)

	// existing device has no new auth sets
	apiDev, err := s.getDeviceByMacSn(ctx, mac, sn)
	require.NoError(err)
	s.trackDevice(apiDev.GetId())
	require.Len(apiDev.AuthSets, 1)
	assert.True(comparePubkeysPEM(apiDev.AuthSets[0].GetPubkey(), kp.PublicKeyPEM()))
	assert.Equal(model.DevStatusPending, apiDev.AuthSets[0].GetStatus())
}

func (s *DevauthManagementV2Suite) TestPreauthFailBadRequest() {
	// ported from test_devauth.py::TestPreauth::test_fail_bad_request
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	kp, err := common.NewKeyPair(common.KeyKindRSA)
	require.NoError(err)

	// id data not json: identity_data is a raw string instead of an
	// object. The generated client can't build this (identity_data is
	// typed as an object), so this one goes over a raw request.
	body, err := json.Marshal(map[string]any{
		"identity_data": `{"mac": "foo"}`,
		"pubkey":        kp.PublicKeyPEM(),
	})
	require.NoError(err)
	resp, err := common.RawRequest(ctx, s.APIClient, http.MethodPost, "/api/management/v2/devauth/devices", body)
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusBadRequest, resp.StatusCode)

	// not a valid key
	mac, sn := randMacSn()
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementPreauthorize(ctx).
		PreAuthSet(client.PreAuthSet{
			IdentityData: client.IdentityData{Mac: types.Pointer(mac), Sn: types.Pointer(sn)},
			Pubkey:       "not a public key",
		}).
		Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)
}

// ---------------------------------------------------------------------
// devs_authsets fixture (test_devauth.py::make_devs_with_authsets and
// friends). Kept local to this file, like the Python original keeps it
// local to test_devauth.py rather than in testutils/common.py.
// ---------------------------------------------------------------------

type devFixtureAuthset struct {
	ID     string
	Status string
	KP     *common.KeyPair
}

type devFixture struct {
	ID       string
	Mac      string
	Sn       string
	IDData   string
	Status   string
	Authsets []*devFixtureAuthset
}

func (s *DevauthManagementV2Suite) createFirstAuthset(ctx context.Context, kind string) (*devFixture, error) {
	mac, sn := randMacSn()
	idData := idDataJSON(mac, sn)
	kp, err := common.NewKeyPair(kind)
	if err != nil {
		return nil, err
	}

	device := common.NewDeviceFromKeyPair(kp, idData)
	if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
		return nil, err
	}

	apiDev, err := s.getDeviceByMacSn(ctx, mac, sn)
	if err != nil {
		return nil, err
	}
	aset := findAuthsetByPubkey(apiDev.AuthSets, kp.PublicKeyPEM())
	if aset == nil {
		return nil, fmt.Errorf("authset not found after auth request for mac=%s sn=%s", mac, sn)
	}
	s.trackDevice(apiDev.GetId())

	return &devFixture{
		ID:     apiDev.GetId(),
		Mac:    mac,
		Sn:     sn,
		IDData: idData,
		Status: model.DevStatusPending,
		Authsets: []*devFixtureAuthset{
			{ID: aset.GetId(), Status: model.DevStatusPending, KP: kp},
		},
	}, nil
}

func (s *DevauthManagementV2Suite) addAuthset(ctx context.Context, dev *devFixture, kind string) error {
	kp, err := common.NewKeyPair(kind)
	if err != nil {
		return err
	}

	device := common.NewDeviceFromKeyPair(kp, dev.IDData)
	if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
		return err
	}

	apiDev, err := s.getDevice(ctx, dev.ID)
	if err != nil {
		return err
	}
	aset := findAuthsetByPubkey(apiDev.AuthSets, kp.PublicKeyPEM())
	if aset == nil {
		return fmt.Errorf("new authset not found for device %s", dev.ID)
	}

	dev.Authsets = append(dev.Authsets, &devFixtureAuthset{ID: aset.GetId(), Status: model.DevStatusPending, KP: kp})
	return nil
}

func (s *DevauthManagementV2Suite) makePendingDevice(ctx context.Context, kind string, numAuthsets int) (*devFixture, error) {
	dev, err := s.createFirstAuthset(ctx, kind)
	if err != nil {
		return nil, err
	}
	for i := 1; i < numAuthsets; i++ {
		if err := s.addAuthset(ctx, dev, kind); err != nil {
			return nil, err
		}
	}
	return dev, nil
}

func (s *DevauthManagementV2Suite) makeAcceptedDevice(ctx context.Context, kind string, numAuthsets, numAccepted int) (*devFixture, error) {
	dev, err := s.makePendingDevice(ctx, kind, numAuthsets)
	if err != nil {
		return nil, err
	}
	for i := 0; i < numAccepted; i++ {
		if err := s.setAuthsetStatus(ctx, dev.ID, dev.Authsets[i].ID, model.DevStatusAccepted); err != nil {
			return nil, err
		}
		dev.Authsets[i].Status = model.DevStatusAccepted
	}
	dev.Status = model.DevStatusAccepted
	return dev, nil
}

func (s *DevauthManagementV2Suite) makeRejectedDevice(ctx context.Context, kind string, numAuthsets int) (*devFixture, error) {
	dev, err := s.makePendingDevice(ctx, kind, numAuthsets)
	if err != nil {
		return nil, err
	}
	for _, aset := range dev.Authsets {
		if err := s.setAuthsetStatus(ctx, dev.ID, aset.ID, model.DevStatusRejected); err != nil {
			return nil, err
		}
		aset.Status = model.DevStatusRejected
	}
	dev.Status = model.DevStatusRejected
	return dev, nil
}

func (s *DevauthManagementV2Suite) makePreauthdDevice(ctx context.Context, kind string) (*devFixture, error) {
	kp, err := common.NewKeyPair(kind)
	if err != nil {
		return nil, err
	}
	mac, sn := randMacSn()

	id, r, err := s.preauthorize(ctx, mac, sn, kp.PublicKeyPEM(), false)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected preauthorize status %d", r.StatusCode)
	}
	s.trackDevice(id)

	apiDev, err := s.getDevice(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(apiDev.AuthSets) != 1 {
		return nil, fmt.Errorf("expected 1 authset for preauthd device %s, got %d", id, len(apiDev.AuthSets))
	}

	return &devFixture{
		ID:     id,
		Mac:    mac,
		Sn:     sn,
		IDData: idDataJSON(mac, sn),
		Status: model.DevStatusPreauth,
		Authsets: []*devFixtureAuthset{
			{ID: apiDev.AuthSets[0].GetId(), Status: model.DevStatusPreauth, KP: kp},
		},
	}, nil
}

func (s *DevauthManagementV2Suite) makePreauthdDeviceWithPending(ctx context.Context, kind string, numPending int) (*devFixture, error) {
	dev, err := s.makePreauthdDevice(ctx, kind)
	if err != nil {
		return nil, err
	}
	for i := 0; i < numPending; i++ {
		if err := s.addAuthset(ctx, dev, common.KeyKindRSA); err != nil {
			return nil, err
		}
	}
	return dev, nil
}

// makeDevsWithAuthsets ports make_devs_with_authsets: a good number of
// devices, some with >1 authsets, in every status.
func (s *DevauthManagementV2Suite) makeDevsWithAuthsets(ctx context.Context) ([]*devFixture, error) {
	var devs []*devFixture
	add := func(dev *devFixture, err error) error {
		if err != nil {
			return err
		}
		devs = append(devs, dev)
		return nil
	}

	// vanilla 'pending' devices, single authset
	for i := 0; i < 3; i++ {
		if err := add(s.makePendingDevice(ctx, common.KeyKindRSA, 1)); err != nil {
			return nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.makePendingDevice(ctx, common.KeyKindECP256, 1)); err != nil {
			return nil, err
		}
	}
	if err := add(s.makePendingDevice(ctx, common.KeyKindEd25519, 1)); err != nil {
		return nil, err
	}

	// pending devices with >1 authsets
	for i := 0; i < 2; i++ {
		if err := add(s.makePendingDevice(ctx, common.KeyKindRSA, 3)); err != nil {
			return nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.makePendingDevice(ctx, common.KeyKindECP256, 3)); err != nil {
			return nil, err
		}
	}
	if err := add(s.makePendingDevice(ctx, common.KeyKindEd25519, 3)); err != nil {
		return nil, err
	}

	// accepted devices, single authset
	for i := 0; i < 3; i++ {
		if err := add(s.makeAcceptedDevice(ctx, common.KeyKindRSA, 1, 1)); err != nil {
			return nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.makeAcceptedDevice(ctx, common.KeyKindECP256, 1, 1)); err != nil {
			return nil, err
		}
	}
	if err := add(s.makeAcceptedDevice(ctx, common.KeyKindEd25519, 1, 1)); err != nil {
		return nil, err
	}

	// accepted devices with >1 authsets
	for i := 0; i < 2; i++ {
		if err := add(s.makeAcceptedDevice(ctx, common.KeyKindRSA, 3, 1)); err != nil {
			return nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.makeAcceptedDevice(ctx, common.KeyKindECP256, 2, 1)); err != nil {
			return nil, err
		}
	}
	if err := add(s.makeAcceptedDevice(ctx, common.KeyKindEd25519, 2, 1)); err != nil {
		return nil, err
	}

	// rejected devices
	for i := 0; i < 2; i++ {
		if err := add(s.makeRejectedDevice(ctx, common.KeyKindRSA, 3)); err != nil {
			return nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.makeRejectedDevice(ctx, common.KeyKindECP256, 2)); err != nil {
			return nil, err
		}
	}
	if err := add(s.makeRejectedDevice(ctx, common.KeyKindEd25519, 2)); err != nil {
		return nil, err
	}

	// preauthd devices
	if err := add(s.makePreauthdDevice(ctx, common.KeyKindRSA)); err != nil {
		return nil, err
	}
	if err := add(s.makePreauthdDevice(ctx, common.KeyKindECP256)); err != nil {
		return nil, err
	}
	if err := add(s.makePreauthdDevice(ctx, common.KeyKindEd25519)); err != nil {
		return nil, err
	}

	// preauthd devices with extra 'pending' sets
	for i := 0; i < 2; i++ {
		if err := add(s.makePreauthdDeviceWithPending(ctx, common.KeyKindRSA, 2)); err != nil {
			return nil, err
		}
	}
	if err := add(s.makePreauthdDeviceWithPending(ctx, common.KeyKindECP256, 2)); err != nil {
		return nil, err
	}
	if err := add(s.makePreauthdDeviceWithPending(ctx, common.KeyKindEd25519, 2)); err != nil {
		return nil, err
	}

	// mirrors deviceauth's own device listing order: sorted by (status,
	// id) ascending -- see GetDevices in
	// services/deviceauth/store/mongo/datastore_mongo.go.
	slices.SortStableFunc(devs, func(a, b *devFixture) int {
		if c := cmp.Compare(a.Status, b.Status); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return devs, nil
}

func filterAndPageDevs(devs []*devFixture, page, perPage int, status string) []*devFixture {
	filtered := devs
	if status != "" {
		filtered = nil
		for _, d := range devs {
			if d.Status == status {
				filtered = append(filtered, d)
			}
		}
	}
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	lo := (page - 1) * perPage
	hi := lo + perPage
	if lo > len(filtered) {
		lo = len(filtered)
	}
	if hi > len(filtered) {
		hi = len(filtered)
	}
	return filtered[lo:hi]
}

func findAuthsetByPubkey(asets []client.AuthSet, pubkeyPEM string) *client.AuthSet {
	for i := range asets {
		if asets[i].Pubkey != nil && comparePubkeysPEM(*asets[i].Pubkey, pubkeyPEM) {
			return &asets[i]
		}
	}
	return nil
}

func computeDevStatus(authsets []*devFixtureAuthset) string {
	if len(authsets) == 0 {
		return model.DevStatusNoAuth
	}
	for _, a := range authsets {
		if a.Status == model.DevStatusAccepted {
			return model.DevStatusAccepted
		}
	}
	for _, a := range authsets {
		if a.Status == model.DevStatusPreauth {
			return model.DevStatusPreauth
		}
	}
	for _, a := range authsets {
		if a.Status == model.DevStatusPending {
			return model.DevStatusPending
		}
	}
	return model.DevStatusRejected
}

func removeAuthset(authsets []*devFixtureAuthset, target *devFixtureAuthset) []*devFixtureAuthset {
	out := make([]*devFixtureAuthset, 0, len(authsets))
	for _, a := range authsets {
		if a != target {
			out = append(out, a)
		}
	}
	return out
}

func (s *DevauthManagementV2Suite) compareDev(dev *devFixture, apiDev client.Device) {
	require := require.New(s.T())
	assert := assert.New(s.T())

	assert.Equal(dev.ID, apiDev.GetId())
	require.NotNil(apiDev.IdentityData)
	assert.Equal(dev.Mac, apiDev.IdentityData.GetMac())
	assert.Equal(dev.Sn, apiDev.IdentityData.GetSn())
	assert.Equal(dev.Status, apiDev.GetStatus())

	require.Len(apiDev.AuthSets, len(dev.Authsets))

	// GOTCHA: don't rely on indexing, authsets can get reshuffled
	// depending on actual contents.
	for _, apiAset := range apiDev.AuthSets {
		var match *devFixtureAuthset
		for _, aset := range dev.Authsets {
			if apiAset.Pubkey != nil && comparePubkeysPEM(*apiAset.Pubkey, aset.KP.PublicKeyPEM()) {
				match = aset
				break
			}
		}
		require.NotNil(match, "no matching local authset for api authset %s", apiAset.GetId())
		assert.Equal(match.ID, apiAset.GetId())
		assert.Equal(match.Status, apiAset.GetStatus())
	}
}

func (s *DevauthManagementV2Suite) compareDevs(devs []*devFixture, apiDevs []client.Device) {
	require := require.New(s.T())
	require.Len(apiDevs, len(devs))
	for i := range apiDevs {
		s.compareDev(devs[i], apiDevs[i])
	}
}

func (s *DevauthManagementV2Suite) verifyDevAfterStatusUpdate(ctx context.Context, dev *devFixture) {
	require := require.New(s.T())

	apiDev, err := s.getDevice(ctx, dev.ID)
	require.NoError(err)
	s.compareDev(dev, *apiDev)

	err = common.RetryUntil(ctx, 10*time.Second, 250*time.Millisecond, func() (bool, error) {
		inv, _, err := s.APIClient.DeviceInventoryManagementAPIAPI.GetDeviceInventory(ctx, dev.ID).Execute()
		if err != nil || inv == nil {
			return false, nil
		}
		for _, a := range inv.Attributes {
			if a.GetScope() == client.IDENTITY && a.GetName() == "status" {
				return a.GetValue().String != nil && *a.GetValue().String == dev.Status, nil
			}
		}
		return false, nil
	})
	require.NoError(err, "timeout waiting for inventory status to become %q for device %s", dev.Status, dev.ID)
}

func (s *DevauthManagementV2Suite) verifyDevProvisioned(ctx context.Context, dev *devFixture) {
	require := require.New(s.T())
	err := common.RetryUntil(ctx, 10*time.Second, 250*time.Millisecond, func() (bool, error) {
		_, r, err := s.APIClient.DeviceInventoryManagementAPIAPI.GetDeviceInventory(ctx, dev.ID).Execute()
		return err == nil && r != nil && r.StatusCode == http.StatusOK, nil
	})
	require.NoError(err, "timeout waiting for device %s to get provisioned in inventory", dev.ID)
}

// ---------------------------------------------------------------------
// TestDeviceMgmt (test_devauth.py::TestDeviceMgmt)
// ---------------------------------------------------------------------

func (s *DevauthManagementV2Suite) TestDeviceMgmtOkGetDevices() {
	// ported from test_devauth.py::TestDeviceMgmt::test_ok_get_devices
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// the shared environment isn't reset between tests, so restrict the
	// listing to the devices this test created (the Python original runs
	// against a freshly wiped DB per test via the clean_migrated_mongo
	// fixture and doesn't need to).
	ids := deviceIDs(devs)

	cases := []struct {
		status        string
		page, perPage int
	}{
		{"", 0, 0},
		{model.DevStatusPending, 0, 0},
		{model.DevStatusAccepted, 0, 0},
		{model.DevStatusRejected, 0, 0},
		{model.DevStatusPreauth, 0, 0},
		{"", 1, 10},
		{"", 3, 10},
		{"", 2, 5},
		{model.DevStatusAccepted, 1, 4},
		{model.DevStatusAccepted, 2, 4},
		{model.DevStatusAccepted, 5, 2},
		{model.DevStatusPending, 2, 2},
	}

	for _, tc := range cases {
		apiDevs, err := s.listDevices(ctx, tc.status, int32(tc.page), int32(tc.perPage), ids...)
		require.NoError(err)

		refDevs := filterAndPageDevs(devs, tc.page, tc.perPage, tc.status)
		s.compareDevs(refDevs, apiDevs)
	}
}

func (s *DevauthManagementV2Suite) TestDeviceMgmtGetDevice() {
	// ported from test_devauth.py::TestDeviceMgmt::test_get_device
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, dev := range devs {
		apiDev, err := s.getDevice(ctx, dev.ID)
		require.NoError(err)
		s.compareDev(dev, *apiDev)
	}

	for _, id := range []string{"foo", "bar"} {
		_, r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, id).Execute()
		require.Error(err)
		assert.Equal(http.StatusNotFound, r.StatusCode)
	}
}

func (s *DevauthManagementV2Suite) TestDeviceMgmtDeleteDeviceOk() {
	// ported from test_devauth.py::TestDeviceMgmt::test_delete_device_ok
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// decommission a pending device
	pending := filterAndPageDevs(devs, 0, len(devs), model.DevStatusPending)
	require.NotEmpty(pending)
	devPending := pending[0]

	status, err := s.deleteDevice(ctx, devPending.ID)
	require.NoError(err)
	assert.Equal(http.StatusNoContent, status)
	require.True(s.waitDeviceGone(ctx, devPending.ID), "timeout waiting for device auth to be deleted")

	// log in an accepted device
	accepted := filterAndPageDevs(devs, 0, len(devs), model.DevStatusAccepted)
	require.NotEmpty(accepted)
	devAcc := accepted[0]

	var acceptedAset *devFixtureAuthset
	for _, a := range devAcc.Authsets {
		if a.Status == model.DevStatusAccepted {
			acceptedAset = a
			break
		}
	}
	require.NotNil(acceptedAset)

	device := common.NewDeviceFromKeyPair(acceptedAset.KP, devAcc.IDData)
	ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	require.True(ok)
	dtoken := device.Token

	// decommission the accepted device
	status, err = s.deleteDevice(ctx, devAcc.ID)
	require.NoError(err)
	assert.Equal(http.StatusNoContent, status)

	// verify the device lost access
	devCtx := common.JWTAuthContext(ctx, dtoken)
	_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.CheckUpdate(devCtx).
		DeviceType("foo").ArtifactName("bar").Execute()
	require.Error(err)
	assert.Equal(http.StatusUnauthorized, r.StatusCode)

	require.True(s.waitDeviceGone(ctx, devAcc.ID), "timeout waiting for device auth to be deleted")
}

func (s *DevauthManagementV2Suite) TestDeviceMgmtDeleteDeviceNotFound() {
	// ported from test_devauth.py::TestDeviceMgmt::test_delete_device_not_found
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	status, err := s.deleteDevice(ctx, "foo")
	require.Error(err)
	assert.Equal(http.StatusNotFound, status)

	// device list unmodified
	apiDevs, err := s.listDevices(ctx, "", 0, int32(len(devs)), deviceIDs(devs)...)
	require.NoError(err)
	s.compareDevs(devs, apiDevs)
}

func (s *DevauthManagementV2Suite) TestDeviceMgmtDeviceCount() {
	// ported from test_devauth.py::TestDeviceMgmt::test_device_count
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	// G3: noauth (test_device_count_simple) - delta-based like the other statuses.
	statuses := []string{"", model.DevStatusPending, model.DevStatusAccepted, model.DevStatusRejected, model.DevStatusPreauth, model.DevStatusNoAuth}

	// CountDevices (unlike ListDevices) has no id filter, and the shared
	// environment isn't reset between tests -- so verify counts as
	// deltas against a baseline taken before this test's own devices
	// exist, rather than asserting the exact count the Python original
	// does against its freshly wiped DB.
	baseline := make(map[string]int32, len(statuses))
	for _, status := range statuses {
		count, err := s.countDevices(ctx, status)
		require.NoError(err)
		baseline[status] = count
	}

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, status := range statuses {
		count, err := s.countDevices(ctx, status)
		require.NoError(err)

		refDevs := filterAndPageDevs(devs, 0, len(devs), status)
		assert.Equal(baseline[status]+int32(len(refDevs)), count)
	}

	_, r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementCountDevices(ctx).Status("foo").Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)
}

// ---------------------------------------------------------------------
// TestAuthsetMgmt (test_devauth.py::TestAuthsetMgmt)
// ---------------------------------------------------------------------

func (s *DevauthManagementV2Suite) TestAuthsetMgmtGetAuthsetStatus() {
	// ported from test_devauth.py::TestAuthsetMgmt::test_get_authset_status
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, dev := range devs {
		for _, aset := range dev.Authsets {
			status, r, err := s.getAuthsetStatus(ctx, dev.ID, aset.ID)
			require.NoError(err)
			require.Equal(http.StatusOK, r.StatusCode)
			assert.Equal(aset.Status, status)
		}
	}

	for _, tc := range []struct{ did, aid string }{
		{devs[0].ID, "foo"},
		{"foo", "bar"},
	} {
		_, r, err := s.getAuthsetStatus(ctx, tc.did, tc.aid)
		require.Error(err)
		assert.Equal(http.StatusNotFound, r.StatusCode)
	}
}

func (s *DevauthManagementV2Suite) TestAuthsetMgmtPutStatusAccept() {
	// ported from test_devauth.py::TestAuthsetMgmt::test_put_status_accept
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// select interesting devices - pending, rejected, or accepted/preauthd
	// with extra authsets
	var candidates []*devFixture
	for _, status := range []string{model.DevStatusPending, model.DevStatusRejected, model.DevStatusAccepted, model.DevStatusPreauth} {
		found := filterAndPageDevs(devs, 0, len(devs), status)
		if status == model.DevStatusAccepted || status == model.DevStatusPreauth {
			var multi []*devFixture
			for _, d := range found {
				if len(d.Authsets) > 1 {
					multi = append(multi, d)
				}
			}
			found = multi
		}
		candidates = append(candidates, found...)
	}

	for _, dev := range candidates {
		// for accepted devs - first actually get a device token
		var dtoken string
		if dev.Status == model.DevStatusAccepted {
			var accepted *devFixtureAuthset
			for _, a := range dev.Authsets {
				if a.Status == model.DevStatusAccepted {
					accepted = a
					break
				}
			}
			require.NotNil(accepted)

			device := common.NewDeviceFromKeyPair(accepted.KP, dev.IDData)
			ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
			require.NoError(err)
			require.True(ok)
			dtoken = device.Token
		}

		// find some pending or rejected authset
		var target *devFixtureAuthset
		for _, a := range dev.Authsets {
			if a.Status == model.DevStatusPending || a.Status == model.DevStatusRejected {
				target = a
				break
			}
		}
		require.NotNil(target)

		// accept the authset
		require.NoError(s.setAuthsetStatus(ctx, dev.ID, target.ID, model.DevStatusAccepted))

		// in case of originally accepted devs: the original authset must
		// be rejected now (ported verbatim: the Python original only
		// rejects the old authset for "accepted" devices, not
		// "preauthorized" ones)
		if dev.Status == model.DevStatusAccepted {
			for _, a := range dev.Authsets {
				if a.Status == dev.Status {
					a.Status = model.DevStatusRejected
					break
				}
			}
		}

		dev.Status = model.DevStatusAccepted
		target.Status = model.DevStatusAccepted

		s.verifyDevAfterStatusUpdate(ctx, dev)

		if dtoken != "" {
			devCtx := common.JWTAuthContext(ctx, dtoken)
			_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.CheckUpdate(devCtx).
				DeviceType("foo").ArtifactName("bar").Execute()
			require.Error(err)
			assert.Equal(http.StatusUnauthorized, r.StatusCode)
		}

		s.verifyDevProvisioned(ctx, dev)
	}
}

func (s *DevauthManagementV2Suite) TestAuthsetMgmtPutStatusReject() {
	// ported from test_devauth.py::TestAuthsetMgmt::test_put_status_reject
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	var candidates []*devFixture
	for _, status := range []string{model.DevStatusPending, model.DevStatusAccepted} {
		candidates = append(candidates, filterAndPageDevs(devs, 0, len(devs), status)...)
	}

	for _, dev := range candidates {
		var aset *devFixtureAuthset
		if dev.Status == model.DevStatusAccepted {
			for _, a := range dev.Authsets {
				if a.Status == dev.Status {
					aset = a
					break
				}
			}
			require.NotNil(aset)
		} else {
			aset = dev.Authsets[0]
		}

		// for accepted devs, also have an active device and check it
		// loses api access
		var dtoken string
		if dev.Status == model.DevStatusAccepted {
			device := common.NewDeviceFromKeyPair(aset.KP, dev.IDData)
			ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
			require.NoError(err)
			require.True(ok)
			dtoken = device.Token
		}

		if aset.Status == model.DevStatusAccepted || aset.Status == model.DevStatusPending {
			require.NoError(s.setAuthsetStatus(ctx, dev.ID, aset.ID, model.DevStatusRejected))
		}

		aset.Status = model.DevStatusRejected

		rejCount := 0
		for _, a := range dev.Authsets {
			if a.ID != aset.ID && a.Status == model.DevStatusRejected {
				rejCount++
			}
		}
		if rejCount == len(dev.Authsets)-1 {
			dev.Status = model.DevStatusRejected
		} else {
			dev.Status = model.DevStatusPending
		}

		s.verifyDevAfterStatusUpdate(ctx, dev)

		if dtoken != "" {
			devCtx := common.JWTAuthContext(ctx, dtoken)
			_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.CheckUpdate(devCtx).
				DeviceType("foo").ArtifactName("bar").Execute()
			require.Error(err)
			assert.Equal(http.StatusUnauthorized, r.StatusCode)
		}

		// G1: post-reject re-auth (python test_device.py::test_device_accept_reject_cycle
		// L122-123): a new auth request from the same device after reject
		// gets rejected, and no new authset appears.
		device := common.NewDeviceFromKeyPair(aset.KP, dev.IDData)
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.False(ok)

		apiDevAfter, err := s.getDevice(ctx, dev.ID)
		require.NoError(err)
		s.compareDev(dev, *apiDevAfter)
	}
}

func (s *DevauthManagementV2Suite) TestAuthsetMgmtPutStatusFailed() {
	// ported from test_devauth.py::TestAuthsetMgmt::test_put_status_failed
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// not found: valid device, bogus authset
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, devs[0].ID, "foo").
		Status(client.Status{Status: model.DevStatusAccepted}).
		Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// not found: bogus device
	r, err = s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, "foo", "bar").
		Status(client.Status{Status: model.DevStatusAccepted}).
		Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// not found: bogus device, status "rejected" (G2: test_device_reject_nonexistent)
	r, err = s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, "foo", "bar").
		Status(client.Status{Status: model.DevStatusRejected}).
		Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// bad request - invalid status
	r, err = s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, devs[0].ID, devs[0].Authsets[0].ID).
		Status(client.Status{Status: "invalid"}).
		Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)

	// bad request - invalid payload: the generated client can't send an
	// arbitrary body via the typed Status model, so this one goes over a
	// raw request.
	reqPath := fmt.Sprintf("/api/management/v2/devauth/devices/%s/auth/%s/status", devs[0].ID, devs[0].Authsets[0].ID)
	resp, err := common.RawRequest(ctx, s.APIClient, http.MethodPut, reqPath, []byte(`{"foo": "bar"}`))
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *DevauthManagementV2Suite) TestAuthsetMgmtDeleteStatus() {
	// ported from test_devauth.py::TestAuthsetMgmt::test_delete_status
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, dev := range devs {
		// for accepted or preauthd devices, target the accepted/preauthd
		// set; otherwise just select the first one
		var aset *devFixtureAuthset
		if dev.Status == model.DevStatusAccepted || dev.Status == model.DevStatusPreauth {
			for _, a := range dev.Authsets {
				if a.Status == dev.Status {
					aset = a
					break
				}
			}
			require.NotNil(aset)
		} else {
			aset = dev.Authsets[0]
		}

		var dtoken string
		if dev.Status == model.DevStatusAccepted {
			device := common.NewDeviceFromKeyPair(aset.KP, dev.IDData)
			ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
			require.NoError(err)
			require.True(ok)
			dtoken = device.Token
		}

		status, err := s.deleteAuthset(ctx, dev.ID, aset.ID)
		require.NoError(err)
		require.Equal(http.StatusNoContent, status)

		dev.Authsets = removeAuthset(dev.Authsets, aset)

		if dev.Status == model.DevStatusPreauth {
			// removing the preauth authset: the device should be
			// completely gone
			_, r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
				DeviceAuthManagementGetDevice(ctx, dev.ID).Execute()
			require.Error(err)
			assert.Equal(http.StatusNotFound, r.StatusCode)
			// ... including from inventory (async cascade; the removed
			// integration-repo preauth test asserted this side too). Check
			// specifically for 404 rather than treating any error as
			// "gone" -- a transient request error isn't proof of removal.
			goneErr := common.RetryUntil(ctx, 30*time.Second, time.Second, func() (bool, error) {
				_, r, _ := s.APIClient.
					DeviceInventoryManagementAPIAPI.
					GetDeviceInventory(ctx, dev.ID).Execute()
				return r != nil && r.StatusCode == http.StatusNotFound, nil
			})
			assert.NoError(goneErr, "device %s still present in inventory after preauth authset removal", dev.ID)
			// ported quirk: the Python original (do_test_delete_status)
			// also returns here, ending the whole test on the first
			// preauthorized device -- since devs_authsets is sorted by
			// status ascending (accepted, pending, preauthorized,
			// rejected), rejected devices are never exercised by this
			// test either in the original or in this port.
			return
		}

		dev.Status = computeDevStatus(dev.Authsets)
		s.verifyDevAfterStatusUpdate(ctx, dev)

		if dtoken != "" {
			devCtx := common.JWTAuthContext(ctx, dtoken)
			_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.CheckUpdate(devCtx).
				DeviceType("foo").ArtifactName("bar").Execute()
			require.Error(err)
			assert.Equal(http.StatusUnauthorized, r.StatusCode)
		}
	}
}

func (s *DevauthManagementV2Suite) TestAuthsetMgmtDeleteStatusFailed() {
	// ported from test_devauth.py::TestAuthsetMgmt::test_delete_status_failed
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// not found: valid device, bogus authset
	status, err := s.deleteAuthset(ctx, devs[0].ID, "foo")
	require.Error(err)
	assert.Equal(http.StatusNotFound, status)

	// not found: bogus device
	status, err = s.deleteAuthset(ctx, "foo", "bar")
	require.Error(err)
	assert.Equal(http.StatusNotFound, status)
}

// ---------------------------------------------------------------------
// TestAuthReq (test_devauth.py::TestAuthReq)
// ---------------------------------------------------------------------

func (s *DevauthManagementV2Suite) TestAuthReqSubmitAccept() {
	// ported from test_devauth.py::TestAuthReq::test_submit_accept
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	type devEntry struct {
		mac, sn string
		idData  string
		kp      *common.KeyPair
	}

	devs := make([]*devEntry, 0, len(common.KeyKinds))
	for _, kind := range common.KeyKinds {
		kp, err := common.NewKeyPair(kind)
		require.NoError(err)
		mac, sn := randMacSn()
		devs = append(devs, &devEntry{mac: mac, sn: sn, idData: idDataJSON(mac, sn), kp: kp})
	}

	for _, d := range devs {
		device := common.NewDeviceFromKeyPair(d.kp, d.idData)
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.False(ok)
	}

	for _, d := range devs {
		apiDev, err := s.getDeviceByMacSn(ctx, d.mac, d.sn)
		require.NoError(err)
		s.trackDevice(apiDev.GetId())
		require.Len(apiDev.AuthSets, 1)
		assert.Equal(model.DevStatusPending, apiDev.AuthSets[0].GetStatus())

		// accept and get/verify token
		require.NoError(s.setAuthsetStatus(ctx, apiDev.GetId(), apiDev.AuthSets[0].GetId(), model.DevStatusAccepted))

		device := common.NewDeviceFromKeyPair(d.kp, d.idData)
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		require.True(ok)

		parts := strings.Split(device.Token, ".")
		require.Len(parts, 3)
		payloadRaw, err := base64.RawStdEncoding.DecodeString(parts[1])
		require.NoError(err)

		var payload map[string]any
		require.NoError(json.Unmarshal(payloadRaw, &payload))

		// standard claims
		assert.Equal(apiDev.GetId(), payload["sub"])
		assert.Equal("Mender", payload["iss"])
		assert.NotEmpty(payload["jti"])
		assert.NotNil(payload["exp"])

		// custom claims
		assert.Equal(true, payload["mender.device"])

		// G4: JWT header typ (python test_token.py::test_token_claims)
		headerRaw, err := base64.RawStdEncoding.DecodeString(parts[0])
		require.NoError(err)
		var header map[string]any
		require.NoError(json.Unmarshal(headerRaw, &header))
		assert.Equal("JWT", header["typ"])
	}
}

// ---------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------

func randMacSn() (string, string) {
	return uuid.NewString(), uuid.NewString()
}

func idDataJSON(mac, sn string) string {
	b, _ := json.Marshal(map[string]string{"mac": mac, "sn": sn})
	return string(b)
}

func comparePubkeysPEM(a, b string) bool {
	da, _ := pem.Decode([]byte(a))
	db, _ := pem.Decode([]byte(b))
	if da == nil || db == nil {
		return false
	}
	return bytes.Equal(da.Bytes, db.Bytes)
}

func (s *DevauthManagementV2Suite) preauthorize(ctx context.Context, mac, sn, pubkeyPEM string, force bool) (string, *http.Response, error) {
	preAuthSet := client.PreAuthSet{
		IdentityData: client.IdentityData{Mac: types.Pointer(mac), Sn: types.Pointer(sn)},
		Pubkey:       pubkeyPEM,
	}
	if force {
		preAuthSet.Force = types.Pointer(true)
	}

	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementPreauthorize(ctx).
		PreAuthSet(preAuthSet).
		Execute()
	if err != nil || r == nil {
		return "", r, err
	}

	// V1: guard against an empty Location header instead of letting
	// path.Base silently turn it into ".".
	loc := r.Header.Get("Location")
	if loc == "" {
		return "", r, fmt.Errorf("preauthorize: missing Location header (status %d)", r.StatusCode)
	}
	return path.Base(loc), r, nil
}

func (s *DevauthManagementV2Suite) countDevices(ctx context.Context, status string) (int32, error) {
	req := s.APIClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementCountDevices(ctx)
	if status != "" {
		req = req.Status(status)
	}
	count, _, err := req.Execute()
	if err != nil {
		return 0, err
	}
	return count.GetCount(), nil
}

func (s *DevauthManagementV2Suite) getDevice(ctx context.Context, id string) (*client.Device, error) {
	dev, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementGetDevice(ctx, id).Execute()
	return dev, err
}

func (s *DevauthManagementV2Suite) getDeviceByMacSn(ctx context.Context, mac, sn string) (*client.Device, error) {
	const perPage int32 = 500
	var page int32 = 1
	for {
		devices, err := s.listDevices(ctx, "", page, perPage)
		if err != nil {
			return nil, err
		}
		for i := range devices {
			d := devices[i]
			if d.IdentityData != nil && d.IdentityData.GetMac() == mac && d.IdentityData.GetSn() == sn {
				return &d, nil
			}
		}
		if len(devices) < int(perPage) {
			return nil, fmt.Errorf("device not found by id data (mac=%s sn=%s)", mac, sn)
		}
		page++
	}
}

// listDevices lists devices, optionally restricted to a set of device ids.
// The shared environment isn't reset between tests (see
// docker_compose_environment.go), so callers that need to compare the full
// listing/paging/count against a locally tracked device set must pass ids
// to filter out devices left behind by other tests/suites.
func (s *DevauthManagementV2Suite) listDevices(ctx context.Context, status string, page, perPage int32, ids ...string) ([]client.Device, error) {
	req := s.APIClient.DeviceAuthenticationManagementAPIAPI.DeviceAuthManagementListDevices(ctx)
	if status != "" {
		req = req.Status(status)
	}
	if page > 0 {
		req = req.Page(page)
	}
	if perPage > 0 {
		req = req.PerPage(perPage)
	}
	if len(ids) > 0 {
		req = req.Id(ids)
	}
	devices, _, err := req.Execute()
	return devices, err
}

func deviceIDs(devs []*devFixture) []string {
	ids := make([]string, len(devs))
	for i, d := range devs {
		ids[i] = d.ID
	}
	return ids
}

func (s *DevauthManagementV2Suite) setAuthsetStatus(ctx context.Context, deviceID, authsetID, status string) error {
	_, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementSetAuthenticationStatus(ctx, deviceID, authsetID).
		Status(client.Status{Status: status}).
		Execute()
	return err
}

func (s *DevauthManagementV2Suite) getAuthsetStatus(ctx context.Context, did, aid string) (string, *http.Response, error) {
	status, r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementGetAuthenticationStatus(ctx, did, aid).
		Execute()
	if status == nil {
		return "", r, err
	}
	return status.GetStatus(), r, err
}

func (s *DevauthManagementV2Suite) deleteDevice(ctx context.Context, id string) (int, error) {
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementDecommissionDevice(ctx, id).Execute()
	if r == nil {
		return 0, err
	}
	return r.StatusCode, err
}

func (s *DevauthManagementV2Suite) deleteAuthset(ctx context.Context, did, aid string) (int, error) {
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementRemoveAuthentication(ctx, did, aid).Execute()
	if r == nil {
		return 0, err
	}
	return r.StatusCode, err
}

func (s *DevauthManagementV2Suite) waitDeviceGone(ctx context.Context, id string) bool {
	err := common.RetryUntil(ctx, 10*time.Second, 200*time.Millisecond, func() (bool, error) {
		_, r, _ := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, id).Execute()
		return r != nil && r.StatusCode == http.StatusNotFound, nil
	})
	return err == nil
}
