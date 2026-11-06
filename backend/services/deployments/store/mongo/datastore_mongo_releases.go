// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func (db *DataStoreMongo) InsertRelease(
	ctx context.Context,
	release model.Release,
) error {
	err := db.checkTagQuota(ctx, release.Tags)
	if err != nil {
		return err
	}
	release.Created = new(time.Now())
	type Release struct {
		model.Release `bson:",inline"`
		TenantID      string `bson:"tenant_id"`
	}

	_, err = db.database.Collection(CollectionReleases).
		InsertOne(ctx, Release{
			Release:  release,
			TenantID: identity.FromContext(ctx).GetTenantID(),
		})
	return err
}

func (db *DataStoreMongo) UpdateReleaseArtifactDescription(
	ctx context.Context,
	artifactToEdit *model.Image,
	releaseName string,
) error {
	collReleases := db.database.Collection(CollectionReleases)

	update := bson.M{
		"$set": bson.M{
			StorageKeyReleaseArtifactsIndexDescription: artifactToEdit.ImageMeta.Description,
			StorageKeyReleaseArtifactsIndexModified:    artifactToEdit.Modified,
			StorageKeyReleaseModified:                  time.Now(),
		},
	}
	_, err := collReleases.UpdateOne(
		ctx,
		bson.M{
			StorageKeyReleaseKind: bson.M{
				StorageKeyTenantId: identity.FromContext(ctx).GetTenantID(),
				"$in":              bson.A{nil, model.ReleaseKindRelease},
			},
			StorageKeyReleaseName:        releaseName,
			StorageKeyReleaseArtifactsId: artifactToEdit.Id,
		},
		update,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *DataStoreMongo) AddArtifactToRelease(
	ctx context.Context,
	artifact model.Image,
) error {
	collReleases := db.database.Collection(CollectionReleases)
	var (
		dependsFilters []bson.D
		dependsKeys    []string
	)
	if len(artifact.Depends) == 0 {
		artifact.Depends = map[string][]string{
			"device_type": artifact.DeviceTypesCompatible,
		}
	} else if len(artifact.Depends["device_type"]) == 0 {
		artifact.Depends["device_type"] = artifact.DeviceTypesCompatible
	}
	for key, values := range artifact.Depends {
		dependsKeys = append(dependsKeys, key)
		dependsTerms := make([]Depend, len(values))
		for i, value := range values {
			dependsTerms[i].Key = key
			dependsTerms[i].Value = value
		}
		dependsFilters = append(dependsFilters, bson.D{
			{Key: StorageKeyReleaseArtifacts, Value: bson.D{
				// For each artifact
				{Key: "$elemMatch", Value: bson.D{
					{Key: `meta_artifact.depends`, Value: bson.D{
						// For each depends term
						{Key: "$elemMatch", Value: bson.D{
							// Term matches
							{Key: "$in", Value: dependsTerms},
						}},
					}},
				}},
			}},
		})
	}
	// This artificially produce the same uniqueness constraint as the former depends_idx.
	// The upsert will make a conflicting artifact trigger an insert causing index violation
	// on the release name.
	// First $or term
	filter := bson.M{
		StorageKeyTenantId:    identity.FromContext(ctx).GetTenantID(),
		StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
		StorageKeyReleaseName: artifact.Name,

		// FIXME: Depends conflict is not detected and returned as before (no 409)
		// Uniqueness by depends is overly complicated and does not make any sense.
		// Consider changing with uniqueness on provides and update type:
		// {
		// 	StorageKeyReleaseArtifacts: {
		// 		$elemMatch: {
		// 			meta_artifact.provides: {
		// 				$ne: artifact.Provides
		// 			}
		// 			meta_artifact.updates.type_info.type: {
		// 				$nin: artifact.Updates.map(update => update.TypeInfo.Type)
		// 			}
		// 		}
		// 	}
		// }
		"$or": []bson.D{
			{{Key: StorageKeyReleaseArtifacts, Value: bson.D{
				// For each artifact
				{Key: "$elemMatch", Value: bson.D{
					// No keys outside dependsKeys
					{Key: `meta_artifact.depends.k`, Value: bson.D{
						{Key: "$nin", Value: dependsKeys},
					}},
				}},
			}}},
			// At least one depends term per key matches
			// - or inverse, there is no complete set of key-value pairs that overlap.
			{{Key: "$nor", Value: []bson.D{
				{{Key: "$and", Value: dependsFilters}},
			}}},
		},
	}
	update := bson.M{
		"$set": bson.M{
			StorageKeyReleaseName:     artifact.Name,
			StorageKeyReleaseModified: time.Now(),
		},
	}
	update["$push"] = bson.M{StorageKeyReleaseArtifacts: artifact}
	update["$inc"] = bson.M{
		StorageKeyReleaseArtifactsCount: 1,
	}
	update["$setOnInsert"] = bson.M{
		StorageKeyReleaseCreated: time.Now(),
		StorageKeyReleaseKind:    model.ReleaseKindRelease,
	}
	_, err := collReleases.UpdateOne(ctx, filter, update, mopts.UpdateOne().
		SetUpsert(true))
	// TODO: Handle conflicts a better
	return err
}

func (db *DataStoreMongo) RemoveArtifactFromRelease(
	ctx context.Context,
	artifactID string,
) error {
	collReleases := db.database.Collection(CollectionReleases)
	tenantID := identity.FromContext(ctx).GetTenantID()
	filter := bson.M{
		StorageKeyReleaseKind:        bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
		StorageKeyReleaseArtifactsId: artifactID,
		StorageKeyTenantId:           tenantID,
	}
	update := bson.M{
		"$set":  bson.M{StorageKeyReleaseModified: time.Now()},
		"$pull": bson.M{StorageKeyReleaseArtifacts: bson.M{StorageKeyId: artifactID}},
		"$inc":  bson.M{StorageKeyReleaseArtifactsCount: -1},
	}
	var res model.Release
	err := collReleases.FindOneAndUpdate(ctx, filter, update, mopts.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(mopts.After).
		SetProjection(bson.M{
			StorageKeyReleaseName:           1,
			StorageKeyReleaseArtifactsCount: 1,
		}),
	).
		Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return store.ErrNotFound
		}
		return err
	}
	if res.ArtifactsCount == 0 {
		// Cleanup empty release
		_, err := collReleases.DeleteOne(ctx, bson.M{
			StorageKeyTenantId:              tenantID,
			StorageKeyReleaseName:           res.Name,
			StorageKeyReleaseArtifactsCount: 0,
		})
		if err != nil {
			return fmt.Errorf("failed to clean up empty release: %w", err)
		}
	}
	return err
}

