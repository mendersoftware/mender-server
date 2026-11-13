// Copyright 2026 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package mongo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func TestGetSoftware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetSoftware in short mode.")
	}

	const (
		tenantID      = "123456789012345678901234"
		otherTenantID = "abcdefabcdefabcdefabcdef"
	)

	newImage := func(id, name string, deviceTypes ...string) model.Image {
		return model.Image{
			Id:        id,
			ImageMeta: &model.ImageMeta{Description: "description"},
			ArtifactMeta: &model.ArtifactMeta{
				Name:                  name,
				DeviceTypesCompatible: deviceTypes,
				// Initialize to empty values to match the decoded document.
				Updates: []model.Update{},
				Depends: model.Depends{},
			},
		}
	}

	type document struct {
		model.Software `bson:",inline"`
		TenantID       string `bson:"tenant_id"`
	}

	release := model.Software{
		ReleaseBase: model.ReleaseBase{
			Name: "release-1",
			Kind: model.ReleaseKindRelease,
			Tags: model.Tags{"foo", "bar"},
		},
		Artifacts: []model.Image{
			newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d80", "release-1", "foo"),
			newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d81", "release-1", "bar", "baz"),
		},
	}
	manifest := model.Software{
		ReleaseBase: model.ReleaseBase{
			Name: "manifest-1",
			Kind: model.ReleaseKindManifest,
		},
		Artifacts: []model.Image{
			newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d82", "manifest-1", "system-a"),
		},
	}
	noArtifacts := model.Software{
		ReleaseBase: model.ReleaseBase{
			Name: "no-artifacts",
			Kind: model.ReleaseKindRelease,
		},
		Artifacts: []model.Image{},
	}
	otherTenantRelease := model.Software{
		ReleaseBase: model.ReleaseBase{
			Name: "other-tenant-only",
			Kind: model.ReleaseKindRelease,
		},
		Artifacts: []model.Image{
			newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d83", "other-tenant-only", "foo"),
		},
	}
	// Same name as `release`, but owned by another tenant.
	otherTenantShadow := model.Software{
		ReleaseBase: model.ReleaseBase{
			Name: "release-1",
			Kind: model.ReleaseKindRelease,
		},
		Artifacts: []model.Image{
			newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d84", "release-1", "other"),
		},
	}

	db.Wipe()
	ds := NewDataStoreMongoWithClient(db.Client())
	collReleases := ds.database.Collection(CollectionReleases)

	ctx := context.Background()
	_, err := collReleases.InsertMany(ctx, []any{
		document{Software: release, TenantID: tenantID},
		document{Software: manifest, TenantID: tenantID},
		document{Software: noArtifacts, TenantID: tenantID},
		document{Software: otherTenantRelease, TenantID: otherTenantID},
		document{Software: otherTenantShadow, TenantID: otherTenantID},
		// Legacy release document predating the "kind" attribute.
		bson.M{
			StorageKeyTenantId:    tenantID,
			StorageKeyReleaseName: "legacy",
			"artifacts": bson.A{
				newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d85", "legacy", "foo"),
			},
		},
	})
	require.NoError(t, err)

	tenantCtx := identity.WithContext(ctx, &identity.Identity{Tenant: tenantID})

	testCases := map[string]struct {
		ctx  context.Context
		name string

		software *model.Software
		err      error
	}{
		"ok, release": {
			ctx:      tenantCtx,
			name:     release.Name,
			software: &release,
		},
		"ok, manifest": {
			ctx:      tenantCtx,
			name:     manifest.Name,
			software: &manifest,
		},
		"ok, no artifacts": {
			ctx:      tenantCtx,
			name:     noArtifacts.Name,
			software: &noArtifacts,
		},
		"ok, legacy document without kind": {
			ctx:  tenantCtx,
			name: "legacy",
			software: &model.Software{
				ReleaseBase: model.ReleaseBase{Name: "legacy"},
				Artifacts: []model.Image{
					newImage("6d4f6e27-c3bb-438c-ad9c-d9de30e59d85", "legacy", "foo"),
				},
			},
		},
		"ok, other tenant": {
			ctx: identity.WithContext(ctx,
				&identity.Identity{Tenant: otherTenantID},
			),
			name:     otherTenantShadow.Name,
			software: &otherTenantShadow,
		},
		"error, not found": {
			ctx:  tenantCtx,
			name: "does-not-exist",
			err:  store.ErrNotFound,
		},
		"error, name is case sensitive": {
			ctx:  tenantCtx,
			name: "Release-1",
			err:  store.ErrNotFound,
		},
		"error, belongs to other tenant": {
			ctx:  tenantCtx,
			name: otherTenantRelease.Name,
			err:  store.ErrNotFound,
		},
		"error, no tenant in context": {
			ctx:  ctx,
			name: release.Name,
			err:  store.ErrNotFound,
		},
		"error, context canceled": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(tenantCtx)
				cancel()
				return ctx
			}(),
			name: release.Name,
			err:  context.Canceled,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			software, err := ds.GetSoftware(tc.ctx, tc.name)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				assert.Nil(t, software)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.software, software)
		})
	}
}
