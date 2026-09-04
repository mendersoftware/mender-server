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

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DevauthManagementV2Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
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
	token, _, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotEmpty(token)
	s.JWT = token
}

func (s *DevauthManagementV2Suite) TestPreauth() {
	s.Run("Success/Preauthorize", s.testPreauthOk)
	s.Run("Failure/Duplicate", s.testPreauthFailDuplicate)
	s.Run("Failure/BadRequest", s.testPreauthFailBadRequest)
}

func (s *DevauthManagementV2Suite) testPreauthOk() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	before, _, err := devauthm.DeviceAuthManagementCountDevices(ctx).
		Status(model.DevStatusPreauth).Execute()
	require.NoError(err)

	devs := make([]*common.Device, 0, len(common.KeyKinds))
	for _, kind := range common.KeyKinds {
		kp, err := common.NewKeyPair(kind)
		require.NoError(err)
		mac, err := common.RandomMAC()
		require.NoError(err)
		device := common.NewDeviceFromKeyPair(kp, macIDData(mac.String()))
		device.MAC = mac.String()
		devs = append(devs, device)
	}

	for _, device := range devs {
		r, err := devauthm.DeviceAuthManagementPreauthorize(ctx).
			PreAuthSet(client.PreAuthSet{
				IdentityData: client.IdentityData{Mac: client.PtrString(device.MAC)},
				Pubkey:       common.ExportPublicKeyPEM(device.PublicKey),
			}).Execute()
		require.NoError(err)
		require.Equal(http.StatusCreated, r.StatusCode)
		device.ID = path.Base(r.Header.Get("Location"))
		require.NotEmpty(device.ID)
	}

	after, _, err := devauthm.DeviceAuthManagementCountDevices(ctx).
		Status(model.DevStatusPreauth).Execute()
	require.NoError(err)
	assert.Equal(before.GetCount()+int32(len(devs)), after.GetCount())

	for _, device := range devs {
		dev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, device.ID).Execute()
		require.NoError(err)
		assert.Equal(model.DevStatusPreauth, dev.GetStatus())
		require.Len(dev.AuthSets, 1)
		assert.Equal(device.MAC, fmt.Sprint(dev.AuthSets[0].IdentityData["mac"]))
		assert.Equal(model.DevStatusPreauth, dev.AuthSets[0].GetStatus())

		// the actual device can obtain an auth token
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.True(ok)

		outdev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, device.ID).Execute()
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, outdev.GetStatus())
		require.Len(outdev.AuthSets, 1)
		assert.Equal(model.DevStatusAccepted, outdev.AuthSets[0].GetStatus())
	}

	// re-preauthorize the same devices with fresh keys (force): the new auth
	// set is accepted immediately and the old one rejected
	for i, device := range devs {
		kp, err := common.NewKeyPair(common.KeyKindRSA)
		require.NoError(err)
		fresh := common.NewDeviceFromKeyPair(kp, device.IDData)
		fresh.MAC = device.MAC
		fresh.ID = device.ID
		devs[i] = fresh

		r, err := devauthm.DeviceAuthManagementPreauthorize(ctx).
			PreAuthSet(client.PreAuthSet{
				IdentityData: client.IdentityData{Mac: client.PtrString(fresh.MAC)},
				Pubkey:       common.ExportPublicKeyPEM(fresh.PublicKey),
				Force:        client.PtrBool(true),
			}).Execute()
		require.NoError(err)
		assert.Equal(http.StatusCreated, r.StatusCode)
	}

	for _, device := range devs {
		dev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, device.ID).Execute()
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, dev.GetStatus())
		require.Len(dev.AuthSets, 2)

		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.True(ok)

		outdev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, device.ID).Execute()
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, outdev.GetStatus())
		require.Len(outdev.AuthSets, 2)

		foundAccepted := false
		for _, aset := range outdev.AuthSets {
			if aset.GetStatus() == model.DevStatusAccepted {
				foundAccepted = true
				assert.True(samePubkey(aset.GetPubkey(), common.ExportPublicKeyPEM(device.PublicKey)))
				break
			}
		}
		assert.True(foundAccepted)
	}
}

