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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
	"github.com/mendersoftware/mender-server/pkg/rest_utils"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/utils"
)

func init() {
	rest.ErrorFieldName = "error"
}

const (
	// 15 minutes
	DefaultDownloadLinkExpire = 15 * time.Minute
	// 10 Mb
	DefaultMaxMetaSize         = 1024 * 1024 * 10
	DefaultMaxImageSize        = 10 * 1024 * 1024 * 1024 // 10GiB
	DefaultMaxGenerateDataSize = 512 * 1024 * 1024       // 512MiB

	// Pagination
	DefaultPerPage                      = 20
	MaximumPerPage                      = 500
	MaximumPerPageListDeviceDeployments = 20
)

const (
	// Header Constants
	hdrTotalCount    = "X-Total-Count"
	hdrLink          = "Link"
	hdrForwardedHost = "X-Forwarded-Host"
)

// storage keys
const (
	// Common HTTP form parameters
	ParamArtifactName = "artifact_name"
	ParamDeviceType   = "device_type"
	ParamUpdateType   = "update_type"
	ParamDeploymentID = "deployment_id"
	ParamDeviceID     = "device_id"
	ParamTenantID     = "tenant_id"
	ParamName         = "name"
	ParamTag          = "tag"
	ParamDescription  = "description"
	ParamPage         = "page"
	ParamPerPage      = "per_page"
	ParamSort         = "sort"
	ParamID           = "id"
)

const Redacted = "REDACTED"

// JWT token
const (
	HTTPHeaderAuthorization       = "Authorization"
	HTTPHeaderAuthorizationBearer = "Bearer"
)

const (
	defaultTimeout = time.Second * 10
)

// Errors
var (
	ErrIDNotUUID                      = errors.New("ID is not a valid UUID")
	ErrEmptyID                        = errors.New("id: cannot be blank")
	ErrArtifactUsedInActiveDeployment = errors.New("Artifact is used in active deployment")
	ErrInvalidExpireParam             = errors.New("Invalid expire parameter")
	ErrArtifactNameMissing            = errors.New(
		"request does not contain the name of the artifact",
	)
	ErrArtifactTypeMissing = errors.New(
		"request does not contain the type of artifact",
	)
	ErrArtifactDeviceTypesCompatibleMissing = errors.New(
		"request does not contain the list of compatible device types",
	)
	ErrArtifactFileMissing       = errors.New("request does not contain the artifact file")
	ErrModelArtifactFileTooLarge = errors.New("Artifact file too large")

	ErrInternal                   = errors.New("Internal error")
	ErrDeploymentAlreadyFinished  = errors.New("Deployment already finished")
	ErrUnexpectedDeploymentStatus = errors.New("Unexpected deployment status")
	ErrMissingIdentity            = errors.New("Missing identity data")
	ErrMissingSize                = errors.New("missing size form-data")
	ErrMissingGroupName           = errors.New("Missing group name")

	ErrInvalidSortDirection = fmt.Errorf("invalid form value: must be one of \"%s\" or \"%s\"",
		model.SortDirectionAscending, model.SortDirectionDescending)
)

type Config struct {
	// URL signing parameters:

	// PresignSecret holds the secret value used by the signature algorithm.
	PresignSecret []byte
	// PresignExpire duration until the link expires.
	PresignExpire time.Duration
	// PresignHostname is the signed url hostname.
	PresignHostname string
	// PresignScheme is the URL scheme used for generating signed URLs.
	PresignScheme string
	// MaxImageSize is the maximum image size
	MaxImageSize        int64
	MaxGenerateDataSize int64

	EnableDirectUpload bool
	// EnableDirectUploadSkipVerify allows turning off the verification of uploaded artifacts
	EnableDirectUploadSkipVerify bool

	// DisableNewReleasesFeature is a flag that turns off the new API end-points
	// related to releases; helpful in performing long-running maintenance and data
	// migrations on the artifacts and releases collections.
	DisableNewReleasesFeature bool
}

func NewConfig() *Config {
	return &Config{
		PresignExpire:       DefaultDownloadLinkExpire,
		PresignScheme:       "https",
		MaxImageSize:        DefaultMaxImageSize,
		MaxGenerateDataSize: DefaultMaxGenerateDataSize,
	}
}

func (conf *Config) SetPresignSecret(key []byte) *Config {
	conf.PresignSecret = key
	return conf
}

func (conf *Config) SetPresignExpire(duration time.Duration) *Config {
	conf.PresignExpire = duration
	return conf
}

func (conf *Config) SetPresignHostname(hostname string) *Config {
	conf.PresignHostname = hostname
	return conf
}

func (conf *Config) SetPresignScheme(scheme string) *Config {
	conf.PresignScheme = scheme
	return conf
}

func (conf *Config) SetMaxImageSize(size int64) *Config {
	conf.MaxImageSize = size
	return conf
}

func (conf *Config) SetMaxGenerateDataSize(size int64) *Config {
	conf.MaxGenerateDataSize = size
	return conf
}

func (conf *Config) SetEnableDirectUpload(enable bool) *Config {
	conf.EnableDirectUpload = enable
	return conf
}

func (conf *Config) SetEnableDirectUploadSkipVerify(enable bool) *Config {
	conf.EnableDirectUploadSkipVerify = enable
	return conf
}

func (conf *Config) SetDisableNewReleasesFeature(disable bool) *Config {
	conf.DisableNewReleasesFeature = disable
	return conf
}

type DeploymentsApiHandlers struct {
	view   RESTView
	store  store.DataStore
	app    app.App
	config Config
}

func NewDeploymentsApiHandlers(
	store store.DataStore,
	view RESTView,
	app app.App,
	config ...*Config,
) *DeploymentsApiHandlers {
	conf := NewConfig()
	for _, c := range config {
		if c == nil {
			continue
		}
		if c.PresignSecret != nil {
			conf.PresignSecret = c.PresignSecret
		}
		if c.PresignExpire != 0 {
			conf.PresignExpire = c.PresignExpire
		}
		if c.PresignHostname != "" {
			conf.PresignHostname = c.PresignHostname
		}
		if c.PresignScheme != "" {
			conf.PresignScheme = c.PresignScheme
		}
		if c.MaxImageSize > 0 {
			conf.MaxImageSize = c.MaxImageSize
		}
		if c.MaxGenerateDataSize > 0 {
			conf.MaxGenerateDataSize = c.MaxGenerateDataSize
		}
		conf.DisableNewReleasesFeature = c.DisableNewReleasesFeature
		conf.EnableDirectUpload = c.EnableDirectUpload
		conf.EnableDirectUploadSkipVerify = c.EnableDirectUploadSkipVerify
	}
	return &DeploymentsApiHandlers{
		store:  store,
		view:   view,
		app:    app,
		config: *conf,
	}
}

