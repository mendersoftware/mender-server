package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"
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

func getSoftwareFilter(c *gin.Context) *model.SoftwareFilter {
	filter := &model.SoftwareFilter{
		Names:      c.QueryArray("name"),
		NamePrefix: c.Query("name_prefix"),
		UpdateType: c.Query("update_type"),
		Sort:       c.DefaultQuery("sort", "name:asc"),
	}

	if kind := c.Query("kind"); kind != "" {
		filter.Kind = model.ReleaseKind(kind)
	}

	return filter

}

func (d *DeploymentsApiHandlers) ListSoftware(c *gin.Context) {
	ctx := c.Request.Context()
	filter := getSoftwareFilter(c)
	if err := filter.Validate(); err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	filter.Page = int(page)
	filter.PerPage = int(perPage)

	softwares, totalCount, err := d.app.ListSoftware(ctx, filter)
	if err != nil {
		d.view.RenderError(c, err, http.StatusInternalServerError)
		return
	}

	hasNext := totalCount > int(filter.Page*filter.PerPage)

	hints := rest.NewPagingHints().
		SetPage(int64(filter.Page)).
		SetPerPage(int64(filter.PerPage)).
		SetHasNext(hasNext).
		SetTotalCount(int64(totalCount))

	links, err := rest.MakePagingHeaders(c.Request, hints)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	for _, l := range links {
		c.Writer.Header().Add(hdrLink, l)
	}
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))

	d.view.RenderSuccessGet(c, softwares)

}