func (s *DevauthManagementV2Suite) testPreauthFailDuplicate() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	before, _, err := devauthm.DeviceAuthManagementCountDevices(ctx).Execute()
	require.NoError(err)

	mac, err := common.RandomMAC()
	require.NoError(err)
	idData := macIDData(mac.String())
	kp, err := common.NewKeyPair(common.KeyKindRSA)
	require.NoError(err)

	device := common.NewDeviceFromKeyPair(kp, idData)
	ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	assert.False(ok)

	newKP, err := common.NewKeyPair(common.KeyKindRSA)
	require.NoError(err)

	r, err := devauthm.DeviceAuthManagementPreauthorize(ctx).
		PreAuthSet(client.PreAuthSet{
			IdentityData: client.IdentityData{Mac: client.PtrString(mac.String())},
			Pubkey:       newKP.PublicKeyPEM(),
		}).Execute()
	require.Error(err)
	assert.Equal(http.StatusConflict, r.StatusCode)

	// the failed duplicate did not add a device
	after, _, err := devauthm.DeviceAuthManagementCountDevices(ctx).Execute()
	require.NoError(err)
	assert.Equal(before.GetCount()+1, after.GetCount())

	// the existing device still has only its original auth set
	id, err := s.findDeviceIDByMac(ctx, mac.String())
	require.NoError(err)
	apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, id).Execute()
	require.NoError(err)
	require.Len(apiDev.AuthSets, 1)
	assert.True(samePubkey(apiDev.AuthSets[0].GetPubkey(), kp.PublicKeyPEM()))
	assert.Equal(model.DevStatusPending, apiDev.AuthSets[0].GetStatus())
}

func (s *DevauthManagementV2Suite) testPreauthFailBadRequest() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	kp, err := common.NewKeyPair(common.KeyKindRSA)
	require.NoError(err)

	// identity_data as a raw string instead of an object: inject the bad
	// value through AdditionalProperties, which the typed field can't hold
	_, err = devauthm.DeviceAuthManagementPreauthorize(ctx).
		PreAuthSet(client.PreAuthSet{
			AdditionalProperties: map[string]any{"identity_data": `{"mac": "foo"}`},
			Pubkey:               kp.PublicKeyPEM(),
		}).Execute()
	require.Error(err)

	// not a valid key
	mac, err := common.RandomMAC()
	require.NoError(err)
	r, err := devauthm.DeviceAuthManagementPreauthorize(ctx).
		PreAuthSet(client.PreAuthSet{
			IdentityData: client.IdentityData{Mac: client.PtrString(mac.String())},
			Pubkey:       "not a public key",
		}).Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)
}

func (s *DevauthManagementV2Suite) TestDeviceMgmt() {
	s.Run("Success/GetDevices", s.testDeviceMgmtGetDevices)
	s.Run("Success/GetDevice", s.testDeviceMgmtGetDevice)
	s.Run("Success/DeleteDevice", s.testDeviceMgmtDeleteDeviceOk)
	s.Run("Failure/DeleteDeviceNotFound", s.testDeviceMgmtDeleteDeviceNotFound)
	s.Run("Success/DeviceCount", s.testDeviceMgmtDeviceCount)
}

func (s *DevauthManagementV2Suite) testDeviceMgmtGetDevices() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// the environment isn't reset between tests, so restrict the listing to
	// the devices this test created
	ids := make([]string, len(devs))
	for i, d := range devs {
		ids[i] = d.GetId()
	}

	cases := []struct {
		name          string
		status        string
		page, perPage int
	}{
		{"AllStatuses", "", 0, 0},
		{"FilterPending", model.DevStatusPending, 0, 0},
		{"FilterAccepted", model.DevStatusAccepted, 0, 0},
		{"FilterRejected", model.DevStatusRejected, 0, 0},
		{"FilterPreauthorized", model.DevStatusPreauth, 0, 0},
		{"AllPage1Size10", "", 1, 10},
		{"AllPage3Size10", "", 3, 10},
		{"AllPage2Size5", "", 2, 5},
		{"AcceptedPage1Size4", model.DevStatusAccepted, 1, 4},
		{"AcceptedPage2Size4", model.DevStatusAccepted, 2, 4},
		{"AcceptedPage5Size2", model.DevStatusAccepted, 5, 2},
		{"PendingPage2Size2", model.DevStatusPending, 2, 2},
	}

	for _, tc := range cases {
		s.Run("Success/"+tc.name, func() {
			require, assert := s.Require(), s.Assert()
			req := s.APIClient.DeviceAuthenticationManagementAPIAPI.
				DeviceAuthManagementListDevices(ctx).Id(ids)
			if tc.status != "" {
				req = req.Status(tc.status)
			}
			if tc.page > 0 {
				req = req.Page(int32(tc.page))
			}
			if tc.perPage > 0 {
				req = req.PerPage(int32(tc.perPage))
			}
			apiDevs, _, err := req.Execute()
			require.NoError(err)

			if tc.perPage > 0 {
				assert.LessOrEqual(len(apiDevs), tc.perPage)
			}
			for _, d := range apiDevs {
				if tc.status != "" {
					assert.Equal(tc.status, d.GetStatus())
				}
			}
		})
	}
}

func (s *DevauthManagementV2Suite) testDeviceMgmtGetDevice() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, dev := range devs {
		apiDev, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)
		assert.Equal(dev.GetId(), apiDev.GetId())
		assert.Equal(dev.GetStatus(), apiDev.GetStatus())
		require.Len(apiDev.AuthSets, len(dev.AuthSets))
	}

	for _, id := range []string{"foo", "bar"} {
		_, r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, id).Execute()
		require.Error(err)
		assert.Equal(http.StatusNotFound, r.StatusCode)
	}
}

