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

func (d *Deployments) ListSoftware(
	ctx context.Context,
	filter *model.SoftwareFilter,
) ([]model.Software, int, error) {
	return d.db.ListSoftware(ctx, filter)
}
