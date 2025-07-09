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

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/utils"
)

const (
	// 15 minutes
	DefaultDownloadLinkExpire = 15 * time.Minute
	// 10 Mb
	MaxFormParamSize           = 1024 * 1024             // 1MiB
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

func (d *DeploymentsApiHandlers) AliveHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (d *DeploymentsApiHandlers) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := d.app.HealthCheck(ctx)
	if err != nil {
		d.view.RenderError(c, err, http.StatusServiceUnavailable)
		return
	}
	c.Status(http.StatusNoContent)
}

func getReleaseOrImageFilter(r *http.Request, version listReleasesVersion,
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

func (d *DeploymentsApiHandlers) GetLimit(c *gin.Context) {

	name := c.Param("name")

	if !model.IsValidLimit(name) {
		d.view.RenderError(c,
			errors.Errorf("unsupported limit %s", name),
			http.StatusBadRequest)
		return
	}

	limit, err := d.app.GetLimit(c.Request.Context(), name)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	d.view.RenderSuccessGet(c, limitResponse{
		Limit: limit.Value,
		Usage: 0, // TODO fill this when ready
	})
}

// images

func (d *DeploymentsApiHandlers) GetImage(c *gin.Context) {

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	image, err := d.app.GetImage(c.Request.Context(), id)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if image == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessGet(c, image)
}

func (d *DeploymentsApiHandlers) GetImages(c *gin.Context) {

	defer redactReleaseName(c.Request)
	filter := getReleaseOrImageFilter(c.Request, listReleasesV1, false)

	list, _, err := d.app.ListImages(c.Request.Context(), filter)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	d.view.RenderSuccessGet(c, list)
}

func (d *DeploymentsApiHandlers) ListImages(c *gin.Context) {

	defer redactReleaseName(c.Request)
	filter := getReleaseOrImageFilter(c.Request, listReleasesV1, true)

	list, totalCount, err := d.app.ListImages(c.Request.Context(), filter)
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
		c.Writer.Header().Add(hdrLink, l)
	}
	c.Writer.Header().Add(hdrTotalCount, strconv.Itoa(totalCount))

	d.view.RenderSuccessGet(c, list)
}

func (d *DeploymentsApiHandlers) DownloadLink(c *gin.Context) {

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	expireSeconds := config.Config.GetInt(dconfig.SettingsStorageDownloadExpireSeconds)
	link, err := d.app.DownloadLink(c.Request.Context(), id,
		time.Duration(expireSeconds)*time.Second)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if link == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessGet(c, link)
}

func (d *DeploymentsApiHandlers) UploadLink(c *gin.Context) {

	expireSeconds := config.Config.GetInt(dconfig.SettingsStorageUploadExpireSeconds)
	link, err := d.app.UploadLink(
		c.Request.Context(),
		time.Duration(expireSeconds)*time.Second,
		d.config.EnableDirectUploadSkipVerify,
	)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if link == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessGet(c, link)
}

const maxMetadataSize = 2048

func (d *DeploymentsApiHandlers) CompleteUpload(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)

	artifactID := c.Param(ParamID)

	var metadata *model.DirectUploadMetadata
	if d.config.EnableDirectUploadSkipVerify {
		var directMetadata model.DirectUploadMetadata
		bodyBuffer := make([]byte, maxMetadataSize)
		n, err := io.ReadFull(c.Request.Body, bodyBuffer)
		c.Request.Body.Close()
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
		// c.Writer.Header().Set("Link", "FEAT: Upload status API")
		c.Status(http.StatusAccepted)
	case app.ErrUploadNotFound:
		d.view.RenderErrorNotFound(c)
	default:
		d.view.RenderInternalError(c, err)
	}
}