func (d *DeploymentsApiHandlers) AliveHandler(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (d *DeploymentsApiHandlers) HealthHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := d.app.HealthCheck(ctx)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getReleaseOrImageFilter(r *rest.Request, version listReleasesVersion,
	paginated bool) *model.ReleaseOrImageFilter {

	q := r.URL.Query()

	filter := &model.ReleaseOrImageFilter{
		Name:       q.Get(ParamName),
		UpdateType: q.Get(ParamUpdateType),
	}
	if version == listReleasesV1 {
		filter.Description = q.Get(ParamDescription)
		filter.DeviceType = q.Get(ParamDeviceType)
	} else if version == listReleasesV2 {
		filter.Tags = q[ParamTag]
		for i, t := range filter.Tags {
			filter.Tags[i] = strings.ToLower(t)
		}
	}

	if paginated {
		filter.Sort = q.Get(ParamSort)
		if page := q.Get(ParamPage); page != "" {
			if i, err := strconv.Atoi(page); err == nil {
				filter.Page = i
			}
		}
		if perPage := q.Get(ParamPerPage); perPage != "" {
			if i, err := strconv.Atoi(perPage); err == nil {
				filter.PerPage = i
			}
		}
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.PerPage <= 0 || filter.PerPage > MaximumPerPage {
			filter.PerPage = DefaultPerPage
		}
	}

	return filter
}

type limitResponse struct {
	Limit uint64 `json:"limit"`
	Usage uint64 `json:"usage"`
}

func (d *DeploymentsApiHandlers) GetLimit(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	name := r.PathParam("name")

	if !model.IsValidLimit(name) {
		d.view.RenderError(w, r,
			errors.Errorf("unsupported limit %s", name),
			http.StatusBadRequest, l)
		return
	}

	limit, err := d.app.GetLimit(r.Context(), name)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	d.view.RenderSuccessGet(w, limitResponse{
		Limit: limit.Value,
		Usage: 0, // TODO fill this when ready
	})
}

// images

func (d *DeploymentsApiHandlers) GetImage(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	image, err := d.app.GetImage(r.Context(), id)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if image == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessGet(w, image)
}

func (d *DeploymentsApiHandlers) GetImages(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	defer redactReleaseName(r)
	filter := getReleaseOrImageFilter(r, listReleasesV1, false)

	list, _, err := d.app.ListImages(r.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	d.view.RenderSuccessGet(w, list)
}

func (d *DeploymentsApiHandlers) ListImages(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	defer redactReleaseName(r)
	filter := getReleaseOrImageFilter(r, listReleasesV1, true)

	list, totalCount, err := d.app.ListImages(r.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	hasNext := totalCount > int(filter.Page*filter.PerPage)
	links := rest_utils.MakePageLinkHdrs(r, uint64(filter.Page), uint64(filter.PerPage), hasNext)
	for _, l := range links {
		w.Header().Add(hdrLink, l)
	}
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))

	d.view.RenderSuccessGet(w, list)
}

func (d *DeploymentsApiHandlers) DownloadLink(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	expireSeconds := config.Config.GetInt(dconfig.SettingsStorageDownloadExpireSeconds)
	link, err := d.app.DownloadLink(r.Context(), id, time.Duration(expireSeconds)*time.Second)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if link == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessGet(w, link)
}

func (d *DeploymentsApiHandlers) UploadLink(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	expireSeconds := config.Config.GetInt(dconfig.SettingsStorageUploadExpireSeconds)
	link, err := d.app.UploadLink(
		r.Context(),
		time.Duration(expireSeconds)*time.Second,
		d.config.EnableDirectUploadSkipVerify,
	)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if link == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessGet(w, link)
}

const maxMetadataSize = 2048

func (d *DeploymentsApiHandlers) CompleteUpload(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := log.FromContext(ctx)

	artifactID := r.PathParam(ParamID)

	var metadata *model.DirectUploadMetadata
	if d.config.EnableDirectUploadSkipVerify {
		var directMetadata model.DirectUploadMetadata
		bodyBuffer := make([]byte, maxMetadataSize)
		n, err := io.ReadFull(r.Body, bodyBuffer)
		r.Body.Close()
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			l.Errorf("error reading post body data: %s (read: %d)", err.Error(), n)
		} else {
			err = json.Unmarshal(bodyBuffer[:n], &directMetadata)
			if err == nil {
				if directMetadata.Validate() == nil {
					metadata = &directMetadata
				}
			} else {
				l.Errorf("error parsing json data: %s", err.Error())
			}
		}
	}

	err := d.app.CompleteUpload(ctx, artifactID, d.config.EnableDirectUploadSkipVerify, metadata)
	switch errors.Cause(err) {
	case nil:
		// w.Header().Set("Link", "FEAT: Upload status API")
		w.WriteHeader(http.StatusAccepted)
	case app.ErrUploadNotFound:
		d.view.RenderErrorNotFound(w, r, l)
	default:
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteJson(rest_utils.ApiError{ // nolint:errcheck
			Err:   "internal server error",
			ReqId: requestid.FromContext(ctx),
		})
	}
}

