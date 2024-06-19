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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func (db *DataStoreMongo) UpdateReleaseArtifactDescription(
	ctx context.Context,
	artifactToEdit *model.Image,
	releaseName string,
) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collReleases := database.Collection(CollectionReleases)

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
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collReleases := database.Collection(CollectionReleases)

	opt := &mopts.UpdateOptions{}
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
		upsert := true
		opt.Upsert = &upsert
		update["$push"] = bson.M{StorageKeyReleaseArtifacts: artifactToAdd}
		update["$inc"] = bson.M{
			StorageKeyReleaseArtifactsCount: 1,
		}
	}
	_, err := collReleases.UpdateOne(
		ctx,
		bson.M{StorageKeyReleaseName: releaseName},
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
	l := log.FromContext(ctx)
	tagKeys, err := db.client.
		Database(mstore.DbFromContext(ctx, DatabaseName)).
		Collection(CollectionReleases).
		Distinct(ctx, StorageKeyReleaseTags, bson.D{})
	if err != nil {
		return nil, errors.WithMessage(err,
			"mongo: failed to retrieve distinct tags")
	}
	ret := make([]model.Tag, 0, len(tagKeys))
	for _, elem := range tagKeys {
		if key, ok := elem.(string); ok {
			ret = append(ret, model.Tag(key))
		} else {
			l.Warnf("unexpected data type (%T) received from distinct call: "+
				"ignoring result", elem)
		}
	}

	return ret, err
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

	collReleases := db.client.
		Database(mstore.DbFromContext(ctx, DatabaseName)).
		Collection(CollectionReleases)

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
	res, err := collReleases.UpdateOne(ctx, bson.D{{
		Key: StorageKeyReleaseName, Value: releaseName,
	}}, bson.D{{
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
	collReleases := db.client.
		Database(mstore.DbFromContext(ctx, DatabaseName)).
		Collection(CollectionReleases)

	err := release.Validate()
	if err != nil {
		return errors.Wrap(err, "cant update release due to validation errors")
	}

	// Update release, at the moment we update only the notes,
	// it is on purpose that we take only this field explicitly,
	// once there is a need we can extend
	res, err := collReleases.UpdateOne(
		ctx,
		bson.D{
			{
				Key: StorageKeyReleaseName, Value: releaseName,
			},
		},
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

// Save the possibly new update types
func (db *DataStoreMongo) SaveUpdateTypes(ctx context.Context, updateTypes []string) error {
	database := db.client.Database(DatabaseName)
	c := database.Collection(CollectionUpdateTypes)

	if len(updateTypes) < 1 {
		return nil
	}

	tenantId := ""
	if id := identity.FromContext(ctx); id != nil {
		tenantId = id.Tenant
	}
	options := mopts.UpdateOptions{}
	options.SetUpsert(true)
	_, err := c.UpdateOne(
		ctx,
		bson.M{
			StorageKeyTenantId: tenantId,
		},
		bson.M{
			"$addToSet": bson.M{
				StorageKeyStorageReleaseUpdateTypes: bson.M{
					"$each": updateTypes,
				},
			},
		},
		&options,
	)
	return err
}

// Get the update types
func (db *DataStoreMongo) GetUpdateTypes(ctx context.Context) ([]string, error) {
	database := db.client.Database(DatabaseName)
	c := database.Collection(CollectionUpdateTypes)

	tenantId := ""
	if id := identity.FromContext(ctx); id != nil {
		tenantId = id.Tenant
	}
	result := c.FindOne(
		ctx,
		bson.M{
			StorageKeyTenantId: tenantId,
		},
	)
	type updateType struct {
		UpdateTypes []string `bson:"update_types"`
	}
	var updateTypes updateType
	err := result.Decode(&updateTypes)
	if err == mongo.ErrNoDocuments {
		return []string{}, nil
	}
	if err != nil {
		return []string{}, err
	} else {
		return updateTypes.UpdateTypes, nil
	}
}

func (db *DataStoreMongo) DeleteReleasesByNames(ctx context.Context, names []string) error {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionReleases)
	query := bson.M{
		StorageKeyReleaseName: bson.M{
			"$in": names,
		},
	}
	_, err := collDevs.DeleteMany(ctx, query)
	return err
}