func (d *DeploymentsApiHandlers) DownloadConfiguration(c *gin.Context) {
	if d.config.PresignSecret == nil {
		d.view.RenderErrorNotFound(c)
		return
	}
	var (
		deviceID, _     = url.PathUnescape(c.Param(ParamDeviceID))
		deviceType, _   = url.PathUnescape(c.Param(ParamDeviceType))
		deploymentID, _ = url.PathUnescape(c.Param(ParamDeploymentID))
	)
	if deviceID == "" || deviceType == "" || deploymentID == "" {
		d.view.RenderErrorNotFound(c)
		return
	}

	var (
		tenantID string
		q        = c.Request.URL.Query()
		err      error
	)
	tenantID = q.Get(ParamTenantID)
	sig := model.NewRequestSignature(c.Request, d.config.PresignSecret)
	if err = sig.Validate(); err != nil {
		switch cause := errors.Cause(err); cause {
		case model.ErrLinkExpired:
			d.view.RenderError(c, cause, http.StatusForbidden)
		default:
			d.view.RenderError(c,
				errors.Wrap(err, "invalid request parameters"),
				http.StatusBadRequest,
			)
		}
		return
	}

	if !sig.VerifyHMAC256() {
		d.view.RenderError(c,
			errors.New("signature invalid"),
			http.StatusForbidden,
		)
		return
	}

	// Validate request signature
	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Subject:  deviceID,
		Tenant:   tenantID,
		IsDevice: true,
	})

	artifact, err := d.app.GenerateConfigurationImage(ctx, deviceType, deploymentID)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrModelDeploymentNotFound:
			d.view.RenderError(c,
				errors.Errorf(
					"deployment with id '%s' not found",
					deploymentID,
				),
				http.StatusNotFound,
			)
		default:
			d.view.RenderInternalError(c, err)
		}
		return
	}
	artifactPayload, err := io.ReadAll(artifact)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	rw := c.Writer
	hdr := rw.Header()
	hdr.Set("Content-Disposition", `attachment; filename="artifact.mender"`)
	hdr.Set("Content-Type", app.ArtifactContentType)
	hdr.Set("Content-Length", strconv.Itoa(len(artifactPayload)))
	c.Status(http.StatusOK)
	_, err = rw.Write(artifactPayload)
	if err != nil {
		// There's not anything we can do here in terms of the response.
		_ = c.Error(err)
	}
}

func (d *DeploymentsApiHandlers) DeleteImage(c *gin.Context) {

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	if err := d.app.DeleteImage(c.Request.Context(), id); err != nil {
		switch err {
		default:
			d.view.RenderInternalError(c, err)
		case app.ErrImageMetaNotFound:
			d.view.RenderErrorNotFound(c)
		case app.ErrModelImageInActiveDeployment:
			d.view.RenderError(c, ErrArtifactUsedInActiveDeployment, http.StatusConflict)
		}
		return
	}

	d.view.RenderSuccessDelete(c)
}

