package opensource

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-artifact/handlers"
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
	require.NotEmpty(token)
	require.Equal(http.StatusOK, r.StatusCode)
	u.JWT = token
}

func generateLinkHeader(uri, query, linkName string) string {
	linkTmpl := "<%s?%s>; rel=\"%s\""
	return fmt.Sprintf(linkTmpl, uri, query, linkName)
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

func (u *DeploymentsManagementV1Alpha1Suite) TestListSoftware() {
	allSoftwares := []client.Software{}
	{
		t := u.T()
		ctx := common.JWTAuthContext(t.Context(), u.JWT)
		require := require.New(t)
		for idx := range 2 {
			name := fmt.Sprintf("test-list-software-%d-%s", idx, uuid.NewString())
			_, err := u.createReleaseArtifact(ctx, t, name)
			require.NoError(err)

			allSoftwares = append(allSoftwares, client.Software{
				Name: name,
				Kind: dmodel.ReleaseKindRelease,
			})

		}

		// create a release with update type
		name := fmt.Sprintf("test-not-a-manifest-%s", uuid.NewString())
		i := handlers.NewModuleImage("definitely-not-a-manifest")
		artifact, err := common.CreateArtifact(name, t, common.WithModuleImage(i))
		require.NoError(err)
		res, err := u.uploadArtifact(ctx, artifact, client.PtrString("don't trust me? check it yourself!"))
		require.NoError(err)
		require.Equal(res.StatusCode, http.StatusCreated)
		allSoftwares = append(allSoftwares, client.Software{
			Name: name,
			Kind: dmodel.ReleaseKindRelease,
		})
	}
	compareSoftwares := func(results []client.Software, expected ...client.Software) error {
		if len(expected) != len(results) {
			return fmt.Errorf("length mismatch: expected %d items, got %d", len(expected), len(results))
		}

		for i := range expected {
			if expected[i].Name != results[i].Name {
				return fmt.Errorf("name mismatch at index %d: expected %s, got %s",
					i, expected[i].Name, results[i].Name)
			}
			if expected[i].Kind != results[i].Kind {
				return fmt.Errorf("kind mismatch at index %d for '%s': expected %v, got %v",
					i, expected[i].Name, expected[i].Kind, results[i].Kind)
			}
		}
		return nil
	}

	u.Run("Success/NoFilter", func() {
		ctx := u.T().Context()
		softwares, res, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Execute()
		require.NoError(u.T(), err)

		// the count of softwares is effacted by other tests in same suite.
		// there shoud be minimum 3 softwares.
		require.True(u.T(), len(softwares) > 2)
		require.NotEmpty(u.T(), res.Header.Get("X-Total-Count"))
	})

	u.Run("Success/CompatibleTypes", func() {
		var (
			ctx     = common.JWTAuthContext(u.T().Context(), u.JWT)
			require = require.New(u.T())
		)

		// add second artifact to first release
		i := handlers.NewModuleImage("single-file")
		artifact, err := common.CreateArtifact(
			allSoftwares[0].Name,
			u.T(),
			common.WithModuleImage(i),
			common.WithCompatibleDevices([]string{"device-type-2"}),
		)
		require.NoError(err)
		_, err = u.uploadArtifact(ctx, artifact, nil)
		require.NoError(err)

		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(ctx).
			Name([]string{allSoftwares[0].Name}).
			Execute()
		require.NoError(err)
		require.Len(softwares, 1, "expected one software to be returned")
		require.ElementsMatch(
			[]string{"device-type-1", "device-type-2"},
			softwares[0].GetCompatibleTypes(),
		)
	})

	u.Run("Success/FilterExactName", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Name([]string{allSoftwares[0].Name}).
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[0])
		require.NoError(u.T(), err)
	})

	u.Run("Success/FilterMultipleExactNames", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Name([]string{allSoftwares[0].Name, allSoftwares[1].Name}).
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[0], allSoftwares[1])
		require.NoError(u.T(), err)
	})

	u.Run("Success/FilterNamePrefix", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			NamePrefix("test-list-software").
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[0], allSoftwares[1])
		require.NoError(u.T(), err)
	})

	u.Run("Failure/FilterNamePrefix&ExactName", func() {
		ctx := u.T().Context()
		_, res, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			NamePrefix("test-list-software").
			Name([]string{allSoftwares[2].Name}).
			Execute()

		require.Error(u.T(), err)
		assert.Equal(u.T(), http.StatusBadRequest, res.StatusCode)

	})

	u.Run("Success/FilterReleases", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Kind(*client.PtrString(dmodel.ReleaseKindRelease)).
			NamePrefix("test-list-software"). // filter so we dont collide with other tests
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[0], allSoftwares[1])
		require.NoError(u.T(), err)
	})

	u.Run("Success/FilterManifest", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Kind(*client.PtrString(dmodel.ReleaseKindManifest)).
			Execute()

		require.NoError(u.T(), err)
		assert.Len(u.T(), softwares, 0)
	})

	u.Run("Success/FilterKind&Name", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Kind(*client.PtrString(dmodel.ReleaseKindRelease)).
			Name([]string{allSoftwares[2].Name}).
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[2])
		require.NoError(u.T(), err)
	})

	u.Run("Success/FilterTags", func() {
		ctx := common.JWTAuthContext(u.T().Context(), u.JWT)
		tags := []string{"tag1"}

		_, err := u.APIClient.DeploymentsV2ManagementAPIAPI.
			AssignReleaseTags(ctx, allSoftwares[0].Name).
			RequestBody(tags).
			Execute()
		require.NoError(u.T(), err)

		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(ctx).
			Kind(*client.PtrString(dmodel.ReleaseKindRelease)).
			Tag(tags).
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[0])
		require.NoError(u.T(), err)
	})

	u.Run("Success/Paging", func() {
		ctx := u.T().Context()
		softwares, res, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Name([]string{allSoftwares[0].Name, allSoftwares[1].Name, allSoftwares[2].Name}).
			Page(2).
			PerPage(1).
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[1])
		require.NoError(u.T(), err)

		actualLinkHeaders := res.Header.Values("Link")

		query := res.Request.URL.Query()

		query.Set("page", "3")
		expectedNextLink := generateLinkHeader(res.Request.URL.Path, query.Encode(), "next")
		assert.Contains(u.T(), actualLinkHeaders, expectedNextLink)

		query.Set("page", "1")
		expectedPrevLink := generateLinkHeader(res.Request.URL.Path, query.Encode(), "prev")
		assert.Contains(u.T(), actualLinkHeaders, expectedPrevLink)

		expectedFirstLink := generateLinkHeader(res.Request.URL.Path, query.Encode(), "first")
		assert.Contains(u.T(), actualLinkHeaders, expectedFirstLink)
	})

	u.Run("Success/FilterUpdateType", func() {
		ctx := u.T().Context()
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			UpdateType("definitely-not-a-manifest").
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[2])
		require.NoError(u.T(), err)

		softwares, _, err = u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			UpdateType("single-file").
			NamePrefix("test-list-software"). // filter so we dont collide with other tests
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, allSoftwares[0], allSoftwares[1])
		require.NoError(u.T(), err)
	})

	u.Run("Success/Sort", func() {
		ctx := u.T().Context()

		// name:asc is the default, we don't need to test it
		nameDescOrder := []client.Software{allSoftwares[1], allSoftwares[0]}
		softwares, _, err := u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Sort("name:desc").
			NamePrefix("test-list-software"). // filter so we dont collide with other tests
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, nameDescOrder...)
		require.NoError(u.T(), err)

		softwares, _, err = u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Sort("modified:asc").
			NamePrefix("test-list-software").
			Execute()
		require.NoError(u.T(), err)

		updatedOrder := []client.Software{allSoftwares[1], allSoftwares[0]}
		err = compareSoftwares(softwares, updatedOrder...)
		require.NoError(u.T(), err)

		updatedDescOrder := []client.Software{allSoftwares[0], allSoftwares[1]}
		softwares, _, err = u.APIClient.DeploymentsV1alpha1ManagementAPIAPI.
			GetDeploymentSoftware(common.JWTAuthContext(ctx, u.JWT)).
			Sort("modified:desc").
			NamePrefix("test-list-software").
			Execute()
		require.NoError(u.T(), err)

		err = compareSoftwares(softwares, updatedDescOrder...)
		require.NoError(u.T(), err)
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
	for range 15 {
		_, res, err := u.APIClient.DeploymentsV2ManagementAPIAPI.GetReleaseWithGivenName(ctx, name).Execute()
		if err != nil {
			if http.StatusNotFound != res.StatusCode {
				return nil, errors.New("unexpected status code from get release")
			}
			time.Sleep(time.Second)
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

func (u *DeploymentsManagementV1Alpha1Suite) uploadArtifact(
	ctx context.Context,
	artifact *os.File,
	description *string,
) (*http.Response, error) {
	defer artifact.Close()

	req := u.APIClient.DeploymentsManagementAPIAPI.
		UploadArtifact(ctx).
		Artifact(artifact)

	if description != nil {
		req = req.Description(*description)
	}
	r, err := req.Execute()
	if r == nil {
		return nil, errors.New("got no response from upload artifact")
	}
	return r, err
}