func (s *DevauthManagementV2Suite) testDeviceMgmtDeleteDeviceOk() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, keys, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// decommission a pending device
	var devPending *client.Device
	for i := range devs {
		if devs[i].GetStatus() == model.DevStatusPending {
			devPending = &devs[i]
			break
		}
	}
	require.NotNil(devPending)

	r, err := devauthm.DeviceAuthManagementDecommissionDevice(ctx, devPending.GetId()).Execute()
	require.NoError(err)
	assert.Equal(http.StatusNoContent, r.StatusCode)
	require.True(s.waitDeviceGone(ctx, devPending.GetId()), "timeout waiting for device auth to be deleted")

	// log an accepted device in, then decommission it and check it lost access
	var devAcc *client.Device
	for i := range devs {
		if devs[i].GetStatus() == model.DevStatusAccepted {
			devAcc = &devs[i]
			break
		}
	}
	require.NotNil(devAcc)

	device := keys[devAcc.GetId()][0]
	ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)
	require.True(ok)
	dtoken := device.Token

	r, err = devauthm.DeviceAuthManagementDecommissionDevice(ctx, devAcc.GetId()).Execute()
	require.NoError(err)
	assert.Equal(http.StatusNoContent, r.StatusCode)

	_, r, err = s.APIClient.DeploymentsDeviceAPIAPI.
		CheckUpdate(common.JWTAuthContext(ctx, dtoken)).
		DeviceType("foo").ArtifactName("bar").Execute()
	require.Error(err)
	assert.Equal(http.StatusUnauthorized, r.StatusCode)

	require.True(s.waitDeviceGone(ctx, devAcc.GetId()), "timeout waiting for device auth to be deleted")
}

func (s *DevauthManagementV2Suite) testDeviceMgmtDeleteDeviceNotFound() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	_, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementDecommissionDevice(ctx, "foo").Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)
}

func (s *DevauthManagementV2Suite) testDeviceMgmtDeviceCount() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	statuses := []string{
		"",
		model.DevStatusPending,
		model.DevStatusAccepted,
		model.DevStatusRejected,
		model.DevStatusPreauth,
		model.DevStatusNoAuth,
	}

	// CountDevices has no id filter and the environment isn't reset between
	// tests, so verify counts as deltas against a baseline taken before this
	// test's own devices exist
	countBy := func(status string) int32 {
		req := devauthm.DeviceAuthManagementCountDevices(ctx)
		if status != "" {
			req = req.Status(status)
		}
		count, _, err := req.Execute()
		require.NoError(err)
		return count.GetCount()
	}

	baseline := make(map[string]int32, len(statuses))
	for _, status := range statuses {
		baseline[status] = countBy(status)
	}

	devs, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, status := range statuses {
		var created int32
		for _, d := range devs {
			if status == "" || d.GetStatus() == status {
				created++
			}
		}
		assert.Equal(baseline[status]+created, countBy(status), "status=%q", status)
	}

	_, r, err := devauthm.DeviceAuthManagementCountDevices(ctx).Status("foo").Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)
}

func (s *DevauthManagementV2Suite) TestAuthsetMgmt() {
	s.Run("Success/GetStatus", s.testAuthsetMgmtGetStatus)
	s.Run("Success/PutStatusAccept", s.testAuthsetMgmtPutStatusAccept)
	s.Run("Success/PutStatusReject", s.testAuthsetMgmtPutStatusReject)
	s.Run("Failure/PutStatus", s.testAuthsetMgmtPutStatusFailed)
	s.Run("Success/DeleteStatus", s.testAuthsetMgmtDeleteStatus)
	s.Run("Failure/DeleteStatus", s.testAuthsetMgmtDeleteStatusFailed)
}

func (s *DevauthManagementV2Suite) testAuthsetMgmtGetStatus() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, dev := range devs {
		apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)
		for _, aset := range apiDev.AuthSets {
			status, r, err := devauthm.
				DeviceAuthManagementGetAuthenticationStatus(ctx, dev.GetId(), aset.GetId()).
				Execute()
			require.NoError(err)
			require.Equal(http.StatusOK, r.StatusCode)
			assert.Equal(aset.GetStatus(), status.GetStatus())
		}
	}

	notFoundCases := []struct {
		name     string
		did, aid string
	}{
		{"UnknownAuthSet", devs[0].GetId(), "foo"},
		{"UnknownDevice", "foo", "bar"},
	}
	for _, tc := range notFoundCases {
		s.Run("Failure/"+tc.name, func() {
			require, assert := s.Require(), s.Assert()
			_, r, err := devauthm.
				DeviceAuthManagementGetAuthenticationStatus(ctx, tc.did, tc.aid).Execute()
			require.Error(err)
			assert.Equal(http.StatusNotFound, r.StatusCode)
		})
	}
}

