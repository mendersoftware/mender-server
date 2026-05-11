package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

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
		ret = append(ret, model.Tag(elem))
	}

	return ret, nil
}