func (d *DeploymentsApiHandlers) DownloadConfiguration(w rest.ResponseWriter, r *rest.Request) {
	if d.config.PresignSecret == nil {
		rest.NotFound(w, r)
		return
	}
	var (
		deviceID, _     = url.PathUnescape(r.PathParam(ParamDeviceID))
		deviceType, _   = url.PathUnescape(r.PathParam(ParamDeviceType))
		deploymentID, _ = url.PathUnescape(r.PathParam(ParamDeploymentID))
	)
	if deviceID == "" || deviceType == "" || deploymentID == "" {
		rest.NotFound(w, r)
		return
	}

	var (
		tenantID string
		l        = log.FromContext(r.Context())
		q        = r.URL.Query()
		err      error
	)
	tenantID = q.Get(ParamTenantID)
	sig := model.NewRequestSignature(r.Request, d.config.PresignSecret)
	if err = sig.Validate(); err != nil {
		switch cause := errors.Cause(err); cause {
		case model.ErrLinkExpired:
			d.view.RenderError(w, r, cause, http.StatusForbidden, l)
		default:
			d.view.RenderError(w, r,
				errors.Wrap(err, "invalid request parameters"),
				http.StatusBadRequest, l,
			)
		}
		return
	}

	if !sig.VerifyHMAC256() {
		d.view.RenderError(w, r,
			errors.New("signature invalid"),
			http.StatusForbidden, l,
		)
		return
	}

	// Validate request signature
	ctx := identity.WithContext(r.Context(), &identity.Identity{
		Subject:  deviceID,
		Tenant:   tenantID,
		IsDevice: true,
	})

	artifact, err := d.app.GenerateConfigurationImage(ctx, deviceType, deploymentID)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrModelDeploymentNotFound:
			d.view.RenderError(w, r,
				errors.Errorf(
					"deployment with id '%s' not found",
					deploymentID,
				),
				http.StatusNotFound, l,
			)
		default:
			l.Error(err.Error())
			d.view.RenderInternalError(w, r, err, l)
		}
		return
	}
	artifactPayload, err := io.ReadAll(artifact)
	if err != nil {
		l.Error(err.Error())
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	rw := w.(http.ResponseWriter)
	hdr := rw.Header()
	hdr.Set("Content-Disposition", `attachment; filename="artifact.mender"`)
	hdr.Set("Content-Type", app.ArtifactContentType)
	hdr.Set("Content-Length", strconv.Itoa(len(artifactPayload)))
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(artifactPayload)
	if err != nil {
		// There's not anything we can do here in terms of the response.
		l.Error(err.Error())
	}
}

func (d *DeploymentsApiHandlers) DeleteImage(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	if err := d.app.DeleteImage(r.Context(), id); err != nil {
		switch err {
		default:
			d.view.RenderInternalError(w, r, err, l)
		case app.ErrImageMetaNotFound:
			d.view.RenderErrorNotFound(w, r, l)
		case app.ErrModelImageInActiveDeployment:
			d.view.RenderError(w, r, ErrArtifactUsedInActiveDeployment, http.StatusConflict, l)
		}
		return
	}

	d.view.RenderSuccessDelete(w)
}