func (d *DeploymentsApiHandlers) EditImage(c *gin.Context) {

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	constructor, err := getImageMetaFromBody(c)
	if err != nil {
		d.view.RenderError(
			c,
			errors.Wrap(err, "Validating request body"),
			http.StatusBadRequest,
		)
		return
	}

	found, err := d.app.EditImage(c.Request.Context(), id, constructor)
	if err != nil {
		if err == app.ErrModelImageUsedInAnyDeployment {
			d.view.RenderError(c, err, http.StatusUnprocessableEntity)
			return
		}
		d.view.RenderInternalError(c, err)
		return
	}

	if !found {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessPut(c)
}

func getImageMetaFromBody(c *gin.Context) (*model.ImageMeta, error) {

	var constructor *model.ImageMeta

	if err := c.ShouldBindJSON(&constructor); err != nil {
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
func (d *DeploymentsApiHandlers) NewImage(c *gin.Context) {
	d.newImageWithContext(c.Request.Context(), c)
}

func (d *DeploymentsApiHandlers) NewImageForTenantHandler(c *gin.Context) {

	tenantID := c.Param("tenant")

	if tenantID == "" {
		rest.RenderError(
			c,
			http.StatusBadRequest,
			fmt.Errorf("missing tenant id in path"),
		)
		return
	}

	var ctx context.Context
	if tenantID != "default" {
		ident := &identity.Identity{Tenant: tenantID}
		ctx = identity.WithContext(c.Request.Context(), ident)
	} else {
		ctx = c.Request.Context()
	}

	d.newImageWithContext(ctx, c)
}

func (d *DeploymentsApiHandlers) newImageWithContext(
	ctx context.Context,
	c *gin.Context,
) {

	formReader, err := c.Request.MultipartReader()
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	// parse multipart message
	multipartUploadMsg, err := d.ParseMultipart(formReader)

	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	imgID, err := d.app.CreateImage(ctx, multipartUploadMsg)
	if err == nil {
		d.view.RenderSuccessPost(c, imgID)
		return
	}
	var cErr *model.ConflictError
	if errors.As(err, &cErr) {
		_ = cErr.WithRequestID(requestid.FromContext(ctx))
		c.JSON(http.StatusConflict, cErr)
		return
	}
	cause := errors.Cause(err)
	switch cause {
	default:
		d.view.RenderInternalError(c, err)
		return
	case app.ErrModelArtifactNotUnique:
		d.view.RenderError(c, cause, http.StatusUnprocessableEntity)
		return
	case app.ErrModelParsingArtifactFailed:
		d.view.RenderError(c, formatArtifactUploadError(err), http.StatusBadRequest)
		return
	case utils.ErrStreamTooLarge, ErrModelArtifactFileTooLarge:
		d.view.RenderError(c, ErrModelArtifactFileTooLarge, http.StatusRequestEntityTooLarge)
		return
	case app.ErrModelMissingInputMetadata, app.ErrModelMissingInputArtifact,
		app.ErrModelInvalidMetadata, app.ErrModelMultipartUploadMsgMalformed,
		io.ErrUnexpectedEOF:
		d.view.RenderError(c, cause, http.StatusBadRequest)
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
func (d *DeploymentsApiHandlers) GenerateImage(c *gin.Context) {

	formReader, err := c.Request.MultipartReader()
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	// parse multipart message
	multipartMsg, err := d.ParseGenerateImageMultipart(formReader)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	tokenFields := strings.Fields(c.Request.Header.Get("Authorization"))
	if len(tokenFields) == 2 && strings.EqualFold(tokenFields[0], "Bearer") {
		multipartMsg.Token = tokenFields[1]
	}

	imgID, err := d.app.GenerateImage(c.Request.Context(), multipartMsg)
	cause := errors.Cause(err)
	switch cause {
	default:
		d.view.RenderInternalError(c, err)
	case nil:
		d.view.RenderSuccessPost(c, imgID)
	case app.ErrModelArtifactNotUnique:
		d.view.RenderError(c, cause, http.StatusUnprocessableEntity)
	case app.ErrModelParsingArtifactFailed:
		d.view.RenderError(c, formatArtifactUploadError(err), http.StatusBadRequest)
	case utils.ErrStreamTooLarge, ErrModelArtifactFileTooLarge:
		d.view.RenderError(c, ErrModelArtifactFileTooLarge, http.StatusRequestEntityTooLarge)
	case app.ErrModelMissingInputMetadata, app.ErrModelMissingInputArtifact,
		app.ErrModelInvalidMetadata, app.ErrModelMultipartUploadMsgMalformed,
		io.ErrUnexpectedEOF:
		d.view.RenderError(c, cause, http.StatusBadRequest)
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
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			dscr, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'description'",
				)
			}
			uploadMsg.MetaConstructor.Description = string(dscr)

		case "size":
			// Add size limit to the metadata
			reader := utils.ReadAtMost(part, 20)
			sz, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'size'",
				)
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
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			b, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'artifact_id'",
				)
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
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			b, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'args'",
				)
			}
			msg.Args = string(b)

		case "description":
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			b, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'description'",
				)
			}
			msg.Description = string(b)

		case "device_types_compatible":
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			b, err := io.ReadAll(reader)
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
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			b, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'name'",
				)
			}
			msg.Name = string(b)

		case "type":
			reader := utils.ReadAtMost(part, MaxFormParamSize)
			b, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'type'",
				)
			}
			msg.Type = string(b)

		case "size":
			// Add size limit to the metadata
			reader := utils.ReadAtMost(part, 20)
			sz, err := io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to read form value 'size'",
				)
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
	c *gin.Context,
	ctx context.Context,
	group string,
) {
	constructor, err := d.getDeploymentConstructorFromBody(c, group)
	if err != nil {
		d.view.RenderError(
			c,
			errors.Wrap(err, "Validating request body"),
			http.StatusBadRequest,
		)
		return
	}

	id, err := d.app.CreateDeployment(ctx, constructor)
	switch err {
	case nil:
		location := fmt.Sprintf("%s/%s", ApiUrlManagement+ApiUrlManagementDeployments, id)
		c.Writer.Header().Add("Location", location)
		c.Status(http.StatusCreated)
	case app.ErrNoArtifact:
		d.view.RenderError(c, err, http.StatusUnprocessableEntity)
	case app.ErrNoDevices:
		d.view.RenderError(c, err, http.StatusBadRequest)
	case app.ErrConflictingDeployment:
		d.view.RenderError(c, err, http.StatusConflict)
	default:
		d.view.RenderInternalError(c, err)
	}
}