func (s *DevauthManagementV2Suite) testAuthsetMgmtPutStatusAccept() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, keys, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// interesting devices: pending, rejected, or accepted/preauthd with an
	// extra auth set to accept
	var candidates []client.Device
	for _, status := range []string{model.DevStatusPending, model.DevStatusRejected, model.DevStatusAccepted, model.DevStatusPreauth} {
		for _, d := range devs {
			if d.GetStatus() != status {
				continue
			}
			// for accepted/preauthd devices only those with an extra auth set
			if (status == model.DevStatusAccepted || status == model.DevStatusPreauth) && len(d.AuthSets) <= 1 {
				continue
			}
			candidates = append(candidates, d)
		}
	}

	for _, dev := range candidates {
		apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)

		// for an accepted device, first get a live token for the accepted set
		var dtoken string
		if dev.GetStatus() == model.DevStatusAccepted {
			accepted := authSetByStatus(apiDev.AuthSets, model.DevStatusAccepted)
			require.NotNil(accepted)
			device := authSetKey(keys[dev.GetId()], *accepted)
			ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
			require.NoError(err)
			require.True(ok)
			dtoken = device.Token
		}

		// accept a pending or rejected auth set
		var target *client.AuthSet
		for i := range apiDev.AuthSets {
			if st := apiDev.AuthSets[i].GetStatus(); st == model.DevStatusPending || st == model.DevStatusRejected {
				target = &apiDev.AuthSets[i]
				break
			}
		}
		require.NotNil(target)

		_, err = devauthm.
			DeviceAuthManagementSetAuthenticationStatus(ctx, dev.GetId(), target.GetId()).
			Status(client.Status{Status: model.DevStatusAccepted}).Execute()
		require.NoError(err)

		s.waitInventoryStatus(ctx, dev.GetId(), model.DevStatusAccepted)

		updated, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)
		assert.Equal(model.DevStatusAccepted, updated.GetStatus())
		assert.Equal(model.DevStatusAccepted, authSetByID(updated.AuthSets, target.GetId()).GetStatus())

		if dtoken != "" {
			// the previously active device lost access
			_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
				CheckUpdate(common.JWTAuthContext(ctx, dtoken)).
				DeviceType("foo").ArtifactName("bar").Execute()
			require.Error(err)
			assert.Equal(http.StatusUnauthorized, r.StatusCode)
		}
	}
}

func (s *DevauthManagementV2Suite) testAuthsetMgmtPutStatusReject() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, keys, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	var candidates []client.Device
	for _, status := range []string{model.DevStatusPending, model.DevStatusAccepted} {
		for _, d := range devs {
			if d.GetStatus() == status {
				candidates = append(candidates, d)
			}
		}
	}

	for _, dev := range candidates {
		apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)

		var target *client.AuthSet
		if dev.GetStatus() == model.DevStatusAccepted {
			target = authSetByStatus(apiDev.AuthSets, model.DevStatusAccepted)
		} else {
			target = &apiDev.AuthSets[0]
		}
		require.NotNil(target)

		// for an accepted device, also hold a live token and check it loses access
		var dtoken string
		if dev.GetStatus() == model.DevStatusAccepted {
			device := authSetKey(keys[dev.GetId()], *target)
			ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
			require.NoError(err)
			require.True(ok)
			dtoken = device.Token
		}

		if st := target.GetStatus(); st == model.DevStatusAccepted || st == model.DevStatusPending {
			_, err = devauthm.
				DeviceAuthManagementSetAuthenticationStatus(ctx, dev.GetId(), target.GetId()).
				Status(client.Status{Status: model.DevStatusRejected}).Execute()
			require.NoError(err)
		}

		targetDev := authSetKey(keys[dev.GetId()], *target)
		updated, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)
		assert.Equal(model.DevStatusRejected, authSetByID(updated.AuthSets, target.GetId()).GetStatus())

		if dtoken != "" {
			_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
				CheckUpdate(common.JWTAuthContext(ctx, dtoken)).
				DeviceType("foo").ArtifactName("bar").Execute()
			require.Error(err)
			assert.Equal(http.StatusUnauthorized, r.StatusCode)
		}

		// a fresh auth request from the rejected key stays rejected and adds
		// no new auth set
		ok, err := targetDev.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.False(ok)

		after, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)
		assert.Len(after.AuthSets, len(updated.AuthSets))
	}
}