func (db *DataStoreMongo) ListReleaseTags(ctx context.Context) (model.Tags, error) {
	filter := bson.M{
		StorageKeyTenantId: identity.FromContext(ctx).GetTenantID(),

		StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
	}
	res := db.database.
		Collection(CollectionReleases).
		Distinct(ctx, StorageKeyReleaseTags, filter)

	var ret model.Tags
	raw, err := res.Raw()
	if err == nil {
		err = bson.UnmarshalValue(bson.TypeArray, raw, &ret)
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (db *DataStoreMongo) checkTagQuota(ctx context.Context, tags model.Tags) error {
	// Check preconditions
	if len(tags) > model.TagsMaxUnique {
		return model.ErrTooManyUniqueTags
	}

	if len(tags) > 0 {
		inUseTags, err := db.ListSoftwareTags(ctx, nil)
		if err != nil {
			return errors.WithMessage(err, "mongo: failed to count in-use tags")
		}
		tagSet := make(map[model.Tag]struct{}, len(inUseTags))
		for _, tagKey := range inUseTags {
			tagSet[tagKey] = struct{}{}
		}
		for _, tag := range tags {
			delete(tagSet, tag)
		}
		if len(tags)+len(tagSet) > model.TagsMaxUnique {
			return model.ErrTooManyUniqueTags
		}
	}
	return nil
}

func (db *DataStoreMongo) ReplaceReleaseTags(
	ctx context.Context,
	releaseName string,
	tags model.Tags,
) error {

	collReleases := db.database.Collection(CollectionReleases)

	err := db.checkTagQuota(ctx, tags)
	if err != nil {
		return err
	}

	// Update release tags
	res, err := collReleases.UpdateOne(ctx, bson.D{
		{Key: StorageKeyTenantId, Value: identity.FromContext(ctx).GetTenantID()},
		{Key: StorageKeyReleaseKind, Value: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}}},
		{Key: StorageKeyReleaseName, Value: releaseName},
	}, bson.D{{
		Key:   mongoOpSet,
		Value: bson.D{{Key: StorageKeyReleaseTags, Value: tags}},
	}})
	if err != nil {
		return errors.WithMessage(err, "mongo: failed to update release tags")
	} else if res.MatchedCount <= 0 {
		return store.ErrNotFound
	}
	return nil
}