func (d *DeploymentsApiHandlers) EditImage(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	constructor, err := getImageMetaFromBody(r)
	if err != nil {
		d.view.RenderError(
			w,
			r,
			errors.Wrap(err, "Validating request body"),
			http.StatusBadRequest,
			l,
		)
		return
	}

	found, err := d.app.EditImage(r.Context(), id, constructor)
	if err != nil {
		if err == app.ErrModelImageUsedInAnyDeployment {
			d.view.RenderError(w, r, err, http.StatusUnprocessableEntity, l)
			return
		}
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if !found {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessPut(w)
}

func getImageMetaFromBody(r *rest.Request) (*model.ImageMeta, error) {

	var constructor *model.ImageMeta

	if err := r.DecodeJsonPayload(&constructor); err != nil {
		return nil, err
	}

	if err := constructor.Validate(); err != nil {
		return nil, err
	}

	return constructor, nil
}

// NewImage is the Multipart Image/Meta upload handler.
// Request should be of type "multipart/form-data". The parts are
// key/value pairs of metadata information except the last one,
// which must contain the artifact file.
func (d *DeploymentsApiHandlers) NewImage(w rest.ResponseWriter, r *rest.Request) {
	d.newImageWithContext(r.Context(), w, r)
}

func (d *DeploymentsApiHandlers) NewImageForTenantHandler(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	tenantID := r.PathParam("tenant")

	if tenantID == "" {
		rest_utils.RestErrWithLog(
			w,
			r,
			l,
			fmt.Errorf("missing tenant id in path"),
			http.StatusBadRequest,
		)
		return
	}

	var ctx context.Context
	if tenantID != "default" {
		ident := &identity.Identity{Tenant: tenantID}
		ctx = identity.WithContext(r.Context(), ident)
	} else {
		ctx = r.Context()
	}

	d.newImageWithContext(ctx, w, r)
}

func (d *DeploymentsApiHandlers) newImageWithContext(
	ctx context.Context,
	w rest.ResponseWriter,
	r *rest.Request,
) {
	l := requestlog.GetRequestLogger(r)

	formReader, err := r.MultipartReader()
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	// parse multipart message
	multipartUploadMsg, err := d.ParseMultipart(formReader)

	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	imgID, err := d.app.CreateImage(ctx, multipartUploadMsg)
	if err == nil {
		d.view.RenderSuccessPost(w, r, imgID)
		return
	}
	var cErr *model.ConflictError
	if errors.As(err, &cErr) {
		w.WriteHeader(http.StatusConflict)
		_ = cErr.WithRequestID(requestid.FromContext(ctx))
		err = w.WriteJson(cErr)
		if err != nil {
			l.Error(err)
		} else {
			l.Error(cErr.Error())
		}
		return
	}
	cause := errors.Cause(err)
	switch cause {
	default:
		d.view.RenderInternalError(w, r, err, l)
		return
	case app.ErrModelArtifactNotUnique:
		l.Error(err.Error())
		d.view.RenderError(w, r, cause, http.StatusUnprocessableEntity, l)
		return
	case app.ErrModelParsingArtifactFailed:
		l.Error(err.Error())
		d.view.RenderError(w, r, formatArtifactUploadError(err), http.StatusBadRequest, l)
		return
	case utils.ErrStreamTooLarge, ErrModelArtifactFileTooLarge:
		d.view.RenderError(w, r, ErrModelArtifactFileTooLarge, http.StatusRequestEntityTooLarge, l)
		return
	case app.ErrModelMissingInputMetadata, app.ErrModelMissingInputArtifact,
		app.ErrModelInvalidMetadata, app.ErrModelMultipartUploadMsgMalformed,
		io.ErrUnexpectedEOF:
		l.Error(err.Error())
		d.view.RenderError(w, r, cause, http.StatusBadRequest, l)
		return
	}
}

func formatArtifactUploadError(err error) error {
	// remove generic message
	errMsg := strings.TrimSuffix(err.Error(), ": "+app.ErrModelParsingArtifactFailed.Error())

	// handle specific cases

	if strings.Contains(errMsg, "invalid checksum") {
		return errors.New(errMsg[strings.Index(errMsg, "invalid checksum"):])
	}

	if strings.Contains(errMsg, "unsupported version") {
		return errors.New(errMsg[strings.Index(errMsg, "unsupported version"):] +
			"; supported versions are: 1, 2")
	}

	return errors.New(errMsg)
}

// GenerateImage s the multipart Raw Data/Meta upload handler.
// Request should be of type "multipart/form-data". The parts are
// key/valyue pairs of metadata information except the last one,
// which must contain the file containing the raw data to be processed
// into an artifact.
func (d *DeploymentsApiHandlers) GenerateImage(w rest.ResponseWriter, r *rest.Request) {
	l := requestlog.GetRequestLogger(r)

	formReader, err := r.MultipartReader()
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	// parse multipart message
	multipartMsg, err := d.ParseGenerateImageMultipart(formReader)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	tokenFields := strings.Fields(r.Header.Get("Authorization"))
	if len(tokenFields) == 2 && strings.EqualFold(tokenFields[0], "Bearer") {
		multipartMsg.Token = tokenFields[1]
	}

	imgID, err := d.app.GenerateImage(r.Context(), multipartMsg)
	cause := errors.Cause(err)
	switch cause {
	default:
		d.view.RenderInternalError(w, r, err, l)
	case nil:
		d.view.RenderSuccessPost(w, r, imgID)
	case app.ErrModelArtifactNotUnique:
		l.Error(err.Error())
		d.view.RenderError(w, r, cause, http.StatusUnprocessableEntity, l)
	case app.ErrModelParsingArtifactFailed:
		l.Error(err.Error())
		d.view.RenderError(w, r, formatArtifactUploadError(err), http.StatusBadRequest, l)
	case utils.ErrStreamTooLarge, ErrModelArtifactFileTooLarge:
		d.view.RenderError(w, r, ErrModelArtifactFileTooLarge, http.StatusRequestEntityTooLarge, l)
	case app.ErrModelMissingInputMetadata, app.ErrModelMissingInputArtifact,
		app.ErrModelInvalidMetadata, app.ErrModelMultipartUploadMsgMalformed,
		io.ErrUnexpectedEOF:
		l.Error(err.Error())
		d.view.RenderError(w, r, cause, http.StatusBadRequest, l)
	}
}

// ParseMultipart parses multipart/form-data message.
func (d *DeploymentsApiHandlers) ParseMultipart(
	r *multipart.Reader,
) (*model.MultipartUploadMsg, error) {
	uploadMsg := &model.MultipartUploadMsg{
		MetaConstructor: &model.ImageMeta{},
	}
	var size int64
	// Parse the multipart form sequentially. To remain backward compatible
	// all form names that are not part of the API are ignored.
	for {
		part, err := r.NextPart()
		if err != nil {
			if err == io.EOF {
				// The whole message has been consumed without
				// the "artifact" form part.
				return nil, ErrArtifactFileMissing
			}
			return nil, err
		}
		switch strings.ToLower(part.FormName()) {
		case "description":
			// Add description to the metadata
			dscr, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			uploadMsg.MetaConstructor.Description = string(dscr)

		case "size":
			// Add size limit to the metadata
			sz, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			size, err = strconv.ParseInt(string(sz), 10, 64)
			if err != nil {
				return nil, err
			}
			if size > d.config.MaxImageSize {
				return nil, ErrModelArtifactFileTooLarge
			}

		case "artifact_id":
			// Add artifact id to the metadata (must be a valid UUID).
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			id := string(b)
			if !govalidator.IsUUID(id) {
				return nil, errors.New(
					"artifact_id is not a valid UUID",
				)
			}
			uploadMsg.ArtifactID = id

		case "artifact":
			// Assign the form-data payload to the artifact reader
			// and return. The content is consumed elsewhere.
			if size > 0 {
				uploadMsg.ArtifactReader = utils.ReadExactly(part, size)
			} else {
				uploadMsg.ArtifactReader = utils.ReadAtMost(
					part,
					d.config.MaxImageSize,
				)
			}
			return uploadMsg, nil

		default:
			// Ignore all non-API sections.
			continue
		}
	}
}

// ParseGenerateImageMultipart parses multipart/form-data message.
func (d *DeploymentsApiHandlers) ParseGenerateImageMultipart(
	r *multipart.Reader,
) (*model.MultipartGenerateImageMsg, error) {
	msg := &model.MultipartGenerateImageMsg{}
	var size int64

ParseLoop:
	for {
		part, err := r.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch strings.ToLower(part.FormName()) {
		case "args":
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'args'",
				)
			}
			msg.Args = string(b)

		case "description":
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'description'",
				)
			}
			msg.Description = string(b)

		case "device_types_compatible":
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'device_types_compatible'",
				)
			}
			msg.DeviceTypesCompatible = strings.Split(string(b), ",")

		case "file":
			if size > 0 {
				msg.FileReader = utils.ReadExactly(part, size)
			} else {
				msg.FileReader = utils.ReadAtMost(part, d.config.MaxGenerateDataSize)
			}
			break ParseLoop

		case "name":
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'name'",
				)
			}
			msg.Name = string(b)

		case "type":
			b, err := io.ReadAll(part)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'type'",
				)
			}
			msg.Type = string(b)

		case "size":
			// Add size limit to the metadata
			sz, err := io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			size, err = strconv.ParseInt(string(sz), 10, 64)
			if err != nil {
				return nil, err
			}
			if size > d.config.MaxGenerateDataSize {
				return nil, ErrModelArtifactFileTooLarge
			}

		default:
			// Ignore non-API sections.
			continue
		}
	}

	return msg, errors.Wrap(msg.Validate(), "api: invalid form parameters")
}

