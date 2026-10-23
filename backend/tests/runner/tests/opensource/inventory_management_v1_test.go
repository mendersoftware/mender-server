//nolint:all // This is all test code
package opensource

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sort"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	authmodel "github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	// the total device count comes from the X-Total-Count header of a minimal
	// listing request.
	_, r, err := s.APIClient.DeviceInventoryManagementAPIAPI.ListDeviceInventories(ctx).PerPage(1).Execute()
	require.NoError(err)
	var count int
	_, err = fmt.Sscanf(r.Header.Get("X-Total-Count"), "%d", &count)
	require.NoError(err)

	for range 3 {
		_, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
	}

	_, r, err = s.APIClient.DeviceInventoryManagementAPIAPI.ListDeviceInventories(ctx).PerPage(1).Execute()
	require.NoError(err)
	var newCount int
	_, err = fmt.Sscanf(r.Header.Get("X-Total-Count"), "%d", &newCount)
	require.NoError(err)
	require.Equal(count+3, newCount)
}

func (s *InventoryManagementV1Suite) TestFilterDevicesOk() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	devs := make([]*common.Device, 3)
	for i := range devs {
		d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
		require.NoError(err)
		devs[i] = d
	}

	// give one device a distinct int-typed attribute and filter on its value
	numAttr := float32(31337)
	r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
		AssignAttributes(common.JWTAuthContext(ctx, devs[2].Token)).
		DeviceAttributeRequest([]client.DeviceAttributeRequest{
			{Name: "test-num-attr", Value: client.Float32AsAttributeValueRequest(&numAttr)},
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)

	// the v1 listing filters by an arbitrary attribute through query parameters
	// that the generated client doesn't model, so issue the raw request.
	resp, err := common.RawRequest(ctx, s.APIClient, http.MethodGet,
		"/api/management/v1/inventory/devices?"+url.Values{
			"per_page":      {"100"},
			"test-num-attr": {"31337"},
		}.Encode(), nil)
	require.NoError(err)
	defer resp.Body.Close()
	var apiDevs []client.DeviceInventoryResponse
	require.NoError(json.NewDecoder(resp.Body).Decode(&apiDevs))
	require.Len(apiDevs, 1)
}

func (s *InventoryManagementV1Suite) TestDevicePatchAttributes() {
	s.Run("Success/PatchAttributes", s.testDevicePatchAttributesOk)
	s.Run("Success/EmptyValue", s.testDevicePatchAttributesEmptyValue)
}

func (s *InventoryManagementV1Suite) testDevicePatchAttributesOk() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	// a random "mac" identity attribute is added automatically; add "sn" on top
	// so the identity data carries both (the test patches "sn" to empty and
	// asserts the exact identity-scope attribute count: status + mac + sn).
	sn := uuid.NewString()
	d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken,
		common.WithIdentityAttributes(map[string]any{"sn": sn}))
	require.NoError(err)

	payload := []client.DeviceAttributeRequest{
		{Name: "mac", Value: client.StringAsAttributeValueRequest(new("mac-new-" + d.ID))},
		// empty value for existing
		{Name: "sn", Value: client.StringAsAttributeValueRequest(new(""))},
		// empty value for new
		{Name: "new-empty", Value: client.StringAsAttributeValueRequest(new(""))},
	}
	devCtx := common.JWTAuthContext(ctx, d.Token)
	r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
		AssignAttributes(devCtx).
		DeviceAttributeRequest(payload).
		Execute()
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)

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
			require.Equal("mac-new-"+apiDev.GetId(), str)
		case a.GetName() == "sn" && a.GetScope() == client.INVENTORY:
			require.Equal("", str)
		case a.GetName() == "new-empty" && a.GetScope() == client.INVENTORY:
			require.Equal("", str)
		case a.GetName() == "status" && a.GetScope() == client.IDENTITY:
			require.Contains([]string{authmodel.DevStatusAccepted, authmodel.DevStatusPending}, str)
		case a.GetScope() != client.INVENTORY:
			require.NotEqual("", str, "unexpected empty value for %s", a.GetName())
		default:
			require.Fail("unexpected attribute " + a.GetName())
		}
	}
}

