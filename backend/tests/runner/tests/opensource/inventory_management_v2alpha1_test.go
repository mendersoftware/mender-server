package opensource

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
)

type InventoryManagementV2Alpha1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
}

func (i *BackendIntegrationSuite) TestInventoryManagementV2Alpha1() {
	suite.Run(i.T(), &InventoryManagementV2Alpha1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (u *InventoryManagementV2Alpha1Suite) SetupSuite() {
	require := require.New(u.T())

	ctx := common.BasicAuthContext(u.T().Context(), u.User)
	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.NotEmpty(token)
	require.Equal(http.StatusOK, r.StatusCode)
	u.JWT = token
}

func (u *InventoryManagementV2Alpha1Suite) TestSearchByDeviceIdentities() {
	var (
		ctx         = common.JWTAuthContext(u.T().Context(), u.JWT)
		deviceCount = 4
		devices     = make([]*common.Device, 0, deviceCount)

		mac                  = "mac"
		nameTagName          = "name"
		customNameIDAttrName = "custom"
		nameTagPrefix        = rand.Text()
	)

	{
		require := require.New(u.T())
		for idx := range deviceCount {
			// Used for search matching on tag "Name"
			nameTag := fmt.Sprintf("%s-%d", nameTagPrefix, idx+1)
			// Used to search for all, expecting to match on device with tag "Name" and
			// different device different identity with same value
			nameCustomID := fmt.Sprintf("%s-%d", nameTagPrefix, deviceCount-idx-1)

			device, err := common.NewDevice(
				u.APIClient,
				u.Tenant.TenantToken,
				common.WithIdentityAttributes(
					map[string]any{customNameIDAttrName: nameCustomID}),
			)
			require.NoError(err, "failed to create device identity")

			_, err = device.SubmitAuthRequest(ctx)
			require.NoError(err)

			err = device.AcceptFirst(ctx)
			require.NoError(err)

			_, err = u.APIClient.DeviceInventoryManagementAPIAPI.
				AddTags(ctx, device.ID).
				Tag([]client.Tag{
					{Name: nameTagName, Value: nameTag},
				}).
				Execute()
			require.NoError(err)

			devices = append(devices, device)
		}
	}

	u.Run("Success/DeviceID", func() {
		var (
			require        = require.New(u.T())
			expectedDevice = devices[0]
			valuePrefix    = expectedDevice.ID[:6]
		)

		results, _, err := u.APIClient.InventoryV2alpha1ManagementAPIAPI.
			SearchInventoryByIdentity(ctx).
			SearchIdentityParams(client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(client.IDENTITY), model.AttrNameID),
				ValuePrefix: valuePrefix,
				Attributes: []client.SelectAttribute{
					{Scope: client.IDENTITY, Attribute: mac},
				},
			}).Execute()
		require.NoError(err)

		// Check that the expected device is in the result
		require.True(slices.ContainsFunc(
			results,
			func(device client.DeviceInventoryResponse) bool { return expectedDevice.ID == device.GetId() },
		), "search did not find expected device")

		// Check that all devices in the result has id starting with the prefix
		// and that only expected attributes are returned
		for _, device := range results {
			require.True(
				strings.HasPrefix(device.GetId(), valuePrefix),
				"search found unexpected device",
			)
			expectedAttributes := []string{mac}
			for _, a := range device.GetAttributes() {
				require.Contains(
					expectedAttributes, a.Name, "unexpected attribute found",
				)
			}

			for _, a := range expectedAttributes {
				require.True(slices.ContainsFunc(
					device.GetAttributes(),
					func(aa client.AttributeResponse) bool { return aa.Name == a },
				), "expected attribute '%s' missing", a)
			}
		}
	})

	u.Run("Success/Mac", func() {
		var (
			require        = require.New(u.T())
			expectedDevice = devices[2]
			valuePrefix    = expectedDevice.MAC[:8]

			scope = client.IDENTITY
			name  = mac
		)

		results, _, err := u.APIClient.InventoryV2alpha1ManagementAPIAPI.
			SearchInventoryByIdentity(ctx).
			SearchIdentityParams(client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(scope), name),
				ValuePrefix: valuePrefix,
			}).Execute()
		require.NoError(err)

		// Check that the expected device is in the result
		require.True(slices.ContainsFunc(
			results,
			func(d client.DeviceInventoryResponse) bool {
				identity := pickStringAttributeValue(d.GetAttributes(), scope, name)
				return identity == expectedDevice.MAC
			},
		), "search did not find expected device")

		// Check that all devices in the result has mac starting with the prefix
		for _, device := range results {
			identity := pickStringAttributeValue(device.GetAttributes(), scope, name)
			require.True(
				strings.HasPrefix(identity, valuePrefix), "search found unexpected device",
			)
		}
	})

	u.Run("Success/Name", func() {
		var (
			require     = require.New(u.T())
			valuePrefix = nameTagPrefix

			scope    = client.TAGS
			name     = nameTagName
			pageSize = deviceCount / 2
			page     = 2
		)

		results, res, err := u.APIClient.InventoryV2alpha1ManagementAPIAPI.
			SearchInventoryByIdentity(ctx).
			SearchIdentityParams(client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(scope), name),
				ValuePrefix: valuePrefix,
				PerPage:     new(int32(pageSize)),
				Page:        new(int32(page)),
			}).Execute()
		require.NoError(err)
		require.NotNil(res)

		// Check that all devices in the result has name starting with the expected prefix
		// and is part of the expected page
		require.Len(results, pageSize, "unexpected number of results in results page")
		for _, device := range results {
			identity := pickStringAttributeValue(device.GetAttributes(), scope, name)
			require.True(
				strings.HasPrefix(identity, valuePrefix), "search found unexpected device",
			)

			nameFirstDeviceInPage := fmt.Sprintf("%s-%d", nameTagPrefix, pageSize+1)
			require.GreaterOrEqual(
				0, strings.Compare(nameFirstDeviceInPage, identity),
				"unexpected device in results page",
			)
		}

		// Check that the devices are consistently sorted
		isSorted := slices.IsSortedFunc(results, func(a, b client.DeviceInventoryResponse) int {
			identityA := pickStringAttributeValue(a.GetAttributes(), scope, name)
			identityB := pickStringAttributeValue(b.GetAttributes(), scope, name)
			return strings.Compare(identityA, identityB)
		})
		require.True(isSorted, "search results are not sorted correctly")

		// Check headers
		require.Equal(strconv.Itoa(deviceCount), res.Header.Get(rest.HeaderXTotalCount))
	})
	u.Run("Success/Status", func() {
		var (
			require     = require.New(u.T())
			valuePrefix = model.DeviceStatusAccepted

			scope = client.IDENTITY
			name  = model.AttrNameStatus
		)

		results, res, err := u.APIClient.InventoryV2alpha1ManagementAPIAPI.
			SearchInventoryByIdentity(ctx).
			SearchIdentityParams(client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(scope), name),
				ValuePrefix: valuePrefix,
			}).Execute()
		require.NoError(err)
		require.NotNil(res)

		// Check that all devices in the result has name starting with the expected prefix
		// and is part of the expected page
		require.Len(results, 0, "expected 0 results searching for status")
		// Check headers
		require.Equal("0", res.Header.Get(rest.HeaderXTotalCount))
	})

	u.Run("Success/All", func() {
		var (
			require     = require.New(u.T())
			valuePrefix = fmt.Sprintf("%s-%d", nameTagPrefix, 1)
		)

		results, res, err := u.APIClient.InventoryV2alpha1ManagementAPIAPI.
			SearchInventoryByIdentity(ctx).
			SearchIdentityParams(client.SearchIdentityParams{ValuePrefix: valuePrefix}).
			Execute()
		require.NoError(err)
		require.NotNil(res)

		// Check that all devices in the result has name or custom identity attribute
		// starting with the expected prefix
		for _, device := range results {
			name := pickStringAttributeValue(device.GetAttributes(), client.TAGS, nameTagName)
			custom := pickStringAttributeValue(device.GetAttributes(), client.IDENTITY, customNameIDAttrName)
			require.True(
				strings.HasPrefix(name, valuePrefix) || strings.HasPrefix(custom, valuePrefix),
				"search found unexpected device",
			)
		}

		// Check that the devices are consistently sorted
		isSorted := slices.IsSortedFunc(results, func(a, b client.DeviceInventoryResponse) int {
			identityA := pickStringAttributeValue(a.GetAttributes(), client.SYSTEM, model.AttrNameUpdated)
			identityB := pickStringAttributeValue(b.GetAttributes(), client.SYSTEM, model.AttrNameUpdated)
			return strings.Compare(identityA, identityB)
		})
		require.True(isSorted, "search results are not sorted correctly")
	})

	tcs := []struct {
		name   string
		params client.SearchIdentityParams
	}{
		{
			name: "MissingScope",
			params: client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute("", "irrelevant"),
				ValuePrefix: "also-irrelevant",
			},
		},
		{
			name: "MissingName",
			params: client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(client.IDENTITY), ""),
				ValuePrefix: "also-irrelevant",
			},
		},
		{
			name: "MissingValuePrefix",
			params: client.SearchIdentityParams{
				Attribute: client.NewSearchIdentitySpecificAttribute(string(client.IDENTITY), "irrelevant"),
			},
		},
		{
			name: "InvalidScope",
			params: client.SearchIdentityParams{
				Attribute: client.NewSearchIdentitySpecificAttribute("invalid", "irrelevant"),
			},
		},
		{
			name: "InvalidScopeForSearch",
			params: client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(client.INVENTORY), "irrelevant"),
				ValuePrefix: "also-irrelevant",
			},
		},
		{
			name: "WrongScopeForID",
			params: client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(client.INVENTORY), "id"),
				ValuePrefix: "abc",
			},
		},
		{
			name: "WrongNameForTag",
			params: client.SearchIdentityParams{
				Attribute:   client.NewSearchIdentitySpecificAttribute(string(client.TAGS), "my-name"),
				ValuePrefix: "abc",
			},
		},
	}

	for _, tc := range tcs {
		u.Run(fmt.Sprintf("Failure/%s", tc.name), func() {
			require := require.New(u.T())
			_, res, err := u.APIClient.InventoryV2alpha1ManagementAPIAPI.
				SearchInventoryByIdentity(ctx).
				SearchIdentityParams(tc.params).
				Execute()
			require.Error(err)
			require.Equal(http.StatusBadRequest, res.StatusCode)
		})
	}
}

func pickStringAttributeValue(attrs []client.AttributeResponse, scope client.Scope, name string) string {
	for _, a := range attrs {
		if a.Scope == scope && a.Name == name && a.Value.String != nil {
			return *a.Value.String
		}
	}
	return ""
}