// deployments
func (d *DeploymentsApiHandlers) createDeployment(
	w rest.ResponseWriter,
	r *rest.Request,
	ctx context.Context,
	l *log.Logger,
	group string,
) {
	constructor, err := d.getDeploymentConstructorFromBody(r, group)
	if err != nil {
		d.view.RenderError(
			w,
			r,
			errors.Wrap(err, "Validating request body"),
			http.StatusBadRequest,
			l,
		)
		return
	}

	id, err := d.app.CreateDeployment(ctx, constructor)
	switch err {
	case nil:
		// in case of deployment to group remove "/group/{name}" from path before creating location
		// haeder
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/group/"+constructor.Group)
		d.view.RenderSuccessPost(w, r, id)
	case app.ErrNoArtifact:
		d.view.RenderError(w, r, err, http.StatusUnprocessableEntity, l)
	case app.ErrNoDevices:
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
	case app.ErrConflictingDeployment:
		d.view.RenderError(w, r, err, http.StatusConflict, l)
	default:
		d.view.RenderInternalError(w, r, err, l)
	}
}

func (d *DeploymentsApiHandlers) PostDeployment(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	d.createDeployment(w, r, ctx, l, "")
}

func (d *DeploymentsApiHandlers) DeployToGroup(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	group := r.PathParam("name")
	if len(group) < 1 {
		d.view.RenderError(w, r, ErrMissingGroupName, http.StatusBadRequest, l)
	}
	d.createDeployment(w, r, ctx, l, group)
}

// parseDeviceConfigurationDeploymentPathParams parses expected params
// and check if the params are not empty
func parseDeviceConfigurationDeploymentPathParams(r *rest.Request) (string, string, string, error) {
	tenantID := r.PathParam("tenant")
	deviceID := r.PathParam(ParamDeviceID)
	if deviceID == "" {
		return "", "", "", errors.New("device ID missing")
	}
	deploymentID := r.PathParam(ParamDeploymentID)
	if deploymentID == "" {
		return "", "", "", errors.New("deployment ID missing")
	}
	return tenantID, deviceID, deploymentID, nil
}

// getConfigurationDeploymentConstructorFromBody extracts configuration
// deployment constructor from the request body and validates it
func getConfigurationDeploymentConstructorFromBody(r *rest.Request) (
	*model.ConfigurationDeploymentConstructor, error) {

	var constructor *model.ConfigurationDeploymentConstructor

	if err := r.DecodeJsonPayload(&constructor); err != nil {
		return nil, err
	}

	if err := constructor.Validate(); err != nil {
		return nil, err
	}

	return constructor, nil
}

// device configuration deployment handler
func (d *DeploymentsApiHandlers) PostDeviceConfigurationDeployment(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	l := requestlog.GetRequestLogger(r)

	// get path params
	tenantID, deviceID, deploymentID, err := parseDeviceConfigurationDeploymentPathParams(r)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	// add tenant id to the context
	ctx := identity.WithContext(r.Context(), &identity.Identity{Tenant: tenantID})

	constructor, err := getConfigurationDeploymentConstructorFromBody(r)
	if err != nil {
		d.view.RenderError(
			w,
			r,
			errors.Wrap(err, "Validating request body"),
			http.StatusBadRequest,
			l,
		)
		return
	}

	id, err := d.app.CreateDeviceConfigurationDeployment(ctx, constructor, deviceID, deploymentID)
	switch err {
	default:
		d.view.RenderInternalError(w, r, err, l)
	case nil:
		r.URL.Path = "./deployments"
		d.view.RenderSuccessPost(w, r, id)
	case app.ErrDuplicateDeployment:
		d.view.RenderError(w, r, err, http.StatusConflict, l)
	case app.ErrInvalidDeploymentID:
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
	}
}

func (d *DeploymentsApiHandlers) getDeploymentConstructorFromBody(
	r *rest.Request,
	group string,
) (*model.DeploymentConstructor, error) {
	var constructor *model.DeploymentConstructor
	if err := r.DecodeJsonPayload(&constructor); err != nil {
		return nil, err
	}

	constructor.Group = group

	if err := constructor.ValidateNew(); err != nil {
		return nil, err
	}

	return constructor, nil
}

func (d *DeploymentsApiHandlers) GetDeployment(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	deployment, err := d.app.GetDeployment(ctx, id)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if deployment == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessGet(w, deployment)
}

func (d *DeploymentsApiHandlers) GetDeploymentStats(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	stats, err := d.app.GetDeploymentStats(ctx, id)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if stats == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessGet(w, stats)
}

func (d *DeploymentsApiHandlers) GetDeploymentsStats(w rest.ResponseWriter, r *rest.Request) {

	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	ids := model.DeploymentIDs{}
	if err := r.DecodeJsonPayload(&ids); err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	if len(ids.IDs) == 0 {
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(struct{}{})
		return
	}

	if err := ids.Validate(); err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	stats, err := d.app.GetDeploymentsStats(ctx, ids.IDs...)
	if err != nil {
		if errors.Is(err, app.ErrModelDeploymentNotFound) {
			d.view.RenderError(w, r, err, http.StatusNotFound, l)
			return
		}
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	w.WriteHeader(http.StatusOK)

	_ = w.WriteJson(stats)
}

func (d *DeploymentsApiHandlers) GetDeploymentDeviceList(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	deployment, err := d.app.GetDeployment(ctx, id)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if deployment == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderSuccessGet(w, deployment.DeviceList)
}

func (d *DeploymentsApiHandlers) AbortDeployment(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	// receive request body
	var status struct {
		Status model.DeviceDeploymentStatus
	}

	err := r.DecodeJsonPayload(&status)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}
	// "aborted" is the only supported status
	if status.Status != model.DeviceDeploymentStatusAborted {
		d.view.RenderError(w, r, ErrUnexpectedDeploymentStatus, http.StatusBadRequest, l)
	}

	l.Infof("Abort deployment: %s", id)

	// Check if deployment is finished
	isDeploymentFinished, err := d.app.IsDeploymentFinished(ctx, id)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}
	if isDeploymentFinished {
		d.view.RenderError(w, r, ErrDeploymentAlreadyFinished, http.StatusUnprocessableEntity, l)
		return
	}

	// Abort deployments for devices and update deployment stats
	if err := d.app.AbortDeployment(ctx, id); err != nil {
		d.view.RenderInternalError(w, r, err, l)
	}

	d.view.RenderEmptySuccessResponse(w)
}

