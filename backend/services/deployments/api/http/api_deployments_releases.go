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

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

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

func redactReleaseName(r *rest.Request) {
	q := r.URL.Query()
	if q.Get(ParamName) != "" {
		q.Set(ParamName, Redacted)
		r.URL.RawQuery = q.Encode()
	}
}

func (d *DeploymentsApiHandlers) GetReleases(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	defer redactReleaseName(r)
	filter := getReleaseOrImageFilter(r, listReleasesV1, false)
	releases, _, err := d.store.GetReleases(r.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	d.view.RenderSuccessGet(w, model.ConvertReleasesToV1(releases))
}

func (d *DeploymentsApiHandlers) listReleases(w rest.ResponseWriter, r *rest.Request,
	version listReleasesVersion) {
	l := requestlog.GetRequestLogger(r)

	defer redactReleaseName(r)
	filter := getReleaseOrImageFilter(r, version, true)
	releases, totalCount, err := d.store.GetReleases(r.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	hasNext := totalCount > int(filter.Page*filter.PerPage)
	links := rest_utils.MakePageLinkHdrs(r, uint64(filter.Page), uint64(filter.PerPage), hasNext)
	for _, l := range links {
		w.Header().Add("Link", l)
	}
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))

	if version == listReleasesV1 {
		d.view.RenderSuccessGet(w, model.ConvertReleasesToV1(releases))
	} else {
		d.view.RenderSuccessGet(w, releases)
	}
}

func (d *DeploymentsApiHandlers) ListReleases(w rest.ResponseWriter, r *rest.Request) {
	d.listReleases(w, r, listReleasesV1)
}

func (d *DeploymentsApiHandlers) ListReleasesV2(w rest.ResponseWriter, r *rest.Request) {
	d.listReleases(w, r, listReleasesV2)
}

func (d *DeploymentsApiHandlers) GetRelease(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	releaseName := r.PathParam(ParamName)
	if releaseName == "" {
		err := errors.New("path parameter 'release_name' cannot be empty")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusNotFound)
		return
	}

	release, err := d.app.GetRelease(ctx, releaseName)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrReleaseNotFound) {
			status = http.StatusNotFound
		}
		rest_utils.RestErrWithLog(w, r, l, err, status)
		return
	}

	d.view.RenderSuccessGet(w, release)
}

func (d *DeploymentsApiHandlers) PatchRelease(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	releaseName := r.PathParam(ParamName)
	if releaseName == "" {
		err := errors.New("path parameter 'release_name' cannot be empty")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusNotFound)
		return
	}

	var release model.ReleasePatch
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&release); err != nil {
		rest_utils.RestErrWithLog(w, r, l,
			errors.WithMessage(err,
				"malformed JSON in request body"),
			http.StatusBadRequest)
		return
	}
	if err := release.Validate(); err != nil {
		rest_utils.RestErrWithLog(w, r, l,
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
		rest_utils.RestErrWithLog(w, r, l, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DeploymentsApiHandlers) PutReleaseTags(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	releaseName := r.PathParam(ParamName)
	if releaseName == "" {
		err := errors.New("path parameter 'release_name' cannot be empty")
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusNotFound)
		return
	}

	var tags model.Tags
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&tags); err != nil {
		rest_utils.RestErrWithLog(w, r, l,
			errors.WithMessage(err,
				"malformed JSON in request body"),
			http.StatusBadRequest)
		return
	}
	if err := tags.Validate(); err != nil {
		rest_utils.RestErrWithLog(w, r, l,
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
		rest_utils.RestErrWithLog(w, r, l, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *DeploymentsApiHandlers) GetReleaseTagKeys(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	tags, err := d.app.ListReleaseTags(ctx)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = w.WriteJson(tags)
	if err != nil {
		l.Errorf("failed to serialize JSON response: %s", err.Error())
	}
}

func (d *DeploymentsApiHandlers) GetReleasesUpdateTypes(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	updateTypes, err := d.app.GetReleasesUpdateTypes(ctx)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = w.WriteJson(updateTypes)
	if err != nil {
		l.Errorf("failed to serialize JSON response: %s", err.Error())
	}
}

func (d *DeploymentsApiHandlers) DeleteReleases(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	names := r.URL.Query()[ParamName]

	if len(names) == 0 {
		rest_utils.RestErrWithLog(w, r, l,
			ErrReleaseNameNotProvided,
			http.StatusBadRequest)
		return
	}

	ids, err := d.app.DeleteReleases(ctx, names)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusInternalServerError)
		return
	}

	if len(ids) > 0 {
		w.WriteHeader(http.StatusConflict)
		deleteErr := model.ReleasesDeleteError{
			Error:             ErrReleaseUsedInActiveDeployment.Error(),
			RequestID:         requestid.GetReqId(r),
			ActiveDeployments: ids,
		}
		err = w.WriteJson(deleteErr)
		if err != nil {
			l.Errorf("failed to serialize JSON response: %s", err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