func (s *DevauthManagementV2Suite) testAuthsetMgmtPutStatusFailed() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, devs[0].GetId()).Execute()
	require.NoError(err)
	asetID := apiDev.AuthSets[0].GetId()

	// not found: valid device, bogus auth set
	r, err := devauthm.
		DeviceAuthManagementSetAuthenticationStatus(ctx, devs[0].GetId(), "foo").
		Status(client.Status{Status: model.DevStatusAccepted}).Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// not found: bogus device
	r, err = devauthm.
		DeviceAuthManagementSetAuthenticationStatus(ctx, "foo", "bar").
		Status(client.Status{Status: model.DevStatusAccepted}).Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// not found: bogus device, status rejected
	r, err = devauthm.
		DeviceAuthManagementSetAuthenticationStatus(ctx, "foo", "bar").
		Status(client.Status{Status: model.DevStatusRejected}).Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// bad request: invalid status
	r, err = devauthm.
		DeviceAuthManagementSetAuthenticationStatus(ctx, devs[0].GetId(), asetID).
		Status(client.Status{Status: "invalid"}).Execute()
	require.Error(err)
	assert.Equal(http.StatusBadRequest, r.StatusCode)
}

func (s *DevauthManagementV2Suite) testAuthsetMgmtDeleteStatus() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, keys, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	for _, dev := range devs {
		apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)

		// for accepted/preauthd devices target that auth set; otherwise the first
		var target *client.AuthSet
		if dev.GetStatus() == model.DevStatusAccepted || dev.GetStatus() == model.DevStatusPreauth {
			target = authSetByStatus(apiDev.AuthSets, dev.GetStatus())
			require.NotNil(target)
		} else {
			target = &apiDev.AuthSets[0]
		}

		var dtoken string
		if dev.GetStatus() == model.DevStatusAccepted {
			device := authSetKey(keys[dev.GetId()], *target)
			ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
			require.NoError(err)
			require.True(ok)
			dtoken = device.Token
		}

		r, err := devauthm.
			DeviceAuthManagementRemoveAuthentication(ctx, dev.GetId(), target.GetId()).Execute()
		require.NoError(err)
		require.Equal(http.StatusNoContent, r.StatusCode)

		if dev.GetStatus() == model.DevStatusPreauth {
			// removing the preauth auth set removes the device entirely.
			// Deviceauth does this asynchronously, so poll for the 404 (a
			// specific 404, not any error) rather than asserting it
			// immediately. The budget is generous because the cascade runs
			// slower against the shared environment under CI load.
			goneErr := common.RetryUntil(ctx, 3*time.Minute, time.Second, func() (bool, error) {
				_, r, _ := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
				return r != nil && r.StatusCode == http.StatusNotFound, nil
			})
			assert.NoError(goneErr, "device %s still present in device auth after preauth authset removal", dev.GetId())
			goneErr = common.RetryUntil(ctx, 3*time.Minute, time.Second, func() (bool, error) {
				_, r, _ := s.APIClient.DeviceInventoryManagementAPIAPI.
					GetDeviceInventory(ctx, dev.GetId()).Execute()
				return r != nil && r.StatusCode == http.StatusNotFound, nil
			})
			assert.NoError(goneErr, "device %s still present in inventory after preauth authset removal", dev.GetId())
			// devices are listed sorted by status ascending (accepted,
			// pending, preauthorized, rejected), so the first preauthorized
			// device is the last one this test needs to exercise
			return
		}

		// the device survives with its remaining auth sets
		after, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, dev.GetId()).Execute()
		require.NoError(err)
		assert.Len(after.AuthSets, len(apiDev.AuthSets)-1)

		if dtoken != "" {
			_, r, err := s.APIClient.DeploymentsDeviceAPIAPI.
				CheckUpdate(common.JWTAuthContext(ctx, dtoken)).
				DeviceType("foo").ArtifactName("bar").Execute()
			require.Error(err)
			assert.Equal(http.StatusUnauthorized, r.StatusCode)
		}
	}
}

func (s *DevauthManagementV2Suite) testAuthsetMgmtDeleteStatusFailed() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs, _, err := s.makeDevsWithAuthsets(ctx)
	require.NoError(err)

	// not found: valid device, bogus auth set
	r, err := devauthm.
		DeviceAuthManagementRemoveAuthentication(ctx, devs[0].GetId(), "foo").Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)

	// not found: bogus device
	r, err = devauthm.
		DeviceAuthManagementRemoveAuthentication(ctx, "foo", "bar").Execute()
	require.Error(err)
	assert.Equal(http.StatusNotFound, r.StatusCode)
}