func (d *DeploymentsApiHandlers) GetDeploymentForDevice(w rest.ResponseWriter, r *rest.Request) {
	var (
		installed *model.InstalledDeviceDeployment
		ctx       = r.Context()
		l         = requestlog.GetRequestLogger(r)
		idata     = identity.FromContext(ctx)
	)
	if idata == nil {
		d.view.RenderError(w, r, ErrMissingIdentity, http.StatusBadRequest, l)
		return
	}

	q := r.URL.Query()
	defer func() {
		var reEncode bool = false
		if name := q.Get(ParamArtifactName); name != "" {
			q.Set(ParamArtifactName, Redacted)
			reEncode = true
		}
		if typ := q.Get(ParamDeviceType); typ != "" {
			q.Set(ParamDeviceType, Redacted)
			reEncode = true
		}
		if reEncode {
			r.URL.RawQuery = q.Encode()
		}
	}()
	if strings.EqualFold(r.Method, http.MethodPost) {
		// POST
		installed = new(model.InstalledDeviceDeployment)
		if err := r.DecodeJsonPayload(&installed); err != nil {
			d.view.RenderError(w, r,
				errors.Wrap(err, "invalid schema"),
				http.StatusBadRequest, l)
			return
		}
	} else {
		// GET or HEAD
		installed = &model.InstalledDeviceDeployment{
			ArtifactName: q.Get(ParamArtifactName),
			DeviceType:   q.Get(ParamDeviceType),
		}
	}

	if err := installed.Validate(); err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	request := &model.DeploymentNextRequest{
		DeviceProvides: installed,
	}

	d.getDeploymentForDevice(w, r, idata, request)
}

func (d *DeploymentsApiHandlers) getDeploymentForDevice(
	w rest.ResponseWriter,
	r *rest.Request,
	idata *identity.Identity,
	request *model.DeploymentNextRequest,
) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	deployment, err := d.app.GetDeploymentForDeviceWithCurrent(ctx, idata.Subject, request)
	if err != nil {
		if err == app.ErrConflictingRequestData {
			d.view.RenderError(w, r, err, http.StatusConflict, l)
		} else {
			d.view.RenderInternalError(w, r, err, l)
		}
		return
	}

	if deployment == nil {
		d.view.RenderNoUpdateForDevice(w)
		return
	} else if deployment.Type == model.DeploymentTypeConfiguration {
		// Generate pre-signed URL
		var hostName string = d.config.PresignHostname
		if hostName == "" {
			if hostName = r.Header.Get(hdrForwardedHost); hostName == "" {
				d.view.RenderInternalError(w, r,
					errors.New("presign.hostname not configured; "+
						"unable to generate download link "+
						" for configuration deployment"), l)
				return
			}
		}
		req, _ := http.NewRequest(
			http.MethodGet,
			FMTConfigURL(
				d.config.PresignScheme, hostName,
				deployment.ID, request.DeviceProvides.DeviceType,
				idata.Subject,
			),
			nil,
		)
		if idata.Tenant != "" {
			q := req.URL.Query()
			q.Set(model.ParamTenantID, idata.Tenant)
			req.URL.RawQuery = q.Encode()
		}
		sig := model.NewRequestSignature(req, d.config.PresignSecret)
		expireTS := time.Now().Add(d.config.PresignExpire)
		sig.SetExpire(expireTS)
		deployment.Artifact.Source = model.Link{
			Uri:    sig.PresignURL(),
			Expire: expireTS,
		}
	}

	d.view.RenderSuccessGet(w, deployment)
}

func (d *DeploymentsApiHandlers) PutDeploymentStatusForDevice(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	did := r.PathParam("id")

	idata := identity.FromContext(ctx)
	if idata == nil {
		d.view.RenderError(w, r, ErrMissingIdentity, http.StatusBadRequest, l)
		return
	}

	// receive request body
	var report model.StatusReport

	err := r.DecodeJsonPayload(&report)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	l.Infof("status: %+v", report)
	if err := d.app.UpdateDeviceDeploymentStatus(ctx, did,
		idata.Subject, model.DeviceDeploymentState{
			Status:   report.Status,
			SubState: report.SubState,
		}); err != nil {

		if err == app.ErrDeploymentAborted || err == app.ErrDeviceDecommissioned {
			d.view.RenderError(w, r, err, http.StatusConflict, l)
		} else if err == app.ErrStorageNotFound {
			d.view.RenderErrorNotFound(w, r, l)
		} else {
			d.view.RenderInternalError(w, r, err, l)
		}
		return
	}

	d.view.RenderEmptySuccessResponse(w)
}

func (d *DeploymentsApiHandlers) GetDeviceStatusesForDeployment(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	did := r.PathParam("id")

	if !govalidator.IsUUID(did) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	statuses, err := d.app.GetDeviceStatusesForDeployment(ctx, did)
	if err != nil {
		switch err {
		case app.ErrModelDeploymentNotFound:
			d.view.RenderError(w, r, err, http.StatusNotFound, l)
			return
		default:
			d.view.RenderInternalError(w, r, ErrInternal, l)
			return
		}
	}

	d.view.RenderSuccessGet(w, statuses)
}

func (d *DeploymentsApiHandlers) GetDevicesListForDeployment(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	did := r.PathParam("id")

	if !govalidator.IsUUID(did) {
		d.view.RenderError(w, r, ErrIDNotUUID, http.StatusBadRequest, l)
		return
	}

	page, perPage, err := rest_utils.ParsePagination(r)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	lq := store.ListQuery{
		Skip:         int((page - 1) * perPage),
		Limit:        int(perPage),
		DeploymentID: did,
	}
	if status := r.URL.Query().Get("status"); status != "" {
		lq.Status = &status
	}
	if err = lq.Validate(); err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	statuses, totalCount, err := d.app.GetDevicesListForDeployment(ctx, lq)
	if err != nil {
		switch err {
		case app.ErrModelDeploymentNotFound:
			d.view.RenderError(w, r, err, http.StatusNotFound, l)
			return
		default:
			d.view.RenderInternalError(w, r, ErrInternal, l)
			return
		}
	}

	hasNext := totalCount > int(page*perPage)
	links := rest_utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add(hdrLink, l)
	}
	w.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))
	d.view.RenderSuccessGet(w, statuses)
}