func (s *InventoryManagementV1Suite) testDevicePatchAttributesEmptyValue() {
	require := require.New(s.T())
	ctx := common.JWTAuthContext(s.T().Context(), s.JWT)

	d, err := common.NewAcceptedDevice(ctx, s.APIClient, s.Tenant.TenantToken)
	require.NoError(err)

	r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
		AssignAttributes(common.JWTAuthContext(ctx, d.Token)).
		DeviceAttributeRequest([]client.DeviceAttributeRequest{
			{Name: "mac", Value: client.StringAsAttributeValueRequest(new(""))},
		}).Execute()
	require.NoError(err)
	require.Equal(http.StatusOK, r.StatusCode)
}

func (s *InventoryManagementV1Suite) TestDeviceFilteringSearchV2() {
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
		require.NoError(s.T(), err)

		payload := []client.DeviceAttributeRequest{
			{Name: "artifact", Value: client.ArrayOfStringAsAttributeValueRequest(&[]string{"v1"})},
			{Name: "idx", Value: client.Float32AsAttributeValueRequest(new(float32(i)))},
		}
		devCtx := common.JWTAuthContext(ctx, d.Token)
		r, err := s.APIClient.DeviceInventoryDeviceAPIAPI.
			AssignAttributes(devCtx).
			DeviceAttributeRequest(payload).
			Execute()
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusOK, r.StatusCode)

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
			{name: "idx", scope: client.INVENTORY, value: client.Float32AsAttributeValueRequest(new(dev.idx))},
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
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(new(float32(1)))},
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
					{Type: "$eq", Attribute: "id_data", Scope: client.INVENTORY, Value: client.StringAsAttributeValueRequest(new("illegal_data"))},
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
					{Attribute: "artifact", Scope: client.INVENTORY, Value: client.StringAsAttributeValueRequest(new("v1"))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/InvalidFilterScope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.Scope("user_defined"), Value: client.Float32AsAttributeValueRequest(new(float32(1)))},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Failure/InvalidSortScope",
			params: client.SearchParams{
				DeviceIds: ids,
				Filters: []client.FilterPredicate{
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(new(float32(1)))},
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
					{Type: "$eq", Attribute: "idx", Scope: client.INVENTORY, Value: client.Float32AsAttributeValueRequest(new(float32(1)))},
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
					{Type: "$eq", Attribute: "idx", Scope: client.Scope("../../../../Windows/system.ini"), Value: client.Float32AsAttributeValueRequest(new(float32(1)))},
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
			require := require.New(s.T())
			results, r, err := s.APIClient.DeviceInventoryFiltersAndSearchManagementAPIAPI.
				InventoryV2SearchDeviceInventories(ctx).
				SearchParams(tc.params).
				Execute()
			if tc.wantStatus == http.StatusOK {
				require.NoError(err)
			} else {
				require.Error(err)
			}
			require.Equal(tc.wantStatus, r.StatusCode)

			if tc.wantStatus != http.StatusOK {
				return
			}

			require.Len(results, len(tc.expected), "unexpected number of results")
			for i, exp := range tc.expected {
				require.Equal(exp.id, results[i].GetId(), "unexpected device in response")
				for _, wantAttr := range exp.attrs {
					require.True(
						slices.ContainsFunc(results[i].Attributes, func(a client.AttributeResponse) bool {
							if a.GetName() != wantAttr.name || a.GetScope() != wantAttr.scope {
								return false
							}
							v := a.GetValue()
							switch {
							case wantAttr.value.String != nil:
								return v.String != nil && *v.String == *wantAttr.value.String
							case wantAttr.value.Float32 != nil:
								return v.Float32 != nil && *v.Float32 == *wantAttr.value.Float32
							case wantAttr.value.ArrayOfString != nil:
								if v.ArrayOfString == nil || len(*v.ArrayOfString) != len(*wantAttr.value.ArrayOfString) {
									return false
								}
								for j, want := range *wantAttr.value.ArrayOfString {
									if (*v.ArrayOfString)[j] != want {
										return false
									}
								}
								return true
							case wantAttr.value.ArrayOfFloat32 != nil:
								if v.ArrayOfFloat32 == nil || len(*v.ArrayOfFloat32) != len(*wantAttr.value.ArrayOfFloat32) {
									return false
								}
								for j, want := range *wantAttr.value.ArrayOfFloat32 {
									if (*v.ArrayOfFloat32)[j] != want {
										return false
									}
								}
								return true
							}
							return false
						}),
						"missing inventory attribute %s on device %s", wantAttr.name, exp.id,
					)
				}
			}
		})
	}
}