func (s *DevauthManagementV2Suite) TestAuthReqSubmitAccept() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devauthm := s.APIClient.DeviceAuthenticationManagementAPIAPI

	devs := make([]*common.Device, 0, len(common.KeyKinds))
	for _, kind := range common.KeyKinds {
		kp, err := common.NewKeyPair(kind)
		require.NoError(err)
		mac, err := common.RandomMAC()
		require.NoError(err)
		device := common.NewDeviceFromKeyPair(kp, macIDData(mac.String()))
		device.MAC = mac.String()
		devs = append(devs, device)
	}

	for _, device := range devs {
		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		assert.False(ok)
	}

	for _, device := range devs {
		id, err := s.findDeviceIDByMac(ctx, device.MAC)
		require.NoError(err)
		apiDev, _, err := devauthm.DeviceAuthManagementGetDevice(ctx, id).Execute()
		require.NoError(err)
		require.Len(apiDev.AuthSets, 1)
		assert.Equal(model.DevStatusPending, apiDev.AuthSets[0].GetStatus())

		_, err = devauthm.
			DeviceAuthManagementSetAuthenticationStatus(ctx, apiDev.GetId(), apiDev.AuthSets[0].GetId()).
			Status(client.Status{Status: model.DevStatusAccepted}).Execute()
		require.NoError(err)

		ok, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		require.True(ok)

		parts := strings.Split(device.Token, ".")
		require.Len(parts, 3)

		payloadRaw, err := base64.RawStdEncoding.DecodeString(parts[1])
		require.NoError(err)
		var payload map[string]any
		require.NoError(json.Unmarshal(payloadRaw, &payload))
		assert.Equal(apiDev.GetId(), payload["sub"])
		assert.Equal("Mender", payload["iss"])
		assert.NotEmpty(payload["jti"])
		assert.NotNil(payload["exp"])
		assert.Equal(true, payload["mender.device"])

		headerRaw, err := base64.RawStdEncoding.DecodeString(parts[0])
		require.NoError(err)
		var header map[string]any
		require.NoError(json.Unmarshal(headerRaw, &header))
		assert.Equal("JWT", header["typ"])
	}
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

// authSetKey returns the key (as a common.Device) that produced the given API
// auth set, matched by public key. The API never returns private keys, so the
// per-auth-set common.Device is kept to re-submit auth requests.
func authSetKey(keys []*common.Device, aset client.AuthSet) *common.Device {
	for _, k := range keys {
		if aset.Pubkey != nil && samePubkey(*aset.Pubkey, common.ExportPublicKeyPEM(k.PublicKey)) {
			return k
		}
	}
	return nil
}

func macIDData(mac string) string {
	b, _ := json.Marshal(map[string]string{"mac": mac})
	return string(b)
}

func samePubkey(a, b string) bool {
	da, _ := pem.Decode([]byte(a))
	db, _ := pem.Decode([]byte(b))
	if da == nil || db == nil {
		return false
	}
	return bytes.Equal(da.Bytes, db.Bytes)
}

func authSetByID(asets []client.AuthSet, id string) *client.AuthSet {
	for i := range asets {
		if asets[i].GetId() == id {
			return &asets[i]
		}
	}
	return nil
}

func authSetByStatus(asets []client.AuthSet, status string) *client.AuthSet {
	for i := range asets {
		if asets[i].GetStatus() == status {
			return &asets[i]
		}
	}
	return nil
}

func authSetByPubkey(asets []client.AuthSet, pubkeyPEM string) *client.AuthSet {
	for i := range asets {
		if asets[i].Pubkey != nil && samePubkey(*asets[i].Pubkey, pubkeyPEM) {
			return &asets[i]
		}
	}
	return nil
}

// findDeviceIDByMac resolves the device id for the given identity mac via an
// indexed inventory search by mac (deviceauth's own listing has no mac filter,
// so paging it is O(number-of-devices) against the shared, never-reset
// environment). A pending device's id isn't set on its common.Device (only
// Accept sets it), so this is how the fixture learns it. The device id is the
// same in deviceauth and inventory.
func (s *DevauthManagementV2Suite) findDeviceIDByMac(ctx context.Context, mac string) (string, error) {
	var id string
	err := common.RetryUntil(ctx, 15*time.Second, 500*time.Millisecond, func() (bool, error) {
		devices, _, err := s.APIClient.DeviceInventoryFiltersAndSearchManagementAPIAPI.
			InventoryV2SearchDeviceInventories(ctx).
			SearchParams(client.SearchParams{
				Filters: []client.FilterPredicate{{
					Scope:     client.IDENTITY,
					Attribute: "mac",
					Type:      "$eq",
					Value:     client.AttributeValueRequest{String: client.PtrString(mac)},
				}},
			}).Execute()
		if err != nil {
			return false, err
		}
		if len(devices) > 0 {
			id = devices[0].GetId()
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return "", fmt.Errorf("device with mac %s not found: %w", mac, err)
	}
	return id, nil
}

func (s *DevauthManagementV2Suite) waitDeviceGone(ctx context.Context, id string) bool {
	err := common.RetryUntil(ctx, 10*time.Second, 200*time.Millisecond, func() (bool, error) {
		_, r, _ := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, id).Execute()
		return r != nil && r.StatusCode == http.StatusNotFound, nil
	})
	return err == nil
}

func (s *DevauthManagementV2Suite) waitInventoryStatus(ctx context.Context, id, status string) {
	err := common.RetryUntil(ctx, 10*time.Second, 250*time.Millisecond, func() (bool, error) {
		inv, _, err := s.APIClient.DeviceInventoryManagementAPIAPI.
			GetDeviceInventory(ctx, id).Execute()
		if err != nil || inv == nil {
			return false, nil
		}
		for _, a := range inv.Attributes {
			if a.GetScope() == client.IDENTITY && a.GetName() == "status" {
				return a.GetValue().String != nil && *a.GetValue().String == status, nil
			}
		}
		return false, nil
	})
	require.NoError(s.T(), err, "timeout waiting for inventory status %q for device %s", status, id)
}

// createPending onboards a device with numAuthsets pending auth sets (each with
// a fresh key of the given kind) sharing one identity.
func (s *DevauthManagementV2Suite) createPending(ctx context.Context, kind string, numAuthsets int) ([]*common.Device, error) {
	mac, err := common.RandomMAC()
	if err != nil {
		return nil, err
	}
	idData := macIDData(mac.String())
	var authSets []*common.Device
	for i := 0; i < numAuthsets; i++ {
		kp, err := common.NewKeyPair(kind)
		if err != nil {
			return nil, err
		}
		device := common.NewDeviceFromKeyPair(kp, idData)
		device.MAC = mac.String()
		if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
			return nil, err
		}
		authSets = append(authSets, device)
	}
	id, err := s.findDeviceIDByMac(ctx, mac.String())
	if err != nil {
		return nil, err
	}
	for _, d := range authSets {
		d.ID = id
	}
	return authSets, nil
}