func ParseLookupQuery(vals url.Values) (model.Query, error) {
	query := model.Query{}

	createdBefore := vals.Get("created_before")
	if createdBefore != "" {
		if createdBeforeTime, err := parseEpochToTimestamp(createdBefore); err != nil {
			return query, errors.Wrap(err, "timestamp parsing failed for created_before parameter")
		} else {
			query.CreatedBefore = &createdBeforeTime
		}
	}

	createdAfter := vals.Get("created_after")
	if createdAfter != "" {
		if createdAfterTime, err := parseEpochToTimestamp(createdAfter); err != nil {
			return query, errors.Wrap(err, "timestamp parsing failed created_after parameter")
		} else {
			query.CreatedAfter = &createdAfterTime
		}
	}

	switch strings.ToLower(vals.Get("sort")) {
	case model.SortDirectionAscending:
		query.Sort = model.SortDirectionAscending
	case "", model.SortDirectionDescending:
		query.Sort = model.SortDirectionDescending
	default:
		return query, ErrInvalidSortDirection
	}

	status := vals.Get("status")
	switch status {
	case "inprogress":
		query.Status = model.StatusQueryInProgress
	case "finished":
		query.Status = model.StatusQueryFinished
	case "pending":
		query.Status = model.StatusQueryPending
	case "aborted":
		query.Status = model.StatusQueryAborted
	case "":
		query.Status = model.StatusQueryAny
	default:
		return query, errors.Errorf("unknown status %s", status)

	}

	dType := vals.Get("type")
	if dType == "" {
		return query, nil
	}
	deploymentType := model.DeploymentType(dType)
	if deploymentType == model.DeploymentTypeSoftware ||
		deploymentType == model.DeploymentTypeConfiguration {
		query.Type = deploymentType
	} else {
		return query, errors.Errorf("unknown deployment type %s", dType)
	}

	return query, nil
}

func ParseDeploymentLookupQueryV1(vals url.Values) (model.Query, error) {
	query, err := ParseLookupQuery(vals)
	if err != nil {
		return query, err
	}

	search := vals.Get("search")
	if search != "" {
		query.SearchText = search
	}

	return query, nil
}

func ParseDeploymentLookupQueryV2(vals url.Values) (model.Query, error) {
	query, err := ParseLookupQuery(vals)
	if err != nil {
		return query, err
	}

	query.Names = vals["name"]
	query.IDs = vals["id"]

	return query, nil
}

func parseEpochToTimestamp(epoch string) (time.Time, error) {
	if epochInt64, err := strconv.ParseInt(epoch, 10, 64); err != nil {
		return time.Time{}, errors.New("invalid timestamp: " + epoch)
	} else {
		return time.Unix(epochInt64, 0).UTC(), nil
	}
}

func (d *DeploymentsApiHandlers) LookupDeployment(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)
	q := r.URL.Query()
	defer func() {
		if search := q.Get("search"); search != "" {
			q.Set("search", Redacted)
			r.URL.RawQuery = q.Encode()
		}
	}()

	query, err := ParseDeploymentLookupQueryV1(q)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	page, perPage, err := rest_utils.ParsePagination(r)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}
	query.Skip = int((page - 1) * perPage)
	query.Limit = int(perPage + 1)

	deps, totalCount, err := d.app.LookupDeployment(ctx, query)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}
	w.Header().Add(hdrTotalCount, strconv.FormatInt(totalCount, 10))

	len := len(deps)
	hasNext := false
	if uint64(len) > perPage {
		hasNext = true
		len = int(perPage)
	}

	links := rest_utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add(hdrLink, l)
	}

	d.view.RenderSuccessGet(w, deps[:len])
}

func (d *DeploymentsApiHandlers) LookupDeploymentV2(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)
	q := r.URL.Query()
	defer func() {
		if q.Has("name") {
			q["name"] = []string{Redacted}
			r.URL.RawQuery = q.Encode()
		}
	}()

	query, err := ParseDeploymentLookupQueryV2(q)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	page, perPage, err := rest_utils.ParsePagination(r)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}
	query.Skip = int((page - 1) * perPage)
	query.Limit = int(perPage + 1)

	deps, totalCount, err := d.app.LookupDeployment(ctx, query)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}
	w.Header().Add(hdrTotalCount, strconv.FormatInt(totalCount, 10))

	len := len(deps)
	hasNext := false
	if uint64(len) > perPage {
		hasNext = true
		len = int(perPage)
	}

	links := rest_utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add(hdrLink, l)
	}

	d.view.RenderSuccessGet(w, deps[:len])
}

func (d *DeploymentsApiHandlers) PutDeploymentLogForDevice(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	did := r.PathParam("id")

	idata := identity.FromContext(ctx)
	if idata == nil {
		d.view.RenderError(w, r, ErrMissingIdentity, http.StatusBadRequest, l)
		return
	}

	// reuse DeploymentLog, device and deployment IDs are ignored when
	// (un-)marshaling DeploymentLog to/from JSON
	var log model.DeploymentLog

	err := r.DecodeJsonPayload(&log)
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	if err := d.app.SaveDeviceDeploymentLog(ctx, idata.Subject,
		did, log.Messages); err != nil {

		if err == app.ErrModelDeploymentNotFound {
			d.view.RenderError(w, r, err, http.StatusNotFound, l)
		} else {
			d.view.RenderInternalError(w, r, err, l)
		}
		return
	}

	d.view.RenderEmptySuccessResponse(w)
}

func (d *DeploymentsApiHandlers) GetDeploymentLogForDevice(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	did := r.PathParam("id")
	devid := r.PathParam("devid")

	depl, err := d.app.GetDeviceDeploymentLog(ctx, devid, did)

	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}

	if depl == nil {
		d.view.RenderErrorNotFound(w, r, l)
		return
	}

	d.view.RenderDeploymentLog(w, *depl)
}

func (d *DeploymentsApiHandlers) AbortDeviceDeployments(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")
	err := d.app.AbortDeviceDeployments(ctx, id)

	switch err {
	case nil, app.ErrStorageNotFound:
		d.view.RenderEmptySuccessResponse(w)
	default:
		d.view.RenderInternalError(w, r, err, l)
	}
}

