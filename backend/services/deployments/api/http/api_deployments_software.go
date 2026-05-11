package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

var (
	ErrSoftwareTagsFilterMultipleKinds = errors.New("only one kind query argument can be specified") // nolint:lll
)

func (d *DeploymentsApiHandlers) GetSoftwareTags(c *gin.Context) {
	kind := c.Request.URL.Query()[ParamReleaseKind]
	if len(kind) > 1 {
		d.view.RenderError(c,
			ErrSoftwareTagsFilterMultipleKinds,
			http.StatusBadRequest)
		return
	}

	var filter *model.SoftwareTagsFilter
	if len(kind) > 0 {
		filter = &model.SoftwareTagsFilter{
			Kind: model.ReleaseKind(kind[0]),
		}
		err := filter.Validate()
		if err != nil {
			d.view.RenderError(c, err, http.StatusBadRequest)
			return
		}
	}

	ctx := c.Request.Context()
	tags, err := d.app.ListSoftwareTags(ctx, filter)
	if err != nil {
		d.view.RenderError(c, err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, tags)
}