// createAccepted onboards a pending device and accepts its first numAccepted
// auth sets.
func (s *DevauthManagementV2Suite) createAccepted(ctx context.Context, kind string, numAuthsets, numAccepted int) ([]*common.Device, error) {
	authSets, err := s.createPending(ctx, kind, numAuthsets)
	if err != nil {
		return nil, err
	}
	id := authSets[0].ID
	apiDev, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementGetDevice(ctx, id).Execute()
	if err != nil {
		return nil, err
	}
	for i := 0; i < numAccepted; i++ {
		aset := authSetByPubkey(apiDev.AuthSets, common.ExportPublicKeyPEM(authSets[i].PublicKey))
		if aset == nil {
			return nil, fmt.Errorf("auth set for accepted key %d not found", i)
		}
		if _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementSetAuthenticationStatus(ctx, id, aset.GetId()).
			Status(client.Status{Status: model.DevStatusAccepted}).Execute(); err != nil {
			return nil, err
		}
	}
	return authSets, nil
}

// createRejected onboards a pending device and rejects all its auth sets.
func (s *DevauthManagementV2Suite) createRejected(ctx context.Context, kind string, numAuthsets int) ([]*common.Device, error) {
	authSets, err := s.createPending(ctx, kind, numAuthsets)
	if err != nil {
		return nil, err
	}
	id := authSets[0].ID
	apiDev, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementGetDevice(ctx, id).Execute()
	if err != nil {
		return nil, err
	}
	for _, aset := range apiDev.AuthSets {
		if _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementSetAuthenticationStatus(ctx, id, aset.GetId()).
			Status(client.Status{Status: model.DevStatusRejected}).Execute(); err != nil {
			return nil, err
		}
	}
	return authSets, nil
}

// createPreauthd preauthorizes a device (single preauthorized auth set).
func (s *DevauthManagementV2Suite) createPreauthd(ctx context.Context, kind string) ([]*common.Device, error) {
	kp, err := common.NewKeyPair(kind)
	if err != nil {
		return nil, err
	}
	mac, err := common.RandomMAC()
	if err != nil {
		return nil, err
	}
	idData := macIDData(mac.String())
	r, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
		DeviceAuthManagementPreauthorize(ctx).
		PreAuthSet(client.PreAuthSet{
			IdentityData: client.IdentityData{Mac: client.PtrString(mac.String())},
			Pubkey:       kp.PublicKeyPEM(),
		}).Execute()
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected preauthorize status %d", r.StatusCode)
	}
	id := path.Base(r.Header.Get("Location"))
	if id == "" {
		return nil, fmt.Errorf("preauthorize: missing Location header")
	}
	device := common.NewDeviceFromKeyPair(kp, idData)
	device.MAC = mac.String()
	device.ID = id
	return []*common.Device{device}, nil
}

