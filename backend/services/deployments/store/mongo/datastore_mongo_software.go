package mongo

import (
	"context"
	"errors"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func (db *DataStoreMongo) ListSoftwareTags(
	ctx context.Context,
	filter *model.SoftwareTagsFilter,
) (model.Tags, error) {
	fltr := bson.M{StorageKeyTenantId: identity.FromContext(ctx).GetTenantID()}
	if filter != nil {
		switch filter.Kind {
		case model.ReleaseKindRelease:
			fltr[StorageKeyReleaseKind] = bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}}
		case model.ReleaseKindManifest:
			fltr[StorageKeyReleaseKind] = model.ReleaseKindManifest
		}
	}

	res := db.database.
		Collection(CollectionReleases).
		Distinct(ctx, StorageKeyReleaseTags, fltr)

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

func (db *DataStoreMongo) ListSoftware(
	ctx context.Context,
	filter *model.SoftwareFilter,
) ([]model.Software, int, error) {
	fltr := bson.M{StorageKeyTenantId: identity.FromContext(ctx).GetTenantID()}
	opts := mopts.Find()
	page := 1
	perPage := DefaultDocumentLimit
	if filter != nil {
		switch filter.Kind {
		case model.ReleaseKindRelease:
			fltr[StorageKeyReleaseKind] = bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}}
		case model.ReleaseKindManifest:
			fltr[StorageKeyReleaseKind] = model.ReleaseKindManifest
		}

		if len(filter.Names) > 0 {
			fltr[StorageKeyReleaseName] = bson.M{"$in": filter.Names}
		} else if filter.NamePrefix != "" {
			fltr[StorageKeyReleaseName] = bson.M{
				"$regex": bson.Regex{
					Pattern: "^" + regexp.QuoteMeta(filter.NamePrefix),
				},
			}
		}
		if filter.UpdateType != "" {
			fltr[StorageKeyReleaseArtifactsUpdateTypes] = filter.UpdateType
		}

		if len(filter.Tags) > 0 {
			fltr[StorageKeyReleaseTags] = bson.M{"$in": filter.Tags}
		}

		sortField, sortOrder := getReleaseSortFieldAndOrder(filter.Sort)
		if sortField == "" || sortField == "name" {
			sortField = StorageKeyReleaseName
		}
		if sortOrder == 0 {
			sortOrder = 1
		}
		opts.SetSort(bson.D{{Key: sortField, Value: sortOrder}})

		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PerPage > 0 {
			perPage = filter.PerPage
		}
	}

	opts.SetSkip(int64((page - 1) * perPage))
	opts.SetLimit(int64(perPage))

	collReleases := db.database.Collection(CollectionReleases)

	count, err := collReleases.CountDocuments(ctx, fltr)
	if err != nil || count < 1 {
		return []model.Software{}, 0, err
	}

	softwares := []model.Software{}
	cursor, err := collReleases.Find(ctx, fltr, opts)
	if err != nil {
		return []model.Software{}, 0, err
	}
	if err := cursor.All(ctx, &softwares); err != nil {
		return []model.Software{}, 0, err
	}

	return fillMissingKind(softwares), int(count), nil
}

func fillMissingKind(softwares []model.Software) []model.Software {
	for i := range softwares {
		// no kind means release
		if softwares[i].Kind == "" {
			softwares[i].Kind = model.ReleaseKindRelease
		}
	}
	return softwares
}

func (db *DataStoreMongo) GetSoftware(ctx context.Context, name string) (*model.Software, error) {
	var sw model.Software
	err := db.database.Collection(CollectionReleases).FindOne(ctx, bson.D{
		{Key: StorageKeyTenantId, Value: identity.FromContext(ctx).GetTenantID()},
		{Key: StorageKeyReleaseName, Value: name},
	}).Decode(&sw)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return &sw, nil
}
