// Copyright 2024 Northern.tech AS
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

package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type listReleasesVersion int64

const (
	listReleasesV1 listReleasesVersion = iota
	listReleasesV2
)

// Errors
var (
	ErrReleaseNameNotProvided        = errors.New("at least one release name has to be provided")
	ErrReleaseUsedInActiveDeployment = errors.New("release(s) used in active deployment")
)

func redactReleaseName(r *http.Request) {
	q := r.URL.Query()
	if q.Get(ParamName) != "" {
		q.Set(ParamName, Redacted)
		r.URL.RawQuery = q.Encode()
	}
}

func (d *DeploymentsApiHandlers) GetReleases(c *gin.Context) {

	defer redactReleaseName(c.Request)
	filter := getReleaseOrImageFilter(c.Request, listReleasesV1, false)
	releases, _, err := d.store.GetReleases(c.Request.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	d.view.RenderSuccessGet(c, model.ConvertReleasesToV1(releases))
}

func (d *DeploymentsApiHandlers) listReleases(c *gin.Context,
	version listReleasesVersion) {

	defer redactReleaseName(c.Request)
	filter := getReleaseOrImageFilter(c.Request, version, true)
	releases, totalCount, err := d.store.GetReleases(c.Request.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(c, err)
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
		c.Writer.Header().Add("Link", l)
	}
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))

	if version == listReleasesV1 {
		d.view.RenderSuccessGet(c, model.ConvertReleasesToV1(releases))
	} else {
		d.view.RenderSuccessGet(c, releases)
	}
}

func (d *DeploymentsApiHandlers) ListReleases(c *gin.Context) {
	d.listReleases(c, listReleasesV1)
}

func (d *DeploymentsApiHandlers) ListReleasesV2(c *gin.Context) {
	d.listReleases(c, listReleasesV2)
}

func (d *DeploymentsApiHandlers) GetRelease(c *gin.Context) {
	ctx := c.Request.Context()

	releaseName := c.Param(ParamName)
	if releaseName == "" {
		err := errors.New("path parameter 'release_name' cannot be empty")
		d.view.RenderError(c, err, http.StatusNotFound)
		return
	}

	release, err := d.app.GetRelease(ctx, releaseName)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrReleaseNotFound) {
			status = http.StatusNotFound
		}
		d.view.RenderError(c, err, status)
		return
	}

	d.view.RenderSuccessGet(c, release)
}

func (d *DeploymentsApiHandlers) PatchRelease(c *gin.Context) {
	ctx := c.Request.Context()

	releaseName := c.Param(ParamName)
	if releaseName == "" {
		err := errors.New("path parameter 'release_name' cannot be empty")
		d.view.RenderError(c, err, http.StatusNotFound)
		return
	}

	var release model.ReleasePatch
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&release); err != nil {
		d.view.RenderError(c,
			errors.WithMessage(err,
				"malformed JSON in request body"),
			http.StatusBadRequest)
		return
	}
	if err := release.Validate(); err != nil {
		d.view.RenderError(c,
			errors.WithMessage(err,
				"invalid request body"),
			http.StatusBadRequest)
		return
	}

	err := d.app.UpdateRelease(ctx, releaseName, release)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrReleaseNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, model.ErrTooManyUniqueTags) {
			status = http.StatusConflict
		}
		d.view.RenderError(c, err, status)
		return
	}

	c.Status(http.StatusNoContent)
}

func (d *DeploymentsApiHandlers) PutReleaseTags(
	c *gin.Context,
) {

	ctx := c.Request.Context()

	releaseName := c.Param(ParamName)
	if releaseName == "" {
		err := errors.New("path parameter 'release_name' cannot be empty")
		d.view.RenderError(c, err, http.StatusNotFound)
		return
	}

	var tags model.Tags
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&tags); err != nil {
		d.view.RenderError(c,
			errors.WithMessage(err,
				"malformed JSON in request body"),
			http.StatusBadRequest)
		return
	}
	if err := tags.Validate(); err != nil {
		d.view.RenderError(c,
			errors.WithMessage(err,
				"invalid request body"),
			http.StatusBadRequest)
		return
	}

	err := d.app.ReplaceReleaseTags(ctx, releaseName, tags)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrReleaseNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, model.ErrTooManyUniqueTags) {
			status = http.StatusConflict
		}
		d.view.RenderError(c, err, status)
		return
	}

	c.Status(http.StatusNoContent)
}

func (d *DeploymentsApiHandlers) GetReleaseTagKeys(
	c *gin.Context,
) {
	ctx := c.Request.Context()

	tags, err := d.app.ListReleaseTags(ctx)
	if err != nil {
		d.view.RenderError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, tags)
}

func (d *DeploymentsApiHandlers) GetReleasesUpdateTypes(
	c *gin.Context,
) {
	ctx := c.Request.Context()

	updateTypes, err := d.app.GetReleasesUpdateTypes(ctx)
	if err != nil {
		d.view.RenderError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, updateTypes)
}

func (d *DeploymentsApiHandlers) DeleteReleases(
	c *gin.Context,
) {
	ctx := c.Request.Context()

	names := c.Request.URL.Query()[ParamName]

	if len(names) == 0 {
		d.view.RenderError(c,
			ErrReleaseNameNotProvided,
			http.StatusBadRequest)
		return
	}

	ids, err := d.app.DeleteReleases(ctx, names)
	if err != nil {
		d.view.RenderError(c, err, http.StatusInternalServerError)
		return
	}

	if len(ids) > 0 {
		deleteErr := model.ReleasesDeleteError{
			Error:             ErrReleaseUsedInActiveDeployment.Error(),
			RequestID:         requestid.GetReqId(c.Request),
			ActiveDeployments: ids,
		}
		c.JSON(http.StatusConflict, deleteErr)
		return
	}

	c.Status(http.StatusNoContent)
}
