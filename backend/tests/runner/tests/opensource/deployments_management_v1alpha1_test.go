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
	dmodel "github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/tests/runner/tests/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DeploymentsManagementV1Alpha1Suite struct {
	suite.Suite

	APIClient *client.APIClient
	User      common.User
	Tenant    common.Tenant

	JWT string
}

func (i *BackendIntegrationSuite) TestDeploymentsManagementV1Alpha1() {
	suite.Run(i.T(), &DeploymentsManagementV1Alpha1Suite{
		APIClient: i.environment.APIClient(),
		User:      i.user,
		Tenant:    i.tenant,
	})
}

func (u *DeploymentsManagementV1Alpha1Suite) SetupTest() {
	require := require.New(u.T())

	ctx := common.BasicAuthContext(u.T().Context(), u.User)
	token, r, err := u.APIClient.UserAdministrationManagementAPIAPI.Login(ctx).Execute()

	require.NoError(err)
	require.NotNil(r)
	require.NotZero(len(token))
	require.Equal(200, r.StatusCode)
	u.JWT = token
}

func (u *DeploymentsManagementV1Alpha1Suite) TestListSoftwareTags() {

	// Create two releases and tag them so we have something to list
	allTags := []string{"tag1", "tag2", "tag3"}
	{
		var (
			t       = u.T()
			ctx     = common.JWTAuthContext(t.Context(), u.JWT)
			require = require.New(t)
		)
		for idx := range 2 {
			name := uuid.NewString()

			_, err := u.createReleaseArtifact(ctx, t, name)
			require.NoError(err)

			// tag the first release with tag1 and tag2, the second with tag2 and tag3
			_, err = u.APIClient.DeploymentsV2ManagementAPIAPI.AssignReleaseTags(ctx, name).
				RequestBody(allTags[idx : len(allTags)-1+idx]).
				Execute()
			require.NoError(err)
		}

		// create a release without a tag to check empty omission
		_, err := u.createReleaseArtifact(ctx, t, uuid.NewString())
		require.NoError(err)
	}

	u.Run("Success/NoFilter", func() {
		ctx := u.T().Context()
		softwareTags, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			ListSoftwareTags(common.JWTAuthContext(ctx, u.JWT)).
			Execute()
		require.NoError(u.T(), err)
		assert.ElementsMatch(u.T(), softwareTags, allTags)
	})

	u.Run("Success/FilterReleases", func() {
		ctx := u.T().Context()
		softwareTags, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			ListSoftwareTags(common.JWTAuthContext(ctx, u.JWT)).
			Kind(*client.PtrString(dmodel.ReleaseKindRelease)).
			Execute()
		require.NoError(u.T(), err)
		assert.ElementsMatch(u.T(), softwareTags, allTags)
	})

	u.Run("Success/FilterManifest", func() {
		ctx := u.T().Context()
		softwareTags, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			ListSoftwareTags(common.JWTAuthContext(ctx, u.JWT)).
			Kind(*client.PtrString(dmodel.ReleaseKindManifest)).
			Execute()
		require.NoError(u.T(), err)
		assert.Len(u.T(), softwareTags, 0)
	})

	u.Run("Failure/InvalidKind", func() {
		ctx := u.T().Context()
		_, res, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			ListSoftwareTags(common.JWTAuthContext(ctx, u.JWT)).
			Kind(*client.PtrString("invalid-kind")).
			Execute()
		require.Error(u.T(), err)
		assert.Equal(u.T(), http.StatusBadRequest, res.StatusCode)
	})
}

func (u *DeploymentsManagementV1Alpha1Suite) createReleaseArtifact(
	ctx context.Context,
	t interface{ TempDir() string },
	name string) (*os.File, error) {

	file, err := os.Create(path.Join(t.TempDir(), name))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.WriteString("Hello world!")
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	_, err = u.APIClient.DeploymentsManagementAPIAPI.
		GenerateArtifact(ctx).
		Name(name).
		Type_("single_file").
		DeviceTypesCompatible([]string{"device-type-1"}).
		Args(fmt.Sprintf(`{"dest_dir":"/", "filename":"%s"}`, name)).
		File(file).
		Execute()
	if err != nil {
		return nil, err
	}
	// Wait for the async processing to complete
	created := false
	for range 5 {
		_, res, err := u.APIClient.DeploymentsV2ManagementAPIAPI.GetReleaseWithGivenName(ctx, name).Execute()
		if err != nil {
			if http.StatusNotFound != res.StatusCode {
				return nil, errors.New("unexpected status code from get release")
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		created = true
		break
	}
	if !created {
		return nil, errors.New("artifact not created in time")
	}
	return file, nil
}