func (db *DataStoreMongo) UpdateRelease(
	ctx context.Context,
	releaseName string,
	release model.ReleasePatch,
) error {
	collReleases := db.database.Collection(CollectionReleases)

	err := release.Validate()
	if err != nil {
		return errors.Wrap(err, "cant update release due to validation errors")
	}

	// Update release, at the moment we update only the notes,
	// it is on purpose that we take only this field explicitly,
	// once there is a need we can extend
	filter := bson.M{
		StorageKeyTenantId: identity.FromContext(ctx).GetTenantID(),

		StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
		StorageKeyReleaseName: releaseName,
	}
	res, err := collReleases.UpdateOne(
		ctx,
		filter,
		bson.D{
			{
				Key: mongoOpSet,
				Value: bson.D{
					{
						Key: StorageKeyReleaseNotes, Value: release.Notes,
					},
				},
			},
		},
	)
	if err != nil {
		return errors.WithMessage(err, "mongo: failed to update release")
	} else if res.MatchedCount <= 0 {
		return store.ErrNotFound
	}
	return nil
}

// Get the update types
func (db *DataStoreMongo) GetUpdateTypes(ctx context.Context) ([]string, error) {
	c := db.database.Collection(CollectionReleases)
	res := c.Distinct(ctx, StorageKeyReleaseArtifactsUpdateTypes, bson.M{
		StorageKeyTenantId:                    identity.FromContext(ctx).GetTenantID(),
		StorageKeyReleaseArtifactsUpdateTypes: bson.M{"$type": "string"},
	})
	var result []string
	raw, err := res.Raw()
	if err == mongo.ErrNoDocuments {
		return []string{}, nil
	} else if err == nil {
		// There seems to be a bug where DistinctResult.Decode does not work as expected
		// falling back to bson.UnmarshalValue.
		err = bson.UnmarshalValue(bson.TypeArray, raw, &result)
	}
	if err != nil {
		return []string{}, err
	} else {
		return result, nil
	}
}

func (db *DataStoreMongo) DeleteReleasesByNames(ctx context.Context, names []string) error {
	collDevs := db.database.Collection(CollectionReleases)
	query := bson.M{
		StorageKeyTenantId: identity.FromContext(ctx).GetTenantID(),

		StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
		StorageKeyReleaseName: bson.M{
			"$in": names,
		},
	}
	_, err := collDevs.DeleteMany(ctx, query)
	return err
}

func (db *DataStoreMongo) GetRelease(
	ctx context.Context,
	releaseName string,
) (*model.Release, error) {

	collReleases := db.database.Collection(CollectionReleases)

	release := new(model.Release)
	if err := collReleases.FindOne(ctx, bson.M{
		StorageKeyTenantId: identity.FromContext(ctx).GetTenantID(),

		StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
		StorageKeyReleaseName: releaseName,
	}).
		Decode(release); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return release, nil
}
