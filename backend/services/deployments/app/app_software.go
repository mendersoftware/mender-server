package app

import (
	"context"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func (d *Deployments) ListSoftwareTags(
	ctx context.Context,
	filter *model.SoftwareTagsFilter,
) (model.Tags, error) {
	return d.db.ListSoftwareTags(ctx, filter)
}
