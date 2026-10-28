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
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

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

func (db *DataStoreMongo) UpdateReleaseArtifacts(
	ctx context.Context,
	artifactToAdd *model.Image,
	artifactToRemove *model.Image,
	releaseName string,
) error {
	collReleases := db.database.Collection(CollectionReleases)

	opt := mopts.UpdateOne()
	update := bson.M{
		"$set": bson.M{
			StorageKeyReleaseName:     releaseName,
			StorageKeyReleaseModified: time.Now(),
		},
	}
	if artifactToRemove != nil {
		update["$pull"] = bson.M{
			StorageKeyReleaseArtifacts: bson.M{StorageKeyId: artifactToRemove.Id},
		}
		update["$inc"] = bson.M{
			StorageKeyReleaseArtifactsCount: -1,
		}
	}
	if artifactToAdd != nil {
		opt.SetUpsert(true)
		update["$push"] = bson.M{StorageKeyReleaseArtifacts: artifactToAdd}
		update["$inc"] = bson.M{
			StorageKeyReleaseArtifactsCount: 1,
		}
		update["$setOnInsert"] = bson.M{StorageKeyReleaseCreated: time.Now()}
	}
	_, err := collReleases.UpdateOne(
		ctx,
		bson.M{
			StorageKeyTenantId:    identity.FromContext(ctx).GetTenantID(),
			StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
			StorageKeyReleaseName: releaseName,
		},
		update,
		opt,
	)
	if err != nil {
		return err
	}
	if artifactToRemove != nil {
		r := collReleases.FindOneAndDelete(
			ctx,
			bson.M{
				StorageKeyReleaseKind:      bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
				StorageKeyReleaseName:      releaseName,
				StorageKeyReleaseArtifacts: bson.M{"$size": 0},
			},
		)
		if r.Err() != nil {
			return err
		}
	}
	return nil
}

func (db *DataStoreMongo) ListReleaseTags(ctx context.Context) (model.Tags, error) {
	filter := bson.M{
		StorageKeyTenantId: identity.FromContext(ctx).GetTenantID(),

		StorageKeyReleaseKind: bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}},
	}
	res := db.database.
		Collection(CollectionReleases).
		Distinct(ctx, StorageKeyReleaseTags, filter)

	var tagKeys []string
	if err := res.Decode(&tagKeys); err != nil {
		return nil, err
	}
	ret := make([]model.Tag, 0, len(tagKeys))
	for _, elem := range tagKeys {
		ret = append(ret, model.Tag(elem))
	}

	return ret, nil
}

func (db *DataStoreMongo) ReplaceReleaseTags(
	ctx context.Context,
	releaseName string,
	tags model.Tags,
) error {
	// Check preconditions
	if len(tags) > model.TagsMaxUnique {
		return model.ErrTooManyUniqueTags
	}

	collReleases := db.database.Collection(CollectionReleases)

	// Check if added tags will exceed limits
	if len(tags) > 0 {
		inUseTags, err := db.ListReleaseTags(ctx)
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
	var result []string
	err := c.Distinct(ctx, StorageKeyReleaseArtifactsUpdateTypes, bson.M{
		StorageKeyTenantId:                    identity.FromContext(ctx).GetTenantID(),
		StorageKeyReleaseArtifactsUpdateTypes: bson.M{"$type": "string"},
	}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []string{}, nil
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
		StorageKeyId:          releaseName,
	}).
		Decode(release); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return release, nil
}