// createPreauthdWithPending preauthorizes a device and adds numPending extra
// pending auth sets; the device stays preauthorized.
func (s *DevauthManagementV2Suite) createPreauthdWithPending(ctx context.Context, kind string, numPending int) ([]*common.Device, error) {
	authSets, err := s.createPreauthd(ctx, kind)
	if err != nil {
		return nil, err
	}
	idData := authSets[0].IDData
	id, mac := authSets[0].ID, authSets[0].MAC
	for i := 0; i < numPending; i++ {
		kp, err := common.NewKeyPair(common.KeyKindRSA)
		if err != nil {
			return nil, err
		}
		device := common.NewDeviceFromKeyPair(kp, idData)
		device.MAC = mac
		device.ID = id
		if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
			return nil, err
		}
		authSets = append(authSets, device)
	}
	return authSets, nil
}

// makeDevsWithAuthsets returns the created devices as their live client.Device
// state (sorted the way the management listing returns them, by status then id)
// plus a map from device id to the per-auth-set keys (common.Device) used to
// create it, which the API never returns.
func (s *DevauthManagementV2Suite) makeDevsWithAuthsets(ctx context.Context) ([]client.Device, map[string][]*common.Device, error) {
	var groups [][]*common.Device
	add := func(g []*common.Device, err error) error {
		if err != nil {
			return err
		}
		groups = append(groups, g)
		return nil
	}

	// pending devices, single auth set
	for i := 0; i < 3; i++ {
		if err := add(s.createPending(ctx, common.KeyKindRSA, 1)); err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.createPending(ctx, common.KeyKindECP256, 1)); err != nil {
			return nil, nil, err
		}
	}
	if err := add(s.createPending(ctx, common.KeyKindEd25519, 1)); err != nil {
		return nil, nil, err
	}

	// pending devices with more than one auth set
	for i := 0; i < 2; i++ {
		if err := add(s.createPending(ctx, common.KeyKindRSA, 3)); err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.createPending(ctx, common.KeyKindECP256, 3)); err != nil {
			return nil, nil, err
		}
	}
	if err := add(s.createPending(ctx, common.KeyKindEd25519, 3)); err != nil {
		return nil, nil, err
	}

	// accepted devices, single auth set
	for i := 0; i < 3; i++ {
		if err := add(s.createAccepted(ctx, common.KeyKindRSA, 1, 1)); err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.createAccepted(ctx, common.KeyKindECP256, 1, 1)); err != nil {
			return nil, nil, err
		}
	}
	if err := add(s.createAccepted(ctx, common.KeyKindEd25519, 1, 1)); err != nil {
		return nil, nil, err
	}

	// accepted devices with more than one auth set
	for i := 0; i < 2; i++ {
		if err := add(s.createAccepted(ctx, common.KeyKindRSA, 3, 1)); err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.createAccepted(ctx, common.KeyKindECP256, 2, 1)); err != nil {
			return nil, nil, err
		}
	}
	if err := add(s.createAccepted(ctx, common.KeyKindEd25519, 2, 1)); err != nil {
		return nil, nil, err
	}

	// rejected devices
	for i := 0; i < 2; i++ {
		if err := add(s.createRejected(ctx, common.KeyKindRSA, 3)); err != nil {
			return nil, nil, err
		}
	}
	for i := 0; i < 2; i++ {
		if err := add(s.createRejected(ctx, common.KeyKindECP256, 2)); err != nil {
			return nil, nil, err
		}
	}
	if err := add(s.createRejected(ctx, common.KeyKindEd25519, 2)); err != nil {
		return nil, nil, err
	}

	// preauthorized devices
	if err := add(s.createPreauthd(ctx, common.KeyKindRSA)); err != nil {
		return nil, nil, err
	}
	if err := add(s.createPreauthd(ctx, common.KeyKindECP256)); err != nil {
		return nil, nil, err
	}
	if err := add(s.createPreauthd(ctx, common.KeyKindEd25519)); err != nil {
		return nil, nil, err
	}

	// preauthorized devices with extra pending auth sets
	for i := 0; i < 2; i++ {
		if err := add(s.createPreauthdWithPending(ctx, common.KeyKindRSA, 2)); err != nil {
			return nil, nil, err
		}
	}
	if err := add(s.createPreauthdWithPending(ctx, common.KeyKindECP256, 2)); err != nil {
		return nil, nil, err
	}
	if err := add(s.createPreauthdWithPending(ctx, common.KeyKindEd25519, 2)); err != nil {
		return nil, nil, err
	}

	keys := make(map[string][]*common.Device, len(groups))
	devs := make([]client.Device, 0, len(groups))
	for _, g := range groups {
		id := g[0].ID
		apiDev, _, err := s.APIClient.DeviceAuthenticationManagementAPIAPI.
			DeviceAuthManagementGetDevice(ctx, id).Execute()
		if err != nil {
			return nil, nil, err
		}
		devs = append(devs, *apiDev)
		keys[id] = g
	}
	slices.SortStableFunc(devs, func(a, b client.Device) int {
		if c := cmp.Compare(a.GetStatus(), b.GetStatus()); c != 0 {
			return c
		}
		return cmp.Compare(a.GetId(), b.GetId())
	})

	return devs, keys, nil
}
