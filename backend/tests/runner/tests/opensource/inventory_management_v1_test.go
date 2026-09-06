//nolint:all // This is all test code
package opensource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

// InventoryManagementV1Suite covers the open-source inventory management API.
// Enterprise-only tenant behavior and the internal (tenant-scoped) search
// variant, which behaves identically to the public search when no tenant is
// given, are out of scope.
type InventoryManagementV1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
}

func (i *BackendIntegrationSuite) TestInventoryManagementV1() {
	suite.Run(i.T(), &InventoryManagementV1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (s *InventoryManagementV1Suite) SetupSuite() {
	require := require.New(s.T())

	ctx := common.BasicAuthContext(s.T().Context(), s.User)
	token, _, err := s.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()
	require.NoError(err)
	require.NotEmpty(token)
	s.JWT = token
}

func (s *InventoryManagementV1Suite) TestGetDevicesOk() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	count, err := s.totalDeviceCount(ctx)
	require.NoError(err)

	_, err = s.createAcceptedDevices(ctx, 40)
	require.NoError(err)

	newCount, err := s.totalDeviceCount(ctx)
	require.NoError(err)
	assert.Equal(count+40, newCount)
}

func (s *InventoryManagementV1Suite) TestFilterDevicesOk() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	count, err := s.totalDeviceCount(ctx)
	require.NoError(err)

	devs, err := s.createAcceptedDevices(ctx, 40)
	require.NoError(err)

	newCount, err := s.totalDeviceCount(ctx)
	require.NoError(err)
	assert.Equal(count+40, newCount)

	// upload inventory attributes
	macs := make([]string, len(devs))
	for i, d := range devs {
		mac := "de:ad:be:ef:06:" + fmt.Sprint(i)
		macs[i] = mac
		payload := []client.DeviceAttributeRequest{
			{Name: "mac", Value: client.StringAsAttributeValueRequest(client.PtrString(mac))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
	}

	// get device with exact mac value
	apiDevs, err := s.listDevicesRaw(ctx, url.Values{
		"per_page": {"100"},
		"mac":      {"de:ad:be:ef:06:7"},
	})
	require.NoError(err)
	assert.Len(apiDevs, 1)

	// give one device an int-typed attribute and filter on its exact value
	numAttr := float32(31337)
	r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
		AssignAttributes(common.JWTAuthContext(ctx, devs[7].Token)).
		DeviceAttributeRequest([]client.DeviceAttributeRequest{
			{Name: "test-num-attr", Value: client.Float32AsAttributeValueRequest(&numAttr)},
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)

	apiDevs, err = s.listDevicesRaw(ctx, url.Values{
		"per_page":      {"100"},
		"test-num-attr": {"31337"},
	})
	require.NoError(err)
	assert.Len(apiDevs, 1)
}

func (s *InventoryManagementV1Suite) TestDevicePatchAttributes() {
	s.Run("Success/PatchAttributes", s.testDevicePatchAttributesOk)
	s.Run("Failure/NoAttrValue", s.testDevicePatchAttributesFailNoAttrValue)
}

func (s *InventoryManagementV1Suite) testDevicePatchAttributesOk() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs := make([]*common.Device, 3)
	for i := range devs {
		d, err := s.newAcceptedDeviceWithSN(ctx)
		require.NoError(err)
		devs[i] = d
	}

	for _, d := range devs {
		payload := []client.DeviceAttributeRequest{
			{Name: "mac", Value: client.StringAsAttributeValueRequest(client.PtrString("mac-new-" + d.ID))},
			// empty value for existing
			{Name: "sn", Value: client.StringAsAttributeValueRequest(client.PtrString(""))},
			// empty value for new
			{Name: "new-empty", Value: client.StringAsAttributeValueRequest(client.PtrString(""))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)
	}

	for _, d := range devs {
		apiDev, _, err := s.APIClient.DeviceInventoryManagementAPIAPI.
			GetDeviceInventory(ctx, d.ID).Execute()
		require.NoError(err)

		// Expected inventory count per scope:
		// {"inventory": 3, "identity": 1(status)+2(mac,sn from id_data), "system": 3}
		require.Len(apiDev.Attributes, 9)

		for _, a := range apiDev.Attributes {
			value := a.GetValue()
			str := ""
			if value.String != nil {
				str = *value.String
			}
			switch {
			case a.GetName() == "mac" && a.GetScope() == client.INVENTORY:
				assert.Equal("mac-new-"+apiDev.GetId(), str)
			case a.GetName() == "sn" && a.GetScope() == client.INVENTORY:
				assert.Equal("", str)
			case a.GetName() == "new-empty" && a.GetScope() == client.INVENTORY:
				assert.Equal("", str)
			case a.GetName() == "status" && a.GetScope() == client.IDENTITY:
				assert.Contains([]string{"accepted", "pending"}, str)
			case a.GetScope() != client.INVENTORY:
				assert.NotEqual("", str, "unexpected empty value for %s", a.GetName())
			default:
				assert.Fail("unexpected attribute " + a.GetName())
			}
		}
	}
}

func (s *InventoryManagementV1Suite) testDevicePatchAttributesFailNoAttrValue() {
	require := require.New(s.T())
	assert := assert.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)

	// the generated client requires a Value on DeviceAttributeRequest, so a
	// payload that omits the "value" key entirely has to go over a raw request
	body, err := json.Marshal([]map[string]string{{"name": "mac"}})
	require.NoError(err)
	resp, err := common.RawRequest(common.JWTAuthContext(ctx, d.Token), s.APIClient, http.MethodPatch,
		"/api/devices/v1/inventory/device/attributes", body)
	require.NoError(err)
	defer resp.Body.Close()
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *InventoryManagementV1Suite) TestDeviceFilteringSearchV2() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	type fixtureDev struct {
		id       string
		idx      float32
		artifact []string
	}

	devs := make([]fixtureDev, 3)
	ids := make([]string, 3)
	for i := range devs {
		d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)

		payload := []client.DeviceAttributeRequest{
			{Name: "artifact", Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"v1"})},
			{Name: "idx", Value: client.Float32AsAttributeValueRequest(client.PtrFloat32(float32(i)))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(err)
		require.Equal(http.StatusOK, r.StatusCode)

		devs[i] = fixtureDev{id: d.ID, idx: float32(i), artifact: []string{"v1"}}
		ids[i] = d.ID
	}

	type expectedAttr struct {
		name  string
		scope client.Scope
		value client.AttributeValueRequest
	}
	invAttrs := func(dev fixtureDev) []expectedAttr {
		return []expectedAttr{
			{name: "idx", scope: client.INVENTORY, value: client.Float32AsAttributeValueRequest(client.PtrFloat32(dev.idx))},
			{name: "artifact", scope: client.INVENTORY, value: client.ArrayOfStringAsAttributeValueRequest(&dev.artifact)},
		}
	}

	type expectedDev struct {
		id    string
		attrs []expectedAttr
	}

	// devices sorted by idx descending (all have artifact=["v1"], so none
	// are excluded by the $nin below)
	byIdxDesc := make([]fixtureDev, len(devs))
	copy(byIdxDesc, devs)
	sort.SliceStable(byIdxDesc, func(i, j int) bool { return byIdxDesc[i].idx > byIdxDesc[j].idx })

	cases := []struct {
		name       string
		params     client.SearchParams
		wantStatus int
		expected   []expectedDev // only checked when wantStatus == 200
	}{
		{
			name: "Success/EqSingleMatch",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(client.PtrFloat32(float32(1)))},
				},
			},
			wantStatus: http.StatusOK,
			expected:   []expectedDev{{id: devs[1].id, attrs: invAttrs(devs[1])}},
		},
		{
			name: "Success/EqNoMatch",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "id_data", Scope: client.INVENTORY, Value: client.StringAsAttributeValueRequest(client.PtrString("illegal_data"))},
				},
			},
			wantStatus: http.StatusOK,
			expected:   []expectedDev{},
		},
		{
			name: "Success/NinSortByDescendingIdx",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$nin", Attribute: "artifact", Scope: client.INVENTORY, Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"v3"})},
				},
				Sort: []client.SortCriteria{
					{Attribute: "idx", Scope: client.INVENTORY, Order: "desc"},
				},
			},
			wantStatus: http.StatusOK,
			expected: []expectedDev{
				{id: byIdxDesc[0].id, attrs: invAttrs(byIdxDesc[0])},
				{id: byIdxDesc[1].id, attrs: invAttrs(byIdxDesc[1])},
				{id: byIdxDesc[2].id, attrs: invAttrs(byIdxDesc[2])},
			},
		},
		{
			name: "Failure/MissingTypeParameter",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Attribute: "artifact", Scope: client.INVENTORY, Value: client.StringAsAttributeValueRequest(client.PtrString("v1"))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/InvalidFilterScope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.Scope("user_defined"), Value: client.Float32AsAttributeValueRequest(client.PtrFloat32(float32(1)))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/InvalidSortScope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(client.PtrFloat32(float32(1)))},
				},
				Sort: []client.SortCriteria{
					{Attribute: "idx", Scope: client.Scope("user_defined"), Order: "desc"},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/InvalidAttributeScope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(client.PtrFloat32(float32(1)))},
				},
				Attributes: []client.SelectAttribute{
					{Attribute: "idx", Scope: client.Scope("user_defined")},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/InvalidFilterScopeWithPath",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.Scope("../../../../Windows/system.ini"), Value: client.Float32AsAttributeValueRequest(client.PtrFloat32(float32(1)))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/UnsupportedOperation",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$type", Attribute: "artifact", Scope: client.INVENTORY, Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"int", "string", "array"})},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			results, r, err := s.APIClient.DeviceInventoryFiltersAndSearchManagementAPIAPI.
				InventoryV2SearchDeviceInventories(ctx).
				SearchParams(tc.params).
				Execute()
			if tc.wantStatus == http.StatusOK {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			s.Require().Equal(tc.wantStatus, r.StatusCode)

			if tc.wantStatus != http.StatusOK {
				return
			}

			s.Require().Len(results, len(tc.expected), "unexpected number of results")
			for i, exp := range tc.expected {
				s.Assert().Equal(exp.id, results[i].GetId(), "unexpected device in response")
				for _, wantAttr := range exp.attrs {
					s.Assert().True(
						hasAttribute(results[i].Attributes, wantAttr.name, wantAttr.scope, wantAttr.value),
						"missing inventory attribute %s on device %s", wantAttr.name, exp.id,
					)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

// createAcceptedDevices creates n accepted devices. The callers here onboard
// 40 devices each, and a single onboarding involves several sequential API
// calls plus polling for the async deviceauth->inventory provisioning, so
// doing them serially would add minutes to the suite; run them concurrently
// with a bounded worker pool instead.
func (s *InventoryManagementV1Suite) createAcceptedDevices(ctx context.Context, n int) ([]*common.Device, error) {
	devices := make([]*common.Device, n)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	for i := range n {
		eg.Go(func() error {
			d, err := common.NewAcceptedDevice(egCtx, s.APIClient, s.Tenant.TenantToken)
			if err != nil {
				return err
			}
			devices[i] = d
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return devices, nil
}

// totalDeviceCount reads the X-Total-Count header from a minimal (per_page=1)
// listing request.
func (s *InventoryManagementV1Suite) totalDeviceCount(ctx context.Context) (int, error) {
	_, r, err := s.APIClient.DeviceInventoryManagementAPIAPI.
		ListDeviceInventories(ctx).PerPage(1).Execute()
	if err != nil {
		return 0, err
	}
	var count int
	_, err = fmt.Sscanf(r.Header.Get("X-Total-Count"), "%d", &count)
	return count, err
}

// newAcceptedDeviceWithSN creates and accepts a device whose identity data
// carries both "mac" and "sn". TestDevicePatchAttributes patches an existing
// "sn" attribute to empty and asserts the exact identity-scope attribute count
// (status + mac + sn), so the identity data must include sn.
func (s *InventoryManagementV1Suite) newAcceptedDeviceWithSN(ctx context.Context) (*common.Device, error) {
	kp, err := common.NewKeyPair(common.KeyKindRSA)
	if err != nil {
		return nil, err
	}
	mac := uuid.NewString()
	sn := uuid.NewString()
	idData, err := json.Marshal(map[string]string{"mac": mac, "sn": sn})
	if err != nil {
		return nil, err
	}

	device := common.NewDeviceFromKeyPair(kp, string(idData))
	device.MAC = mac
	if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
		return nil, err
	}
	if err := device.Accept(ctx, s.APIClient); err != nil {
		return nil, err
	}
	// the first SubmitAuthRequest above ran before the device was accepted, so
	// it got a 401 and no token -- submit again now.
	if _, err := device.SubmitAuthRequest(ctx, s.APIClient, s.Tenant.TenantToken); err != nil {
		return nil, err
	}
	return device, nil
}

func hasAttribute(attrs []client.AttributeResponse, name string, scope client.Scope, value client.AttributeValueRequest) bool {
	for _, a := range attrs {
		if a.GetName() != name || a.GetScope() != scope {
			continue
		}
		v := a.GetValue()
		switch {
		case value.String != nil:
			return v.String != nil && *v.String == *value.String
		case value.Float32 != nil:
			return v.Float32 != nil && *v.Float32 == *value.Float32
		case value.ArrayOfString != nil:
			if v.ArrayOfString == nil || len(*v.ArrayOfString) != len(*value.ArrayOfString) {
				return false
			}
			for i, s := range *value.ArrayOfString {
				if (*v.ArrayOfString)[i] != s {
					return false
				}
			}
			return true
		case value.ArrayOfFloat32 != nil:
			if v.ArrayOfFloat32 == nil || len(*v.ArrayOfFloat32) != len(*value.ArrayOfFloat32) {
				return false
			}
			for i, f := range *value.ArrayOfFloat32 {
				if (*v.ArrayOfFloat32)[i] != f {
					return false
				}
			}
			return true
		}
	}
	return false
}

// listDevicesRaw performs a GET against the management v1 inventory listing
// with arbitrary query parameters (e.g. `mac=<value>` attribute filters),
// which the generated client can't express since it only exposes
// page/per_page/sort/group/has_group as typed parameters.
func (s *InventoryManagementV1Suite) listDevicesRaw(ctx context.Context, query url.Values) ([]client.DeviceInventoryResponse, error) {
	resp, err := common.RawRequest(ctx, s.APIClient, http.MethodGet, "/api/management/v1/inventory/devices?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var devices []client.DeviceInventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}
	return devices, nil
}
