package mongo

import (
	"context"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	mstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func (db *DataStoreMongo) ListSoftwareTags(
	ctx context.Context,
	filter *model.SoftwareTagsFilter,
) (model.Tags, error) {
	fltr := bson.M{}
	if filter != nil {
		switch filter.Kind {
		case model.ReleaseKindRelease:
			fltr[StorageKeyReleaseKind] = bson.M{"$in": bson.A{nil, model.ReleaseKindRelease}}
		case model.ReleaseKindManifest:
			fltr[StorageKeyReleaseKind] = model.ReleaseKindManifest
		}
	}

	res := db.client.
		Database(mstore.DbFromContext(ctx, DatabaseName)).
		Collection(CollectionReleases).
		Distinct(ctx, StorageKeyReleaseTags, fltr)

	var tagKeys []string
	if err := res.Decode(&tagKeys); err != nil {
		return nil, err
	}
	ret := make([]model.Tag, 0, len(tagKeys))
	for _, elem := range tagKeys {
		if elem != "" {
			ret = append(ret, model.Tag(elem))
		}
	}

	return ret, nil
}

func (db *DataStoreMongo) ListSoftware(
	ctx context.Context,
	filter *model.SoftwareFilter,
) ([]model.Software, int, error) {
	fltr := bson.M{}
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

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collReleases := database.Collection(CollectionReleases)

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