func (d *DeploymentsApiHandlers) PostDeployment(c *gin.Context) {
	ctx := c.Request.Context()

	d.createDeployment(c, ctx, "")
}

func (d *DeploymentsApiHandlers) DeployToGroup(c *gin.Context) {
	ctx := c.Request.Context()

	group := c.Param("name")
	if len(group) < 1 {
		d.view.RenderError(c, ErrMissingGroupName, http.StatusBadRequest)
	}
	d.createDeployment(c, ctx, group)
}

// parseDeviceConfigurationDeploymentPathParams parses expected params
// and check if the params are not empty
func parseDeviceConfigurationDeploymentPathParams(c *gin.Context) (string, string, string, error) {
	tenantID := c.Param("tenant")
	deviceID := c.Param(ParamDeviceID)
	if deviceID == "" {
		return "", "", "", errors.New("device ID missing")
	}
	deploymentID := c.Param(ParamDeploymentID)
	if deploymentID == "" {
		return "", "", "", errors.New("deployment ID missing")
	}
	return tenantID, deviceID, deploymentID, nil
}

// getConfigurationDeploymentConstructorFromBody extracts configuration
// deployment constructor from the request body and validates it
func getConfigurationDeploymentConstructorFromBody(c *gin.Context) (
	*model.ConfigurationDeploymentConstructor, error) {

	var constructor *model.ConfigurationDeploymentConstructor

	if err := c.ShouldBindJSON(&constructor); err != nil {
		return nil, err
	}

	if err := constructor.Validate(); err != nil {
		return nil, err
	}

	return constructor, nil
}

// device configuration deployment handler
func (d *DeploymentsApiHandlers) PostDeviceConfigurationDeployment(
	c *gin.Context,
) {

	// get path params
	tenantID, deviceID, deploymentID, err := parseDeviceConfigurationDeploymentPathParams(c)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	// add tenant id to the context
	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{Tenant: tenantID})

	constructor, err := getConfigurationDeploymentConstructorFromBody(c)
	if err != nil {
		d.view.RenderError(
			c,
			errors.Wrap(err, "Validating request body"),
			http.StatusBadRequest,
		)
		return
	}

	id, err := d.app.CreateDeviceConfigurationDeployment(ctx, constructor, deviceID, deploymentID)
	switch err {
	default:
		d.view.RenderInternalError(c, err)
	case nil:
		c.Request.URL.Path = "./deployments"
		d.view.RenderSuccessPost(c, id)
	case app.ErrDuplicateDeployment:
		d.view.RenderError(c, err, http.StatusConflict)
	case app.ErrInvalidDeploymentID:
		d.view.RenderError(c, err, http.StatusBadRequest)
	}
}

func (d *DeploymentsApiHandlers) getDeploymentConstructorFromBody(
	c *gin.Context,
	group string,
) (*model.DeploymentConstructor, error) {
	var constructor *model.DeploymentConstructor
	if err := c.ShouldBindJSON(&constructor); err != nil {
		return nil, err
	}

	constructor.Group = group

	if err := constructor.ValidateNew(); err != nil {
		return nil, err
	}

	return constructor, nil
}