func (d *DeploymentsApiHandlers) DeleteDeviceDeploymentsHistory(w rest.ResponseWriter,
	r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")
	err := d.app.DeleteDeviceDeploymentsHistory(ctx, id)

	switch err {
	case nil, app.ErrStorageNotFound:
		d.view.RenderEmptySuccessResponse(w)
	default:
		d.view.RenderInternalError(w, r, err, l)
	}
}

func (d *DeploymentsApiHandlers) ListDeviceDeployments(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	d.listDeviceDeployments(ctx, w, r, true)
}

func (d *DeploymentsApiHandlers) ListDeviceDeploymentsInternal(w rest.ResponseWriter,
	r *rest.Request) {
	ctx := r.Context()
	tenantID := r.PathParam("tenant")
	if tenantID != "" {
		ctx = identity.WithContext(r.Context(), &identity.Identity{
			Tenant:   tenantID,
			IsDevice: true,
		})
	}
	d.listDeviceDeployments(ctx, w, r, true)
}

func (d *DeploymentsApiHandlers) ListDeviceDeploymentsByIDsInternal(w rest.ResponseWriter,
	r *rest.Request) {
	ctx := r.Context()
	tenantID := r.PathParam("tenant")
	if tenantID != "" {
		ctx = identity.WithContext(r.Context(), &identity.Identity{
			Tenant:   tenantID,
			IsDevice: true,
		})
	}
	d.listDeviceDeployments(ctx, w, r, false)
}

func (d *DeploymentsApiHandlers) listDeviceDeployments(ctx context.Context,
	w rest.ResponseWriter, r *rest.Request, byDeviceID bool) {
	l := requestlog.GetRequestLogger(r)

	did := ""
	var IDs []string
	if byDeviceID {
		did = r.PathParam("id")
	} else {
		values := r.URL.Query()
		if values.Has("id") && len(values["id"]) > 0 {
			IDs = values["id"]
		} else {
			d.view.RenderError(w, r, ErrEmptyID, http.StatusBadRequest, l)
			return
		}
	}

	page, perPage, err := rest_utils.ParsePagination(r)
	if err == nil && perPage > MaximumPerPageListDeviceDeployments {
		err = errors.New(rest_utils.MsgQueryParmLimit(ParamPerPage))
	}
	if err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	lq := store.ListQueryDeviceDeployments{
		Skip:     int((page - 1) * perPage),
		Limit:    int(perPage),
		DeviceID: did,
		IDs:      IDs,
	}
	if status := r.URL.Query().Get("status"); status != "" {
		lq.Status = &status
	}
	if err = lq.Validate(); err != nil {
		d.view.RenderError(w, r, err, http.StatusBadRequest, l)
		return
	}

	deps, totalCount, err := d.app.GetDeviceDeploymentListForDevice(ctx, lq)
	if err != nil {
		d.view.RenderInternalError(w, r, err, l)
		return
	}
	w.Header().Add(hdrTotalCount, strconv.FormatInt(int64(totalCount), 10))

	hasNext := totalCount > lq.Skip+len(deps)
	links := rest_utils.MakePageLinkHdrs(r, page, perPage, hasNext)
	for _, l := range links {
		w.Header().Add(hdrLink, l)
	}

	d.view.RenderSuccessGet(w, deps)
}

func (d *DeploymentsApiHandlers) AbortDeviceDeploymentsInternal(w rest.ResponseWriter,
	r *rest.Request) {
	ctx := r.Context()
	tenantID := r.PathParam("tenantID")
	if tenantID != "" {
		ctx = identity.WithContext(r.Context(), &identity.Identity{
			Tenant:   tenantID,
			IsDevice: true,
		})
	}

	l := requestlog.GetRequestLogger(r)

	id := r.PathParam("id")

	// Decommission deployments for devices and update deployment stats
	err := d.app.DecommissionDevice(ctx, id)

	switch err {
	case nil, app.ErrStorageNotFound:
		d.view.RenderEmptySuccessResponse(w)
	default:
		d.view.RenderInternalError(w, r, err, l)

	}
}

// tenants

func (d *DeploymentsApiHandlers) ProvisionTenantsHandler(w rest.ResponseWriter, r *rest.Request) {
	ctx := r.Context()
	l := requestlog.GetRequestLogger(r)

	defer r.Body.Close()

	tenant, err := model.ParseNewTenantReq(r.Body)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = d.app.ProvisionTenant(ctx, tenant.TenantId)
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (d *DeploymentsApiHandlers) DeploymentsPerTenantHandler(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	tenantID := r.PathParam("tenant")
	if tenantID == "" {
		l := requestlog.GetRequestLogger(r)
		rest_utils.RestErrWithLog(w, r, l, errors.New("missing tenant ID"), http.StatusBadRequest)
		return
	}

	r.Request = r.WithContext(identity.WithContext(
		r.Context(),
		&identity.Identity{Tenant: tenantID},
	))
	d.LookupDeployment(w, r)
}

func (d *DeploymentsApiHandlers) GetTenantStorageSettingsHandler(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	l := requestlog.GetRequestLogger(r)

	tenantID := r.PathParam("tenant")

	ctx := identity.WithContext(
		r.Context(),
		&identity.Identity{Tenant: tenantID},
	)

	settings, err := d.app.GetStorageSettings(ctx)
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	d.view.RenderSuccessGet(w, settings)
}

func (d *DeploymentsApiHandlers) PutTenantStorageSettingsHandler(
	w rest.ResponseWriter,
	r *rest.Request,
) {
	l := requestlog.GetRequestLogger(r)

	defer r.Body.Close()

	tenantID := r.PathParam("tenant")

	ctx := identity.WithContext(
		r.Context(),
		&identity.Identity{Tenant: tenantID},
	)

	settings, err := model.ParseStorageSettingsRequest(r.Body)
	if err != nil {
		rest_utils.RestErrWithLog(w, r, l, err, http.StatusBadRequest)
		return
	}

	err = d.app.SetStorageSettings(ctx, settings)
	if err != nil {
		rest_utils.RestErrWithLogInternal(w, r, l, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