func (d *DeploymentsApiHandlers) GetDeployment(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	deployment, err := d.app.GetDeployment(ctx, id)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if deployment == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessGet(c, deployment)
}

func (d *DeploymentsApiHandlers) GetDeploymentStats(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	stats, err := d.app.GetDeploymentStats(ctx, id)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if stats == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessGet(c, stats)
}

func (d *DeploymentsApiHandlers) GetDeploymentsStats(c *gin.Context) {

	ctx := c.Request.Context()

	ids := model.DeploymentIDs{}
	if err := c.ShouldBindJSON(&ids); err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	if len(ids.IDs) == 0 {
		c.JSON(http.StatusOK, struct{}{})
		return
	}

	if err := ids.Validate(); err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	stats, err := d.app.GetDeploymentsStats(ctx, ids.IDs...)
	if err != nil {
		if errors.Is(err, app.ErrModelDeploymentNotFound) {
			d.view.RenderError(c, err, http.StatusNotFound)
			return
		}
		d.view.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (d *DeploymentsApiHandlers) GetDeploymentDeviceList(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	deployment, err := d.app.GetDeployment(ctx, id)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if deployment == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderSuccessGet(c, deployment.DeviceList)
}

func (d *DeploymentsApiHandlers) AbortDeployment(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	if !govalidator.IsUUID(id) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	// receive request body
	var status struct {
		Status model.DeviceDeploymentStatus
	}

	err := c.ShouldBindJSON(&status)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	// "aborted" is the only supported status
	if status.Status != model.DeviceDeploymentStatusAborted {
		d.view.RenderError(c, ErrUnexpectedDeploymentStatus, http.StatusBadRequest)
	}

	l := log.FromContext(ctx)
	l.Infof("Abort deployment: %s", id)

	// Check if deployment is finished
	isDeploymentFinished, err := d.app.IsDeploymentFinished(ctx, id)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}
	if isDeploymentFinished {
		d.view.RenderError(c, ErrDeploymentAlreadyFinished, http.StatusUnprocessableEntity)
		return
	}

	// Abort deployments for devices and update deployment stats
	if err := d.app.AbortDeployment(ctx, id); err != nil {
		d.view.RenderInternalError(c, err)
	}

	d.view.RenderEmptySuccessResponse(c)
}

func (d *DeploymentsApiHandlers) GetDeploymentForDevice(c *gin.Context) {
	var (
		installed *model.InstalledDeviceDeployment
		ctx       = c.Request.Context()
		idata     = identity.FromContext(ctx)
	)
	if idata == nil {
		d.view.RenderError(c, ErrMissingIdentity, http.StatusBadRequest)
		return
	}

	q := c.Request.URL.Query()
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
			c.Request.URL.RawQuery = q.Encode()
		}
	}()
	if strings.EqualFold(c.Request.Method, http.MethodPost) {
		// POST
		installed = new(model.InstalledDeviceDeployment)
		if err := c.ShouldBindJSON(&installed); err != nil {
			d.view.RenderError(c,
				errors.Wrap(err, "invalid schema"),
				http.StatusBadRequest)
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
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	request := &model.DeploymentNextRequest{
		DeviceProvides: installed,
	}

	d.getDeploymentForDevice(c, idata, request)
}

func (d *DeploymentsApiHandlers) getDeploymentForDevice(
	c *gin.Context,
	idata *identity.Identity,
	request *model.DeploymentNextRequest,
) {
	ctx := c.Request.Context()

	deployment, err := d.app.GetDeploymentForDeviceWithCurrent(ctx, idata.Subject, request)
	if err != nil {
		if err == app.ErrConflictingRequestData {
			d.view.RenderError(c, err, http.StatusConflict)
		} else {
			d.view.RenderInternalError(c, err)
		}
		return
	}

	if deployment == nil {
		d.view.RenderNoUpdateForDevice(c)
		return
	} else if deployment.Type == model.DeploymentTypeConfiguration {
		// Generate pre-signed URL
		var hostName string = d.config.PresignHostname
		if hostName == "" {
			if hostName = c.Request.Header.Get(hdrForwardedHost); hostName == "" {
				d.view.RenderInternalError(c,
					errors.New("presign.hostname not configured; "+
						"unable to generate download link "+
						" for configuration deployment"))
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

	d.view.RenderSuccessGet(c, deployment)
}

func (d *DeploymentsApiHandlers) PutDeploymentStatusForDevice(
	c *gin.Context,
) {
	ctx := c.Request.Context()

	did := c.Param("id")

	idata := identity.FromContext(ctx)
	if idata == nil {
		d.view.RenderError(c, ErrMissingIdentity, http.StatusBadRequest)
		return
	}

	// receive request body
	var report model.StatusReport

	err := c.ShouldBindJSON(&report)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	l := log.FromContext(ctx)
	l.Infof("status: %+v", report)
	if err := d.app.UpdateDeviceDeploymentStatus(ctx, did,
		idata.Subject, model.DeviceDeploymentState{
			Status:   report.Status,
			SubState: report.SubState,
		}); err != nil {

		if err == app.ErrDeploymentAborted || err == app.ErrDeviceDecommissioned {
			d.view.RenderError(c, err, http.StatusConflict)
		} else if err == app.ErrStorageNotFound {
			d.view.RenderErrorNotFound(c)
		} else {
			d.view.RenderInternalError(c, err)
		}
		return
	}

	d.view.RenderEmptySuccessResponse(c)
}

func (d *DeploymentsApiHandlers) GetDeviceStatusesForDeployment(
	c *gin.Context,
) {
	ctx := c.Request.Context()

	did := c.Param("id")

	if !govalidator.IsUUID(did) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	statuses, err := d.app.GetDeviceStatusesForDeployment(ctx, did)
	if err != nil {
		switch err {
		case app.ErrModelDeploymentNotFound:
			d.view.RenderError(c, err, http.StatusNotFound)
			return
		default:
			d.view.RenderInternalError(c, err)
			return
		}
	}

	d.view.RenderSuccessGet(c, statuses)
}

func (d *DeploymentsApiHandlers) GetDevicesListForDeployment(
	c *gin.Context,
) {
	ctx := c.Request.Context()

	did := c.Param("id")

	if !govalidator.IsUUID(did) {
		d.view.RenderError(c, ErrIDNotUUID, http.StatusBadRequest)
		return
	}

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	lq := store.ListQuery{
		Skip:         int((page - 1) * perPage),
		Limit:        int(perPage),
		DeploymentID: did,
	}
	if status := c.Request.URL.Query().Get("status"); status != "" {
		lq.Status = &status
	}
	if err = lq.Validate(); err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	statuses, totalCount, err := d.app.GetDevicesListForDeployment(ctx, lq)
	if err != nil {
		switch err {
		case app.ErrModelDeploymentNotFound:
			d.view.RenderError(c, err, http.StatusNotFound)
			return
		default:
			d.view.RenderInternalError(c, err)
			return
		}
	}

	hasNext := totalCount > int(page*perPage)
	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
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
	d.view.RenderSuccessGet(c, statuses)
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

func (d *DeploymentsApiHandlers) LookupDeployment(c *gin.Context) {
	ctx := c.Request.Context()
	q := c.Request.URL.Query()
	defer func() {
		if search := q.Get("search"); search != "" {
			q.Set("search", Redacted)
			c.Request.URL.RawQuery = q.Encode()
		}
	}()

	query, err := ParseDeploymentLookupQueryV1(q)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	query.Skip = int((page - 1) * perPage)
	query.Limit = int(perPage + 1)

	deps, totalCount, err := d.app.LookupDeployment(ctx, query)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	c.Writer.Header().Add(hdrTotalCount, strconv.FormatInt(totalCount, 10))

	len := len(deps)
	hasNext := false
	if int64(len) > perPage {
		hasNext = true
		len = int(perPage)
	}

	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
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

	d.view.RenderSuccessGet(c, deps[:len])
}

func (d *DeploymentsApiHandlers) LookupDeploymentV2(c *gin.Context) {
	ctx := c.Request.Context()
	q := c.Request.URL.Query()
	defer func() {
		if q.Has("name") {
			q["name"] = []string{Redacted}
			c.Request.URL.RawQuery = q.Encode()
		}
	}()

	query, err := ParseDeploymentLookupQueryV2(q)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	query.Skip = int((page - 1) * perPage)
	query.Limit = int(perPage + 1)

	deps, totalCount, err := d.app.LookupDeployment(ctx, query)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}
	c.Writer.Header().Add(hdrTotalCount, strconv.FormatInt(totalCount, 10))

	len := len(deps)
	hasNext := false
	if int64(len) > perPage {
		hasNext = true
		len = int(perPage)
	}

	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
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

	d.view.RenderSuccessGet(c, deps[:len])
}

func (d *DeploymentsApiHandlers) PutDeploymentLogForDevice(c *gin.Context) {
	ctx := c.Request.Context()

	did := c.Param("id")

	idata := identity.FromContext(ctx)
	if idata == nil {
		d.view.RenderError(c, ErrMissingIdentity, http.StatusBadRequest)
		return
	}

	// reuse DeploymentLog, device and deployment IDs are ignored when
	// (un-)marshaling DeploymentLog to/from JSON
	var log model.DeploymentLog

	err := c.ShouldBindJSON(&log)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	if err := d.app.SaveDeviceDeploymentLog(ctx, idata.Subject,
		did, log.Messages); err != nil {

		if err == app.ErrModelDeploymentNotFound {
			d.view.RenderError(c, err, http.StatusNotFound)
		} else {
			d.view.RenderInternalError(c, err)
		}
		return
	}

	d.view.RenderEmptySuccessResponse(c)
}

func (d *DeploymentsApiHandlers) GetDeploymentLogForDevice(c *gin.Context) {
	ctx := c.Request.Context()

	did := c.Param("id")
	devid := c.Param("devid")

	depl, err := d.app.GetDeviceDeploymentLog(ctx, devid, did)

	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	if depl == nil {
		d.view.RenderErrorNotFound(c)
		return
	}

	d.view.RenderDeploymentLog(c, *depl)
}

func (d *DeploymentsApiHandlers) AbortDeviceDeployments(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	err := d.app.AbortDeviceDeployments(ctx, id)

	switch err {
	case nil, app.ErrStorageNotFound:
		d.view.RenderEmptySuccessResponse(c)
	default:
		d.view.RenderInternalError(c, err)
	}
}

func (d *DeploymentsApiHandlers) DeleteDeviceDeploymentsHistory(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	err := d.app.DeleteDeviceDeploymentsHistory(ctx, id)

	switch err {
	case nil, app.ErrStorageNotFound:
		d.view.RenderEmptySuccessResponse(c)
	default:
		d.view.RenderInternalError(c, err)
	}
}

func (d *DeploymentsApiHandlers) ListDeviceDeployments(c *gin.Context) {
	ctx := c.Request.Context()
	d.listDeviceDeployments(ctx, c, true)
}

func (d *DeploymentsApiHandlers) ListDeviceDeploymentsInternal(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.Param("tenant")
	if tenantID != "" {
		ctx = identity.WithContext(c.Request.Context(), &identity.Identity{
			Tenant:   tenantID,
			IsDevice: true,
		})
	}
	d.listDeviceDeployments(ctx, c, true)
}

func (d *DeploymentsApiHandlers) ListDeviceDeploymentsByIDsInternal(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.Param("tenant")
	if tenantID != "" {
		ctx = identity.WithContext(c.Request.Context(), &identity.Identity{
			Tenant:   tenantID,
			IsDevice: true,
		})
	}
	d.listDeviceDeployments(ctx, c, false)
}

func (d *DeploymentsApiHandlers) listDeviceDeployments(ctx context.Context,
	c *gin.Context, byDeviceID bool) {

	did := ""
	var IDs []string
	if byDeviceID {
		did = c.Param("id")
	} else {
		values := c.Request.URL.Query()
		if values.Has("id") && len(values["id"]) > 0 {
			IDs = values["id"]
		} else {
			d.view.RenderError(c, ErrEmptyID, http.StatusBadRequest)
			return
		}
	}

	page, perPage, err := rest.ParsePagingParameters(c.Request)
	if err == nil && perPage > MaximumPerPageListDeviceDeployments {
		err = rest.ErrQueryParmLimit(ParamPerPage)
	}
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	lq := store.ListQueryDeviceDeployments{
		Skip:     int((page - 1) * perPage),
		Limit:    int(perPage),
		DeviceID: did,
		IDs:      IDs,
	}
	if status := c.Request.URL.Query().Get("status"); status != "" {
		lq.Status = &status
	}
	if err = lq.Validate(); err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	deps, totalCount, err := d.app.GetDeviceDeploymentListForDevice(ctx, lq)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}
	c.Writer.Header().Add(hdrTotalCount, strconv.FormatInt(int64(totalCount), 10))

	hasNext := totalCount > lq.Skip+len(deps)

	hints := rest.NewPagingHints().
		SetPage(page).
		SetPerPage(perPage).
		SetHasNext(hasNext).
		SetTotalCount(int64(totalCount))

	links, err := rest.MakePagingHeaders(c.Request, hints)
	if err != nil {
		rest.RenderInternalError(c, err)
		return
	}
	for _, l := range links {
		c.Writer.Header().Add(hdrLink, l)
	}

	d.view.RenderSuccessGet(c, deps)
}

func (d *DeploymentsApiHandlers) AbortDeviceDeploymentsInternal(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.Param("tenantID")
	if tenantID != "" {
		ctx = identity.WithContext(c.Request.Context(), &identity.Identity{
			Tenant:   tenantID,
			IsDevice: true,
		})
	}

	id := c.Param("id")

	// Decommission deployments for devices and update deployment stats
	err := d.app.DecommissionDevice(ctx, id)

	switch err {
	case nil, app.ErrStorageNotFound:
		d.view.RenderEmptySuccessResponse(c)
	default:
		d.view.RenderInternalError(c, err)

	}
}

// tenants

func (d *DeploymentsApiHandlers) ProvisionTenantsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	defer c.Request.Body.Close()

	tenant, err := model.ParseNewTenantReq(c.Request.Body)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	err = d.app.ProvisionTenant(ctx, tenant.TenantId)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (d *DeploymentsApiHandlers) DeploymentsPerTenantHandler(
	c *gin.Context,
) {
	tenantID := c.Param("tenant")
	if tenantID == "" {

		d.view.RenderError(c, errors.New("missing tenant ID"), http.StatusBadRequest)
		return
	}
	c.Request = c.Request.WithContext(identity.WithContext(
		c.Request.Context(),
		&identity.Identity{Tenant: tenantID},
	))
	d.LookupDeployment(c)
}

func (d *DeploymentsApiHandlers) GetTenantStorageSettingsHandler(
	c *gin.Context,
) {

	tenantID := c.Param("tenant")

	ctx := identity.WithContext(
		c.Request.Context(),
		&identity.Identity{Tenant: tenantID},
	)

	settings, err := d.app.GetStorageSettings(ctx)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	d.view.RenderSuccessGet(c, settings)
}

func (d *DeploymentsApiHandlers) PutTenantStorageSettingsHandler(
	c *gin.Context,
) {

	defer c.Request.Body.Close()

	tenantID := c.Param("tenant")

	ctx := identity.WithContext(
		c.Request.Context(),
		&identity.Identity{Tenant: tenantID},
	)

	settings, err := model.ParseStorageSettingsRequest(c.Request.Body)
	if err != nil {
		d.view.RenderError(c, err, http.StatusBadRequest)
		return
	}

	err = d.app.SetStorageSettings(ctx, settings)
	if err != nil {
		d.view.RenderInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
