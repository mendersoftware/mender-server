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

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-artifact/areader"
	"github.com/mendersoftware/mender-artifact/artifact"
	"github.com/mendersoftware/mender-artifact/awriter"
	"github.com/mendersoftware/mender-artifact/handlers"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deployments/client/inventory"
	"github.com/mendersoftware/mender-server/services/deployments/client/reporting"
	"github.com/mendersoftware/mender-server/services/deployments/client/workflows"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/store/mongo"
	"github.com/mendersoftware/mender-server/services/deployments/utils"
)

const (
	ArtifactContentType              = "application/vnd.mender-artifact"
	ArtifactConfigureProvides        = "data-partition.mender-configure.version"
	ArtifactConfigureProvidesCleared = "data-partition.mender-configure.*"

	DefaultUpdateDownloadLinkExpire  = 24 * time.Hour
	DefaultImageGenerationLinkExpire = 7 * 24 * time.Hour
	PerPageInventoryDevices          = 512
	InventoryGroupScope              = "system"
	InventoryIdentityScope           = "identity"
	InventoryGroupAttributeName      = "group"
	InventoryStatusAttributeName     = "status"
	InventoryStatusAccepted          = "accepted"

	fileSuffixTmp = ".tmp"

	inprogressIdleTime = time.Hour
)

var (
	ArtifactConfigureType = "mender-configure"
)

// Errors expected from App interface
var (
	// images
	ErrImageMetaNotFound                = errors.New("Image metadata is not found")
	ErrModelMultipartUploadMsgMalformed = errors.New("Multipart upload message malformed")
	ErrModelMissingInputMetadata        = errors.New("Missing input metadata")
	ErrModelMissingInputArtifact        = errors.New("Missing input artifact")
	ErrModelInvalidMetadata             = errors.New("Metadata invalid")
	ErrModelArtifactNotUnique           = errors.New("Artifact not unique")
	ErrModelImageInActiveDeployment     = errors.New(
		"Image is used in active deployment and cannot be removed",
	)
	ErrModelImageUsedInAnyDeployment = errors.New("Image has already been used in deployment")
	ErrModelParsingArtifactFailed    = errors.New("Cannot parse artifact file")
	ErrUploadNotFound                = errors.New("artifact object not found")
	ErrEmptyArtifact                 = errors.New("artifact cannot be nil")

	ErrMsgArtifactConflict = "An artifact with the same name has conflicting dependencies"

	// deployments
	ErrModelMissingInput       = errors.New("Missing input deployment data")
	ErrModelInvalidDeviceID    = errors.New("Invalid device ID")
	ErrModelDeploymentNotFound = errors.New("Deployment not found")
	ErrModelInternal           = errors.New("Internal error")
	ErrStorageInvalidLog       = errors.New("Invalid deployment log")
	ErrStorageNotFound         = errors.New("Not found")
	ErrDeploymentAborted       = errors.New("Deployment aborted")
	ErrDeviceDecommissioned    = errors.New("Device decommissioned")
	ErrNoArtifact              = errors.New("No artifact for the deployment")
	ErrNoDevices               = errors.New("No devices for the deployment")
	ErrDuplicateDeployment     = errors.New("Deployment with given ID already exists")
	ErrInvalidDeploymentID     = errors.New("Deployment ID must be a valid UUID")
	ErrConflictingRequestData  = errors.New("Device provided conflicting request data")
	ErrConflictingDeployment   = errors.New(
		"Invalid deployment definition: there is already an active deployment with " +
			"the same parameters",
	)
)

//deployments

//go:generate ../utils/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error
	// limits
	GetLimit(ctx context.Context, name string) (*model.Limit, error)
	ProvisionTenant(ctx context.Context, tenant_id string) error

	// Storage Settings
	GetStorageSettings(ctx context.Context) (*model.StorageSettings, error)
	SetStorageSettings(ctx context.Context, storageSettings *model.StorageSettings) error

	// images
	ListImages(
		ctx context.Context,
		filters *model.ReleaseOrImageFilter,
	) ([]*model.Image, int, error)
	DownloadLink(ctx context.Context, imageID string,
		expire time.Duration) (*model.Link, error)
	UploadLink(
		ctx context.Context,
		expire time.Duration,
		skipVerify bool,
	) (*model.UploadLink, error)
	CompleteUpload(
		ctx context.Context,
		intentID string,
		skipVerify bool,
		metadata *model.DirectUploadMetadata,
	) error
	GetImage(ctx context.Context, id string) (*model.Image, error)
	DeleteImage(ctx context.Context, imageID string) error
	CreateImage(ctx context.Context,
		multipartUploadMsg *model.MultipartUploadMsg) (string, error)
	GenerateImage(ctx context.Context,
		multipartUploadMsg *model.MultipartGenerateImageMsg) (string, error)
	GenerateConfigurationImage(
		ctx context.Context,
		deviceType string,
		deploymentID string,
	) (io.Reader, error)
	EditImage(ctx context.Context, id string,
		constructorData *model.ImageMeta) (bool, error)

	// deployments
	CreateDeployment(ctx context.Context,
		constructor *model.DeploymentConstructor) (string, error)
	GetDeployment(ctx context.Context, deploymentID string) (*model.Deployment, error)
	IsDeploymentFinished(ctx context.Context, deploymentID string) (bool, error)
	AbortDeployment(ctx context.Context, deploymentID string) error
	GetDeploymentStats(ctx context.Context, deploymentID string) (model.Stats, error)
	GetDeploymentsStats(ctx context.Context,
		deploymentIDs ...string) ([]*model.DeploymentStats, error)
	GetDeploymentForDeviceWithCurrent(ctx context.Context, deviceID string,
		request *model.DeploymentNextRequest) (*model.DeploymentInstructions, error)
	HasDeploymentForDevice(ctx context.Context, deploymentID string,
		deviceID string) (bool, error)
	UpdateDeviceDeploymentStatus(ctx context.Context, deploymentID string,
		deviceID string, state model.DeviceDeploymentState) error
	GetDeviceStatusesForDeployment(ctx context.Context,
		deploymentID string) ([]model.DeviceDeployment, error)
	GetDevicesListForDeployment(ctx context.Context,
		query store.ListQuery) ([]model.DeviceDeployment, int, error)
	GetDeviceDeploymentListForDevice(ctx context.Context,
		query store.ListQueryDeviceDeployments) ([]model.DeviceDeploymentListItem, int, error)
	LookupDeployment(ctx context.Context,
		query model.Query) ([]*model.Deployment, int64, error)
	SaveDeviceDeploymentLog(ctx context.Context, deviceID string,
		deploymentID string, logs []model.LogMessage) error
	GetDeviceDeploymentLog(ctx context.Context,
		deviceID, deploymentID string) (*model.DeploymentLog, error)
	AbortDeviceDeployments(ctx context.Context, deviceID string) error
	DeleteDeviceDeploymentsHistory(ctx context.Context, deviceId string) error
	DecommissionDevice(ctx context.Context, deviceID string) error
	CreateDeviceConfigurationDeployment(
		ctx context.Context, constructor *model.ConfigurationDeploymentConstructor,
		deviceID, deploymentID string) (string, error)
	UpdateDeploymentsWithArtifactName(
		ctx context.Context,
		artifactName string,
	) error
	GetDeviceDeploymentLastStatus(
		ctx context.Context,
		devicesIds []string,
	) (
		model.DeviceDeploymentLastStatuses,
		error,
	)

	// releases
	ReplaceReleaseTags(ctx context.Context, releaseName string, tags model.Tags) error
	UpdateRelease(ctx context.Context, releaseName string, release model.ReleasePatch) error
	ListReleaseTags(ctx context.Context) (model.Tags, error)
	GetReleasesUpdateTypes(ctx context.Context) ([]string, error)
	DeleteReleases(ctx context.Context, releaseNames []string) ([]string, error)
}

type Deployments struct {
	db              store.DataStore
	objectStorage   storage.ObjectStorage
	workflowsClient workflows.Client
	inventoryClient inventory.Client
	reportingClient reporting.Client
}

// Compile-time check
var _ App = &Deployments{}

func NewDeployments(
	storage store.DataStore,
	objectStorage storage.ObjectStorage,
	maxActiveDeployments int64,
	withAuditLogs bool,
) *Deployments {
	return &Deployments{
		db:              storage,
		objectStorage:   objectStorage,
		workflowsClient: workflows.NewClient(),
		inventoryClient: inventory.NewClient(),
	}
}

func (d *Deployments) SetWorkflowsClient(workflowsClient workflows.Client) {
	d.workflowsClient = workflowsClient
}

func (d *Deployments) SetInventoryClient(inventoryClient inventory.Client) {
	d.inventoryClient = inventoryClient
}

func (d *Deployments) HealthCheck(ctx context.Context) error {
	err := d.db.Ping(ctx)
	if err != nil {
		return errors.Wrap(err, "error reaching MongoDB")
	}
	err = d.objectStorage.HealthCheck(ctx)
	if err != nil {
		return errors.Wrap(
			err,
			"error reaching artifact storage service",
		)
	}

	err = d.workflowsClient.CheckHealth(ctx)
	if err != nil {
		return errors.Wrap(err, "Workflows service unhealthy")
	}

	err = d.inventoryClient.CheckHealth(ctx)
	if err != nil {
		return errors.Wrap(err, "Inventory service unhealthy")
	}

	if d.reportingClient != nil {
		err = d.reportingClient.CheckHealth(ctx)
		if err != nil {
			return errors.Wrap(err, "Reporting service unhealthy")
		}
	}
	return nil
}

func (d *Deployments) contextWithStorageSettings(
	ctx context.Context,
) (context.Context, error) {
	var err error
	settings, ok := storage.SettingsFromContext(ctx)
	if !ok {
		settings, err = d.db.GetStorageSettings(ctx)
		if err != nil {
			return nil, err
		}
	}
	if settings != nil {
		if settings.UseAccelerate && settings.Uri != "" {
			log.FromContext(ctx).
				Warn(`storage settings: custom "uri" and "use_accelerate" ` +
					`are not allowed: disabling transfer acceleration`)
			settings.UseAccelerate = false
		}
		err = settings.Validate()
		if err != nil {
			return nil, err
		}
	}
	return storage.SettingsWithContext(ctx, settings), nil
}

func (d *Deployments) GetLimit(ctx context.Context, name string) (*model.Limit, error) {
	limit, err := d.db.GetLimit(ctx, name)
	if err == mongo.ErrLimitNotFound {
		return &model.Limit{
			Name:  name,
			Value: 0,
		}, nil

	} else if err != nil {
		return nil, errors.Wrap(err, "failed to obtain limit from storage")
	}
	return limit, nil
}

func (d *Deployments) ProvisionTenant(ctx context.Context, tenant_id string) error {
	if err := d.db.ProvisionTenant(ctx, tenant_id); err != nil {
		return errors.Wrap(err, "failed to provision tenant")
	}

	return nil
}

// CreateImage parses artifact and uploads artifact file to the file storage - in parallel,
// and creates image structure in the system.
// Returns image ID and nil on success.
func (d *Deployments) CreateImage(ctx context.Context,
	multipartUploadMsg *model.MultipartUploadMsg) (string, error) {
	return d.handleArtifact(ctx, multipartUploadMsg, false, nil)
}

func (d *Deployments) saveUpdateTypes(ctx context.Context, image *model.Image) {
	l := log.FromContext(ctx)
	if image != nil && image.ArtifactMeta != nil && len(image.ArtifactMeta.Updates) > 0 {
		i := 0
		updateTypes := make([]string, len(image.ArtifactMeta.Updates))
		for _, t := range image.ArtifactMeta.Updates {
			if t.TypeInfo.Type == nil {
				continue
			}
			updateTypes[i] = *t.TypeInfo.Type
			i++
		}
		err := d.db.SaveUpdateTypes(ctx, updateTypes[:i])
		if err != nil {
			l.Errorf(
				"error while saving the update types for the artifact: %s",
				err.Error(),
			)
		}
	}
}

// handleArtifact parses artifact and uploads artifact file to the file storage - in parallel,
// and creates image structure in the system.
// Returns image ID, artifact file ID and nil on success.
func (d *Deployments) handleArtifact(ctx context.Context,
	multipartUploadMsg *model.MultipartUploadMsg,
	skipVerify bool,
	metadata *model.DirectUploadMetadata,
) (string, error) {

	l := log.FromContext(ctx)
	ctx, err := d.contextWithStorageSettings(ctx)
	if err != nil {
		return "", err
	}

	// create pipe
	pR, pW := io.Pipe()

	artifactReader := utils.CountReads(multipartUploadMsg.ArtifactReader)

	tee := io.TeeReader(artifactReader, pW)

	uid, err := uuid.Parse(multipartUploadMsg.ArtifactID)
	if err != nil {
		uid, _ = uuid.NewRandom()
	}
	artifactID := uid.String()

	ch := make(chan error)
	// create goroutine for artifact upload
	//
	// reading from the pipe (which is done in UploadArtifact method) is a blocking operation
	// and cannot be done in the same goroutine as writing to the pipe
	//
	// uploading and parsing artifact in the same process will cause in a deadlock!
	//nolint:errcheck
	go func() (err error) {
		defer func() { ch <- err }()
		if skipVerify {
			err = nil
			io.Copy(io.Discard, pR)
			return nil
		}
		err = d.objectStorage.PutObject(
			ctx, model.ImagePathFromContext(ctx, artifactID), pR,
		)
		if err != nil {
			pR.CloseWithError(err)
		}
		return err
	}()

	// parse artifact
	// artifact library reads all the data from the given reader
	metaArtifactConstructor, err := getMetaFromArchive(&tee, skipVerify)
	if err != nil {
		_ = pW.CloseWithError(err)
		<-ch
		return artifactID, errors.Wrap(ErrModelParsingArtifactFailed, err.Error())
	}
	validMetadata := false
	if skipVerify && metadata != nil {
		// this means we got files and metadata separately
		// we can now put it in the metaArtifactConstructor
		// after validating that the files information match the artifact
		validMetadata = validUpdates(metaArtifactConstructor.Updates, metadata.Updates)
		if validMetadata {
			metaArtifactConstructor.Updates = metadata.Updates
		}
	}
	// validate artifact metadata
	if err = metaArtifactConstructor.Validate(); err != nil {
		return artifactID, ErrModelInvalidMetadata
	}

	if !skipVerify {
		// read the rest of the data,
		// just in case the artifact library did not read all the data from the reader
		_, err = io.Copy(io.Discard, tee)
		if err != nil {
			// CloseWithError will cause the reading end to abort upload.
			_ = pW.CloseWithError(err)
			<-ch
			return artifactID, err
		}
	}

	// close the pipe
	pW.Close()

	// collect output from the goroutine
	if uploadResponseErr := <-ch; uploadResponseErr != nil {
		return artifactID, uploadResponseErr
	}

	size := artifactReader.Count()
	if skipVerify && validMetadata {
		size = metadata.Size
	}
	image := model.NewImage(
		artifactID,
		multipartUploadMsg.MetaConstructor,
		metaArtifactConstructor,
		size,
	)

	// save image structure in the system
	if err = d.db.InsertImage(ctx, image); err != nil {
		// Try to remove the storage from s3.
		if errDelete := d.objectStorage.DeleteObject(
			ctx, model.ImagePathFromContext(ctx, artifactID),
		); errDelete != nil {
			l.Errorf(
				"failed to clean up artifact storage after failure: %s",
				errDelete,
			)
		}
		if idxErr, ok := err.(*model.ConflictError); ok {
			return artifactID, idxErr
		}
		return artifactID, errors.Wrap(err, "Fail to store the metadata")
	}
	d.saveUpdateTypes(ctx, image)

	// update release
	if err := d.updateRelease(ctx, image, nil); err != nil {
		return "", err
	}

	if err := d.UpdateDeploymentsWithArtifactName(ctx, metaArtifactConstructor.Name); err != nil {
		return "", errors.Wrap(err, "fail to update deployments")
	}

	return artifactID, nil
}

func validUpdates(constructorUpdates []model.Update, metadataUpdates []model.Update) bool {
	valid := false
	if len(constructorUpdates) == len(metadataUpdates) {
		valid = true
		for _, update := range constructorUpdates {
			for _, updateExternal := range metadataUpdates {
				if !update.Match(updateExternal) {
					valid = false
					break
				}
			}
		}
	}
	return valid
}

// GenerateImage parses raw data and uploads it to the file storage - in parallel,
// creates image structure in the system, and starts the workflow to generate the
// artifact from them.
// Returns image ID and nil on success.
func (d *Deployments) GenerateImage(ctx context.Context,
	multipartGenerateImageMsg *model.MultipartGenerateImageMsg) (string, error) {

	if multipartGenerateImageMsg == nil {
		return "", ErrModelMultipartUploadMsgMalformed
	}

	imgPath, err := d.handleRawFile(ctx, multipartGenerateImageMsg)
	if err != nil {
		return "", err
	}
	if id := identity.FromContext(ctx); id != nil && len(id.Tenant) > 0 {
		multipartGenerateImageMsg.TenantID = id.Tenant
	}
	err = d.workflowsClient.StartGenerateArtifact(ctx, multipartGenerateImageMsg)
	if err != nil {
		if cleanupErr := d.objectStorage.DeleteObject(ctx, imgPath); cleanupErr != nil {
			return "", errors.Wrap(err, cleanupErr.Error())
		}
		return "", err
	}

	return multipartGenerateImageMsg.ArtifactID, err
}

func (d *Deployments) GenerateConfigurationImage(
	ctx context.Context,
	deviceType string,
	deploymentID string,
) (io.Reader, error) {
	var buf bytes.Buffer
	dpl, err := d.db.FindDeploymentByID(ctx, deploymentID)
	if err != nil {
		return nil, err
	} else if dpl == nil {
		return nil, ErrModelDeploymentNotFound
	}
	var metaData map[string]interface{}
	err = json.Unmarshal(dpl.Configuration, &metaData)
	if err != nil {
		return nil, errors.Wrapf(err, "malformed configuration in deployment")
	}

	artieWriter := awriter.NewWriter(&buf, artifact.NewCompressorNone())
	module := handlers.NewModuleImage(ArtifactConfigureType)
	err = artieWriter.WriteArtifact(&awriter.WriteArtifactArgs{
		Format:  "mender",
		Version: 3,
		Devices: []string{deviceType},
		Name:    dpl.ArtifactName,
		Updates: &awriter.Updates{Updates: []handlers.Composer{module}},
		Depends: &artifact.ArtifactDepends{
			CompatibleDevices: []string{deviceType},
		},
		Provides: &artifact.ArtifactProvides{
			ArtifactName: dpl.ArtifactName,
		},
		MetaData: metaData,
		TypeInfoV3: &artifact.TypeInfoV3{
			Type: &ArtifactConfigureType,
			ArtifactProvides: artifact.TypeInfoProvides{
				ArtifactConfigureProvides: dpl.ArtifactName,
			},
			ArtifactDepends:        artifact.TypeInfoDepends{},
			ClearsArtifactProvides: []string{ArtifactConfigureProvidesCleared},
		},
	})

	return &buf, err
}

// handleRawFile parses raw data, uploads it to the file storage,
// and starts the workflow to generate the artifact.
// Returns the object path to the file and nil on success.
func (d *Deployments) handleRawFile(ctx context.Context,
	multipartMsg *model.MultipartGenerateImageMsg) (filePath string, err error) {
	l := log.FromContext(ctx)
	uid, _ := uuid.NewRandom()
	artifactID := uid.String()
	multipartMsg.ArtifactID = artifactID
	filePath = model.ImagePathFromContext(ctx, artifactID+fileSuffixTmp)

	// check if artifact is unique
	// artifact is considered to be unique if there is no artifact with the same name
	// and supporting the same platform in the system
	isArtifactUnique, err := d.db.IsArtifactUnique(ctx,
		multipartMsg.Name,
		multipartMsg.DeviceTypesCompatible,
	)
	if err != nil {
		return "", errors.Wrap(err, "Fail to check if artifact is unique")
	}
	if !isArtifactUnique {
		return "", ErrModelArtifactNotUnique
	}

	ctx, err = d.contextWithStorageSettings(ctx)
	if err != nil {
		return "", err
	}
	err = d.objectStorage.PutObject(
		ctx, filePath, multipartMsg.FileReader,
	)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			e := d.objectStorage.DeleteObject(ctx, filePath)
			if e != nil {
				l.Errorf("error cleaning up raw file '%s' from objectstorage: %s",
					filePath, e)
			}
		}
	}()

	link, err := d.objectStorage.GetRequest(
		ctx,
		filePath,
		path.Base(filePath),
		DefaultImageGenerationLinkExpire,
	)
	if err != nil {
		return "", err
	}
	multipartMsg.GetArtifactURI = link.Uri

	link, err = d.objectStorage.DeleteRequest(ctx, filePath, DefaultImageGenerationLinkExpire)
	if err != nil {
		return "", err
	}
	multipartMsg.DeleteArtifactURI = link.Uri

	return artifactID, nil
}

// GetImage allows to fetch image object with specified id
// Nil if not found
func (d *Deployments) GetImage(ctx context.Context, id string) (*model.Image, error) {

	image, err := d.db.FindImageByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "Searching for image with specified ID")
	}

	if image == nil {
		return nil, nil
	}

	return image, nil
}

// DeleteImage removes metadata and image file
// Noop for not existing images
// Allowed to remove image only if image is not scheduled or in progress for an updates - then image
// file is needed
// In case of already finished updates only image file is not needed, metadata is attached directly
// to device deployment therefore we still have some information about image that have been used
// (but not the file)
func (d *Deployments) DeleteImage(ctx context.Context, imageID string) error {
	found, err := d.GetImage(ctx, imageID)

	if err != nil {
		return errors.Wrap(err, "Getting image metadata")
	}

	if found == nil {
		return ErrImageMetaNotFound
	}

	inUse, err := d.ImageUsedInActiveDeployment(ctx, imageID)
	if err != nil {
		return errors.Wrap(err, "Checking if image is used in active deployment")
	}

	// Image is in use, not allowed to delete
	if inUse {
		return ErrModelImageInActiveDeployment
	}

	// Delete image file (call to external service)
	// Noop for not existing file
	ctx, err = d.contextWithStorageSettings(ctx)
	if err != nil {
		return err
	}
	imagePath := model.ImagePathFromContext(ctx, imageID)
	if err := d.objectStorage.DeleteObject(ctx, imagePath); err != nil {
		return errors.Wrap(err, "Deleting image file")
	}

	// Delete metadata
	if err := d.db.DeleteImage(ctx, imageID); err != nil {
		return errors.Wrap(err, "Deleting image metadata")
	}

	// update release
	if err := d.updateRelease(ctx, nil, found); err != nil {
		return err
	}

	return nil
}

// ListImages according to specified filers.
func (d *Deployments) ListImages(
	ctx context.Context,
	filters *model.ReleaseOrImageFilter,
) ([]*model.Image, int, error) {
	imageList, count, err := d.db.ListImages(ctx, filters)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Searching for image metadata")
	}

	if imageList == nil {
		return make([]*model.Image, 0), 0, nil
	}

	return imageList, count, nil
}

// EditObject allows editing only if image have not been used yet in any deployment.
func (d *Deployments) EditImage(ctx context.Context, imageID string,
	constructor *model.ImageMeta) (bool, error) {

	if err := constructor.Validate(); err != nil {
		return false, errors.Wrap(err, "Validating image metadata")
	}

	found, err := d.ImageUsedInDeployment(ctx, imageID)
	if err != nil {
		return false, errors.Wrap(err, "Searching for usage of the image among deployments")
	}

	if found {
		return false, ErrModelImageUsedInAnyDeployment
	}

	foundImage, err := d.db.FindImageByID(ctx, imageID)
	if err != nil {
		return false, errors.Wrap(err, "Searching for image with specified ID")
	}

	if foundImage == nil {
		return false, nil
	}

	foundImage.SetModified(time.Now())
	foundImage.ImageMeta = constructor

	_, err = d.db.Update(ctx, foundImage)
	if err != nil {
		return false, errors.Wrap(err, "Updating image matadata")
	}

	if err := d.updateReleaseEditArtifact(ctx, foundImage); err != nil {
		return false, err
	}

	return true, nil
}

// DownloadLink presigned GET link to download image file.
// Returns error if image have not been uploaded.
func (d *Deployments) DownloadLink(ctx context.Context, imageID string,
	expire time.Duration) (*model.Link, error) {

	image, err := d.GetImage(ctx, imageID)
	if err != nil {
		return nil, errors.Wrap(err, "Searching for image with specified ID")
	}

	if image == nil {
		return nil, nil
	}

	ctx, err = d.contextWithStorageSettings(ctx)
	if err != nil {
		return nil, err
	}
	imagePath := model.ImagePathFromContext(ctx, imageID)
	_, err = d.objectStorage.StatObject(ctx, imagePath)
	if err != nil {
		return nil, errors.Wrap(err, "Searching for image file")
	}

	link, err := d.objectStorage.GetRequest(
		ctx,
		imagePath,
		image.Name+model.ArtifactFileSuffix,
		expire,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Generating download link")
	}

	return link, nil
}

func (d *Deployments) UploadLink(
	ctx context.Context,
	expire time.Duration,
	skipVerify bool,
) (*model.UploadLink, error) {
	ctx, err := d.contextWithStorageSettings(ctx)
	if err != nil {
		return nil, err
	}

	artifactID := uuid.New().String()
	path := model.ImagePathFromContext(ctx, artifactID) + fileSuffixTmp
	if skipVerify {
		path = model.ImagePathFromContext(ctx, artifactID)
	}
	link, err := d.objectStorage.PutRequest(ctx, path, expire)
	if err != nil {
		return nil, errors.WithMessage(err, "app: failed to generate signed URL")
	}
	upLink := &model.UploadLink{
		ArtifactID: artifactID,
		IssuedAt:   time.Now(),
		Link:       *link,
	}
	err = d.db.InsertUploadIntent(ctx, upLink)
	if err != nil {
		return nil, errors.WithMessage(err, "app: error recording the upload intent")
	}

	return upLink, err
}

func (d *Deployments) processUploadedArtifact(
	ctx context.Context,
	artifactID string,
	artifact io.ReadCloser,
	skipVerify bool,
	metadata *model.DirectUploadMetadata,
) error {
	linkStatus := model.LinkStatusCompleted

	l := log.FromContext(ctx)
	defer artifact.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() { // Heatbeat routine
		ticker := time.NewTicker(inprogressIdleTime / 2)
		done := ctx.Done()
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := d.db.UpdateUploadIntentStatus(
					ctx,
					artifactID,
					model.LinkStatusProcessing,
					model.LinkStatusProcessing,
				)
				if err != nil {
					l.Errorf("failed to update upload link timestamp: %s", err)
					cancel()
					return
				}
			case <-done:
				return
			}
		}
	}()
	_, err := d.handleArtifact(ctx, &model.MultipartUploadMsg{
		ArtifactID:     artifactID,
		ArtifactReader: artifact,
	},
		skipVerify,
		metadata,
	)
	if err != nil {
		l.Warnf("failed to process artifact %s: %s", artifactID, err)
		linkStatus = model.LinkStatusAborted
	}
	errDB := d.db.UpdateUploadIntentStatus(
		ctx, artifactID,
		model.LinkStatusProcessing, linkStatus,
	)
	if errDB != nil {
		l.Warnf("failed to update upload link status: %s", errDB)
	}
	return err
}

func (d *Deployments) CompleteUpload(
	ctx context.Context,
	intentID string,
	skipVerify bool,
	metadata *model.DirectUploadMetadata,
) error {
	l := log.FromContext(ctx)
	idty := identity.FromContext(ctx)
	ctx, err := d.contextWithStorageSettings(ctx)
	if err != nil {
		return err
	}
	// Create an async context that doesn't cancel when server connection
	// closes.
	ctxAsync := context.Background()
	ctxAsync = log.WithContext(ctxAsync, l)
	ctxAsync = identity.WithContext(ctxAsync, idty)

	settings, _ := storage.SettingsFromContext(ctx)
	ctxAsync = storage.SettingsWithContext(ctxAsync, settings)
	var artifactReader io.ReadCloser
	if skipVerify {
		artifactReader, err = d.objectStorage.GetObject(
			ctxAsync,
			model.ImagePathFromContext(ctx, intentID),
		)
	} else {
		artifactReader, err = d.objectStorage.GetObject(
			ctxAsync,
			model.ImagePathFromContext(ctx, intentID)+fileSuffixTmp,
		)
	}
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return ErrUploadNotFound
		}
		return err
	}

	err = d.db.UpdateUploadIntentStatus(
		ctx,
		intentID,
		model.LinkStatusPending,
		model.LinkStatusProcessing,
	)
	if err != nil {
		errClose := artifactReader.Close()
		if errClose != nil {
			l.Warnf("failed to close artifact reader: %s", errClose)
		}
		if errors.Is(err, store.ErrNotFound) {
			return ErrUploadNotFound
		}
		return err
	}
	go d.processUploadedArtifact( // nolint:errcheck
		ctxAsync, intentID, artifactReader, skipVerify, metadata,
	)
	return nil
}

func getArtifactInfo(info artifact.Info) *model.ArtifactInfo {
	return &model.ArtifactInfo{
		Format:  info.Format,
		Version: uint(info.Version),
	}
}

func getUpdateFiles(uFiles []*handlers.DataFile) ([]model.UpdateFile, error) {
	var files []model.UpdateFile
	for _, u := range uFiles {
		files = append(files, model.UpdateFile{
			Name:     u.Name,
			Size:     u.Size,
			Date:     &u.Date,
			Checksum: string(u.Checksum),
		})
	}
	return files, nil
}

func getMetaFromArchive(r *io.Reader, skipVerify bool) (*model.ArtifactMeta, error) {
	metaArtifact := model.NewArtifactMeta()

	aReader := areader.NewReader(*r)

	// There is no signature verification here.
	// It is just simple check if artifact is signed or not.
	aReader.VerifySignatureCallback = func(message, sig []byte) error {
		metaArtifact.Signed = true
		return nil
	}

	var err error
	if skipVerify {
		err = aReader.ReadArtifactHeaders()
		if err != nil {
			return nil, errors.Wrap(err, "reading artifact error")
		}
	} else {
		err = aReader.ReadArtifact()
		if err != nil {
			return nil, errors.Wrap(err, "reading artifact error")
		}
	}

	metaArtifact.Info = getArtifactInfo(aReader.GetInfo())
	metaArtifact.DeviceTypesCompatible = aReader.GetCompatibleDevices()

	metaArtifact.Name = aReader.GetArtifactName()
	if metaArtifact.Info.Version == 3 {
		metaArtifact.Depends, err = aReader.MergeArtifactDepends()
		if err != nil {
			return nil, errors.Wrap(err,
				"error parsing version 3 artifact")
		}

		metaArtifact.Provides, err = aReader.MergeArtifactProvides()
		if err != nil {
			return nil, errors.Wrap(err,
				"error parsing version 3 artifact")
		}

		metaArtifact.ClearsProvides = aReader.MergeArtifactClearsProvides()
	}

	for _, p := range aReader.GetHandlers() {
		uFiles, err := getUpdateFiles(p.GetUpdateFiles())
		if err != nil {
			return nil, errors.Wrap(err, "Cannot get update files:")
		}

		uMetadata, err := p.GetUpdateMetaData()
		if err != nil {
			return nil, errors.Wrap(err, "Cannot get update metadata")
		}

		metaArtifact.Updates = append(
			metaArtifact.Updates,
			model.Update{
				TypeInfo: model.ArtifactUpdateTypeInfo{
					Type: p.GetUpdateType(),
				},
				Files:    uFiles,
				MetaData: uMetadata,
			})
	}

	return metaArtifact, nil
}

func getArtifactIDs(artifacts []*model.Image) []string {
	artifactIDs := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		artifactIDs = append(artifactIDs, artifact.Id)
	}
	return artifactIDs
}

// deployments
func inventoryDevicesToDevicesIds(devices []model.InvDevice) []string {
	ids := make([]string, len(devices))
	for i, d := range devices {
		ids[i] = d.ID
	}

	return ids
}

// updateDeploymentConstructor fills devices list with device ids
func (d *Deployments) updateDeploymentConstructor(ctx context.Context,
	constructor *model.DeploymentConstructor) (*model.DeploymentConstructor, error) {
	l := log.FromContext(ctx)

	id := identity.FromContext(ctx)
	if id == nil {
		l.Error("identity not present in the context")
		return nil, ErrModelInternal
	}
	searchParams := model.SearchParams{
		Page:    1,
		PerPage: PerPageInventoryDevices,
		Filters: []model.FilterPredicate{
			{
				Scope:     InventoryIdentityScope,
				Attribute: InventoryStatusAttributeName,
				Type:      "$eq",
				Value:     InventoryStatusAccepted,
			},
		},
	}
	if len(constructor.Group) > 0 {
		searchParams.Filters = append(
			searchParams.Filters,
			model.FilterPredicate{
				Scope:     InventoryGroupScope,
				Attribute: InventoryGroupAttributeName,
				Type:      "$eq",
				Value:     constructor.Group,
			})
	}

	for {
		devices, count, err := d.search(ctx, id.Tenant, searchParams)
		if err != nil {
			l.Errorf("error searching for devices")
			return nil, ErrModelInternal
		}
		if count < 1 {
			l.Errorf("no devices found")
			return nil, ErrNoDevices
		}
		if len(devices) < 1 {
			break
		}
		constructor.Devices = append(constructor.Devices, inventoryDevicesToDevicesIds(devices)...)
		if len(constructor.Devices) == count {
			break
		}
		searchParams.Page++
	}

	return constructor, nil
}

// CreateDeviceConfigurationDeployment creates new configuration deployment for the device.
func (d *Deployments) CreateDeviceConfigurationDeployment(
	ctx context.Context, constructor *model.ConfigurationDeploymentConstructor,
	deviceID, deploymentID string) (string, error) {

	if constructor == nil {
		return "", ErrModelMissingInput
	}

	deployment, err := model.NewDeploymentFromConfigurationDeploymentConstructor(
		constructor,
		deploymentID,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to create deployment")
	}

	deployment.DeviceList = []string{deviceID}
	deployment.MaxDevices = 1
	deployment.Configuration = []byte(constructor.Configuration)
	deployment.Type = model.DeploymentTypeConfiguration

	groups, err := d.getDeploymentGroups(ctx, []string{deviceID})
	if err != nil {
		return "", err
	}
	deployment.Groups = groups

	if err := d.db.InsertDeployment(ctx, deployment); err != nil {
		if err == mongo.ErrConflictingDeployment {
			return "", ErrDuplicateDeployment
		}
		if strings.Contains(err.Error(), "id: must be a valid UUID") {
			return "", ErrInvalidDeploymentID
		}
		return "", errors.Wrap(err, "Storing deployment data")
	}

	return deployment.Id, nil
}

// CreateDeployment precomputes new deployment and schedules it for devices.
func (d *Deployments) CreateDeployment(ctx context.Context,
	constructor *model.DeploymentConstructor) (string, error) {

	var err error

	if constructor == nil {
		return "", ErrModelMissingInput
	}

	if err := constructor.Validate(); err != nil {
		return "", errors.Wrap(err, "Validating deployment")
	}

	if len(constructor.Group) > 0 || constructor.AllDevices {
		constructor, err = d.updateDeploymentConstructor(ctx, constructor)
		if err != nil {
			return "", err
		}
	}

	deployment, err := model.NewDeploymentFromConstructor(constructor)
	if err != nil {
		return "", errors.Wrap(err, "failed to create deployment")
	}

	// Assign artifacts to the deployment.
	// When new artifact(s) with the artifact name same as the one in the deployment
	// will be uploaded to the backend, it will also become part of this deployment.
	artifacts, err := d.db.ImagesByName(ctx, deployment.ArtifactName)
	if err != nil {
		return "", errors.Wrap(err, "Finding artifact with given name")
	}

	if len(artifacts) == 0 {
		return "", ErrNoArtifact
	}

	deployment.Artifacts = getArtifactIDs(artifacts)
	deployment.DeviceList = constructor.Devices
	deployment.MaxDevices = len(constructor.Devices)
	deployment.Type = model.DeploymentTypeSoftware
	if len(constructor.Group) > 0 {
		deployment.Groups = []string{constructor.Group}
	}

	// single device deployment case
	if len(deployment.Groups) == 0 && len(constructor.Devices) == 1 {
		groups, err := d.getDeploymentGroups(ctx, constructor.Devices)
		if err != nil {
			return "", err
		}
		deployment.Groups = groups
	}

	if err := d.db.InsertDeployment(ctx, deployment); err != nil {
		if err == mongo.ErrConflictingDeployment {
			return "", ErrConflictingDeployment
		}
		return "", errors.Wrap(err, "Storing deployment data")
	}

	return deployment.Id, nil
}

func (d *Deployments) getDeploymentGroups(
	ctx context.Context,
	devices []string,
) ([]string, error) {
	id := identity.FromContext(ctx)

	//only for single device deployment case
	if len(devices) != 1 {
		return nil, nil
	}

	if id == nil {
		id = &identity.Identity{}
	}

	groups, err := d.inventoryClient.GetDeviceGroups(ctx, id.Tenant, devices[0])
	if err != nil && err != inventory.ErrDevNotFound {
		return nil, err
	}
	return groups, nil
}

// IsDeploymentFinished checks if there is unfinished deployment with given ID
func (d *Deployments) IsDeploymentFinished(
	ctx context.Context,
	deploymentID string,
) (bool, error) {
	deployment, err := d.db.FindUnfinishedByID(ctx, deploymentID)
	if err != nil {
		return false, errors.Wrap(err, "Searching for unfinished deployment by ID")
	}
	if deployment == nil {
		return true, nil
	}

	return false, nil
}

// GetDeployment fetches deployment by ID
func (d *Deployments) GetDeployment(ctx context.Context,
	deploymentID string) (*model.Deployment, error) {

	deployment, err := d.db.FindDeploymentByID(ctx, deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "Searching for deployment by ID")
	}

	if err := d.setDeploymentDeviceCountIfUnset(ctx, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

// ImageUsedInActiveDeployment checks if specified image is in use by deployments. Image is
// considered to be in use if it's participating in at lest one non success/error deployment.
func (d *Deployments) ImageUsedInActiveDeployment(ctx context.Context,
	imageID string) (bool, error) {

	var found bool

	found, err := d.db.ExistUnfinishedByArtifactId(ctx, imageID)
	if err != nil {
		return false, errors.Wrap(err, "Checking if image is used by active deployment")
	}

	return found, nil
}

// ImageUsedInDeployment checks if specified image is in use by deployments.
// Image is considered to be in use if it's participating in any deployment.
func (d *Deployments) ImageUsedInDeployment(ctx context.Context, imageID string) (bool, error) {

	var found bool

	found, err := d.db.ExistByArtifactId(ctx, imageID)
	if err != nil {
		return false, errors.Wrap(err, "Checking if image is used by active deployment")
	}

	return found, nil
}

// Retrieves the model.Deployment and model.DeviceDeployment structures
// for the device. Upon error, nil is returned for both deployment structures.
func (d *Deployments) getDeploymentForDevice(ctx context.Context,
	deviceID string) (*model.Deployment, *model.DeviceDeployment, error) {

	// Retrieve device deployment
	deviceDeployment, err := d.db.FindOldestActiveDeviceDeployment(ctx, deviceID)

	if err != nil {
		return nil, nil, errors.Wrap(err,
			"Searching for oldest active deployment for the device")
	} else if deviceDeployment == nil {
		return d.getNewDeploymentForDevice(ctx, deviceID)
	}

	deployment, err := d.db.FindDeploymentByID(ctx, deviceDeployment.DeploymentId)
	if err != nil {
		return nil, nil, errors.Wrap(err, "checking deployment id")
	}
	if deployment == nil {
		return nil, nil, errors.New("No deployment corresponding to device deployment")
	}

	return deployment, deviceDeployment, nil
}

// getNewDeploymentForDevice returns deployment object and creates and returns
// new device deployment for the device;
//
// we are interested only in the deployments that are newer than the latest
// deployment applied by the device;
// this way we guarantee that the device will not receive deployment
// that is older than the one installed on the device;
func (d *Deployments) getNewDeploymentForDevice(ctx context.Context,
	deviceID string) (*model.Deployment, *model.DeviceDeployment, error) {

	var lastDeployment *time.Time
	//get latest device deployment for the device;
	deviceDeployment, err := d.db.FindLatestInactiveDeviceDeployment(ctx, deviceID)
	if err != nil {
		return nil, nil, errors.Wrap(err,
			"Searching for latest active deployment for the device")
	} else if deviceDeployment == nil {
		lastDeployment = &time.Time{}
	} else {
		lastDeployment = deviceDeployment.Created
	}

	//get deployments newer then last device deployment
	//iterate over deployments and check if the device is part of the deployment or not
	var deploy *model.Deployment
	deploy, err = d.db.FindNewerActiveDeployment(ctx, lastDeployment, deviceID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to search for newer active deployments")
	}
	if deploy != nil {
		deviceDeployment, err := d.createDeviceDeploymentWithStatus(ctx,
			deviceID, deploy, model.DeviceDeploymentStatusPending)
		if err != nil {
			return nil, nil, err
		}
		return deploy, deviceDeployment, nil
	}
	return nil, nil, nil
}

func (d *Deployments) createDeviceDeploymentWithStatus(
	ctx context.Context, deviceID string,
	deployment *model.Deployment, status model.DeviceDeploymentStatus,
) (*model.DeviceDeployment, error) {
	prevStatus := model.DeviceDeploymentStatusNull
	deviceDeployment, err := d.db.GetDeviceDeployment(ctx, deployment.Id, deviceID, true)
	if err != nil && err != mongo.ErrStorageNotFound {
		return nil, err
	} else if deviceDeployment != nil {
		prevStatus = deviceDeployment.Status
	}

	deviceDeployment = model.NewDeviceDeployment(deviceID, deployment.Id)
	deviceDeployment.Status = status
	deviceDeployment.Active = status.Active()
	deviceDeployment.Created = deployment.Created

	if err := d.setDeploymentDeviceCountIfUnset(ctx, deployment); err != nil {
		return nil, err
	}

	if err := d.db.InsertDeviceDeployment(ctx, deviceDeployment,
		prevStatus == model.DeviceDeploymentStatusNull); err != nil {
		return nil, err
	}

	if prevStatus != status {
		beforeStatus := deployment.GetStatus()
		// after inserting new device deployment update deployment stats
		// in the database, and update deployment status
		deployment.Stats, err = d.db.UpdateStatsInc(
			ctx, deployment.Id,
			prevStatus, status,
		)
		if err != nil {
			return nil, err
		}
		newStatus := deployment.GetStatus()
		if beforeStatus != newStatus {
			err = d.db.SetDeploymentStatus(
				ctx, deployment.Id,
				newStatus, time.Now(),
			)
			if err != nil {
				return nil, errors.Wrap(err,
					"failed to update deployment status")
			}
		}
	}

	if !status.Active() {
		err := d.reindexDevice(ctx, deviceID)
		if err != nil {
			l := log.FromContext(ctx)
			l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
		}
		if err := d.reindexDeployment(ctx, deviceDeployment.DeviceId,
			deviceDeployment.DeploymentId, deviceDeployment.Id); err != nil {
			l := log.FromContext(ctx)
			l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
		}
	}

	return deviceDeployment, nil
}

// GetDeploymentForDeviceWithCurrent returns deployment for the device
func (d *Deployments) GetDeploymentForDeviceWithCurrent(ctx context.Context, deviceID string,
	request *model.DeploymentNextRequest) (*model.DeploymentInstructions, error) {

	deployment, deviceDeployment, err := d.getDeploymentForDevice(ctx, deviceID)
	if err != nil {
		return nil, ErrModelInternal
	} else if deployment == nil {
		return nil, nil
	}

	err = d.saveDeviceDeploymentRequest(ctx, deviceID, deviceDeployment, request)
	if err != nil {
		return nil, err
	}
	return d.getDeploymentInstructions(ctx, deployment, deviceDeployment, request)
}

func (d *Deployments) getDeploymentInstructions(
	ctx context.Context,
	deployment *model.Deployment,
	deviceDeployment *model.DeviceDeployment,
	request *model.DeploymentNextRequest,
) (*model.DeploymentInstructions, error) {

	var newArtifactAssigned bool

	l := log.FromContext(ctx)

	if deployment.Type == model.DeploymentTypeConfiguration {
		// There's nothing more we need to do, the link must be filled
		// in by the API layer.
		return &model.DeploymentInstructions{
			ID: deployment.Id,
			Artifact: model.ArtifactDeploymentInstructions{
				// configuration artifacts are created on demand, so they do not have IDs
				// use deployment ID togheter with device ID as artifact ID
				ID:                    deployment.Id + deviceDeployment.DeviceId,
				ArtifactName:          deployment.ArtifactName,
				DeviceTypesCompatible: []string{request.DeviceProvides.DeviceType},
			},
			Type: model.DeploymentTypeConfiguration,
		}, nil
	}

	// assing artifact to the device deployment
	// only if it was not assgined previously
	if deviceDeployment.Image == nil {
		if err := d.assignArtifact(
			ctx, deployment, deviceDeployment, request.DeviceProvides); err != nil {
			return nil, err
		}
		newArtifactAssigned = true
	}

	if deviceDeployment.Image == nil {
		// No artifact - return empty response
		return nil, nil
	}

	// if the deployment is not forcing the installation, and
	// if artifact was recognized as already installed, and this is
	// a new device deployment - indicated by device deployment status "pending",
	// handle already installed artifact case
	if !deployment.ForceInstallation &&
		d.isAlreadyInstalled(request, deviceDeployment) &&
		deviceDeployment.Status == model.DeviceDeploymentStatusPending {
		return nil, d.handleAlreadyInstalled(ctx, deviceDeployment)
	}

	// if new artifact has been assigned to device deployment
	// add artifact size to deployment total size,
	// before returning deployment instruction to the device
	if newArtifactAssigned {
		if err := d.db.IncrementDeploymentTotalSize(
			ctx, deviceDeployment.DeploymentId, deviceDeployment.Image.Size); err != nil {
			l.Errorf("failed to increment deployment total size: %s", err.Error())
		}
	}

	ctx, err := d.contextWithStorageSettings(ctx)
	if err != nil {
		return nil, err
	}

	imagePath := model.ImagePathFromContext(ctx, deviceDeployment.Image.Id)
	link, err := d.objectStorage.GetRequest(
		ctx,
		imagePath,
		deviceDeployment.Image.Name+model.ArtifactFileSuffix,
		DefaultUpdateDownloadLinkExpire,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Generating download link for the device")
	}

	instructions := &model.DeploymentInstructions{
		ID: deviceDeployment.DeploymentId,
		Artifact: model.ArtifactDeploymentInstructions{
			ID: deviceDeployment.Image.Id,
			ArtifactName: deviceDeployment.Image.
				ArtifactMeta.Name,
			Source: *link,
			DeviceTypesCompatible: deviceDeployment.Image.
				ArtifactMeta.DeviceTypesCompatible,
		},
	}

	return instructions, nil
}

func (d *Deployments) saveDeviceDeploymentRequest(ctx context.Context, deviceID string,
	deviceDeployment *model.DeviceDeployment, request *model.DeploymentNextRequest) error {
	if deviceDeployment.Request != nil {
		if !reflect.DeepEqual(deviceDeployment.Request, request) {
			// the device reported different device type and/or artifact name during the
			// update process, this can happen if the mender-store DB in the client is not
			// persistent so a new deployment start without a previous one is still ongoing;
			// mark deployment for this device as failed to force client to rollback
			l := log.FromContext(ctx)
			l.Errorf(
				"Device with id %s reported new data: %s during update process;"+
					"old data: %s",
				deviceID, request, deviceDeployment.Request)

			if err := d.updateDeviceDeploymentStatus(ctx, deviceDeployment,
				model.DeviceDeploymentState{
					Status: model.DeviceDeploymentStatusFailure,
				}); err != nil {
				return errors.Wrap(err, "Failed to update deployment status")
			}
			if err := d.reindexDevice(ctx, deviceDeployment.DeviceId); err != nil {
				l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
			}
			if err := d.reindexDeployment(ctx, deviceDeployment.DeviceId,
				deviceDeployment.DeploymentId, deviceDeployment.Id); err != nil {
				l := log.FromContext(ctx)
				l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
			}
			return ErrConflictingRequestData
		}
	} else {
		// save the request
		if err := d.db.SaveDeviceDeploymentRequest(
			ctx, deviceDeployment.Id, request); err != nil {
			return err
		}
	}
	return nil
}

// updateDeviceDeploymentStatus will update the deployment status for device of
// ID `deviceID`. Returns nil if update was successful.
func (d *Deployments) UpdateDeviceDeploymentStatus(
	ctx context.Context,
	deviceID, deploymentID string,
	ddState model.DeviceDeploymentState,
) error {
	deviceDeployment, err := d.db.GetDeviceDeployment(
		ctx, deviceID, deploymentID, false,
	)
	if err == mongo.ErrStorageNotFound {
		return ErrStorageNotFound
	} else if err != nil {
		return err
	}

	return d.updateDeviceDeploymentStatus(ctx, deviceDeployment, ddState)
}

func (d *Deployments) updateDeviceDeploymentStatus(
	ctx context.Context,
	dd *model.DeviceDeployment,
	ddState model.DeviceDeploymentState,
) error {

	l := log.FromContext(ctx)

	l.Infof("New status: %s for device %s deployment: %v",
		ddState.Status, dd.DeviceId, dd.DeploymentId,
	)

	var finishTime *time.Time = nil
	if model.IsDeviceDeploymentStatusFinished(ddState.Status) {
		now := time.Now()
		finishTime = &now
	}

	currentStatus := dd.Status

	if currentStatus == model.DeviceDeploymentStatusAborted {
		return ErrDeploymentAborted
	}

	if currentStatus == model.DeviceDeploymentStatusDecommissioned {
		return ErrDeviceDecommissioned
	}

	// nothing to do
	if ddState.Status == currentStatus {
		return nil
	}

	// update finish time
	ddState.FinishTime = finishTime

	old, err := d.db.UpdateDeviceDeploymentStatus(ctx,
		dd.DeviceId, dd.DeploymentId, ddState, dd.Status)
	if err != nil {
		return err
	}

	if old != ddState.Status {
		// fetch deployment stats and update deployment status
		deployment, err := d.db.FindDeploymentByID(ctx, dd.DeploymentId)
		if err != nil {
			return errors.Wrap(err, "failed when searching for deployment")
		}
		beforeStatus := deployment.GetStatus()

		deployment.Stats, err = d.db.UpdateStatsInc(ctx, dd.DeploymentId, old, ddState.Status)
		if err != nil {
			return err
		}
		newStatus := deployment.GetStatus()
		if beforeStatus != newStatus {
			err = d.db.SetDeploymentStatus(ctx, dd.DeploymentId, newStatus, time.Now())
			if err != nil {
				return errors.Wrap(err, "failed to update deployment status")
			}
		}
	}

	if !ddState.Status.Active() {
		l := log.FromContext(ctx)
		ldd := model.DeviceDeployment{
			DeviceId:     dd.DeviceId,
			DeploymentId: dd.DeploymentId,
			Id:           dd.Id,
			Status:       ddState.Status,
		}
		if err := d.db.SaveLastDeviceDeploymentStatus(ctx, ldd); err != nil {
			l.Error(errors.Wrap(err, "failed to save last device deployment status").Error())
		}
		if err := d.reindexDevice(ctx, dd.DeviceId); err != nil {
			l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
		}
		if err := d.reindexDeployment(ctx, dd.DeviceId, dd.DeploymentId, dd.Id); err != nil {
			l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
		}
	}

	return nil
}

func (d *Deployments) GetDeploymentStats(ctx context.Context,
	deploymentID string) (model.Stats, error) {

	deployment, err := d.db.FindDeploymentByID(ctx, deploymentID)

	if err != nil {
		return nil, errors.Wrap(err, "checking deployment id")
	}

	if deployment == nil {
		return nil, nil
	}

	return deployment.Stats, nil
}
func (d *Deployments) GetDeploymentsStats(ctx context.Context,
	deploymentIDs ...string) (deploymentStats []*model.DeploymentStats, err error) {

	deploymentStats, err = d.db.FindDeploymentStatsByIDs(ctx, deploymentIDs...)

	if err != nil {
		return nil, errors.Wrap(err, "checking deployment statistics for IDs")
	}

	if deploymentStats == nil {
		return nil, ErrModelDeploymentNotFound
	}

	return deploymentStats, nil
}

// GetDeviceStatusesForDeployment retrieve device deployment statuses for a given deployment.
func (d *Deployments) GetDeviceStatusesForDeployment(ctx context.Context,
	deploymentID string) ([]model.DeviceDeployment, error) {

	deployment, err := d.db.FindDeploymentByID(ctx, deploymentID)
	if err != nil {
		return nil, ErrModelInternal
	}

	if deployment == nil {
		return nil, ErrModelDeploymentNotFound
	}

	statuses, err := d.db.GetDeviceStatusesForDeployment(ctx, deploymentID)
	if err != nil {
		return nil, ErrModelInternal
	}

	return statuses, nil
}

func (d *Deployments) GetDevicesListForDeployment(ctx context.Context,
	query store.ListQuery) ([]model.DeviceDeployment, int, error) {

	deployment, err := d.db.FindDeploymentByID(ctx, query.DeploymentID)
	if err != nil {
		return nil, -1, ErrModelInternal
	}

	if deployment == nil {
		return nil, -1, ErrModelDeploymentNotFound
	}

	statuses, totalCount, err := d.db.GetDevicesListForDeployment(ctx, query)
	if err != nil {
		return nil, -1, ErrModelInternal
	}

	return statuses, totalCount, nil
}

func (d *Deployments) GetDeviceDeploymentListForDevice(ctx context.Context,
	query store.ListQueryDeviceDeployments) ([]model.DeviceDeploymentListItem, int, error) {
	deviceDeployments, totalCount, err := d.db.GetDeviceDeploymentsForDevice(ctx, query)
	if err != nil {
		return nil, -1, errors.Wrap(err, "retrieving the list of deployment statuses")
	}

	deploymentIDs := make([]string, len(deviceDeployments))
	for i, deviceDeployment := range deviceDeployments {
		deploymentIDs[i] = deviceDeployment.DeploymentId
	}
	var deployments []*model.Deployment
	if len(deviceDeployments) > 0 {
		deployments, _, err = d.db.Find(ctx, model.Query{
			IDs:          deploymentIDs,
			Limit:        len(deviceDeployments),
			DisableCount: true,
		})
		if err != nil {
			return nil, -1, errors.Wrap(err, "retrieving the list of deployments")
		}
	}

	deploymentsMap := make(map[string]*model.Deployment, len(deployments))
	for _, deployment := range deployments {
		deploymentsMap[deployment.Id] = deployment
	}

	res := make([]model.DeviceDeploymentListItem, 0, len(deviceDeployments))
	for i, deviceDeployment := range deviceDeployments {
		if deployment, ok := deploymentsMap[deviceDeployment.DeploymentId]; ok {
			res = append(res, model.DeviceDeploymentListItem{
				Id:         deviceDeployment.Id,
				Deployment: deployment,
				Device:     &deviceDeployments[i],
			})
		} else {
			res = append(res, model.DeviceDeploymentListItem{
				Id:     deviceDeployment.Id,
				Device: &deviceDeployments[i],
			})
		}
	}

	return res, totalCount, nil
}

func (d *Deployments) setDeploymentDeviceCountIfUnset(
	ctx context.Context,
	deployment *model.Deployment,
) error {
	if deployment.DeviceCount == nil {
		deviceCount, err := d.db.DeviceCountByDeployment(ctx, deployment.Id)
		if err != nil {
			return errors.Wrap(err, "counting device deployments")
		}
		err = d.db.SetDeploymentDeviceCount(ctx, deployment.Id, deviceCount)
		if err != nil {
			return errors.Wrap(err, "setting the device count for the deployment")
		}
		deployment.DeviceCount = &deviceCount
	}

	return nil
}

func (d *Deployments) LookupDeployment(ctx context.Context,
	query model.Query) ([]*model.Deployment, int64, error) {
	list, totalCount, err := d.db.Find(ctx, query)

	if err != nil {
		return nil, 0, errors.Wrap(err, "searching for deployments")
	}

	if list == nil {
		return make([]*model.Deployment, 0), 0, nil
	}

	for _, deployment := range list {
		if err := d.setDeploymentDeviceCountIfUnset(ctx, deployment); err != nil {
			return nil, 0, err
		}
	}

	return list, totalCount, nil
}

// SaveDeviceDeploymentLog will save the deployment log for device of
// ID `deviceID`. Returns nil if log was saved successfully.
func (d *Deployments) SaveDeviceDeploymentLog(ctx context.Context, deviceID string,
	deploymentID string, logs []model.LogMessage) error {

	// repack to temporary deployment log and validate
	dlog := model.DeploymentLog{
		DeviceID:     deviceID,
		DeploymentID: deploymentID,
		Messages:     logs,
	}
	if err := dlog.Validate(); err != nil {
		return errors.Wrapf(err, ErrStorageInvalidLog.Error())
	}

	if has, err := d.HasDeploymentForDevice(ctx, deploymentID, deviceID); !has {
		if err != nil {
			return err
		} else {
			return ErrModelDeploymentNotFound
		}
	}

	if err := d.db.SaveDeviceDeploymentLog(ctx, dlog); err != nil {
		return err
	}

	return d.db.UpdateDeviceDeploymentLogAvailability(ctx,
		deviceID, deploymentID, true)
}

func (d *Deployments) GetDeviceDeploymentLog(ctx context.Context,
	deviceID, deploymentID string) (*model.DeploymentLog, error) {

	return d.db.GetDeviceDeploymentLog(ctx,
		deviceID, deploymentID)
}

func (d *Deployments) HasDeploymentForDevice(ctx context.Context,
	deploymentID string, deviceID string) (bool, error) {
	return d.db.HasDeploymentForDevice(ctx, deploymentID, deviceID)
}

// AbortDeployment aborts deployment for devices and updates deployment stats
func (d *Deployments) AbortDeployment(ctx context.Context, deploymentID string) error {

	if err := d.db.AbortDeviceDeployments(ctx, deploymentID); err != nil {
		return err
	}

	stats, err := d.db.AggregateDeviceDeploymentByStatus(
		ctx, deploymentID)
	if err != nil {
		return err
	}

	// update statistics
	if err := d.db.UpdateStats(ctx, deploymentID, stats); err != nil {
		return errors.Wrap(err, "failed to update deployment stats")
	}

	// when aborting the deployment we need to set status directly instead of
	// using recalcDeploymentStatus method;
	// it is possible that the deployment does not have any device deployments yet;
	// in that case, all statistics are 0 and calculating status based on statistics
	// will not work - the calculated status will be "pending"
	if err := d.db.SetDeploymentStatus(ctx,
		deploymentID, model.DeploymentStatusFinished, time.Now()); err != nil {
		return errors.Wrap(err, "failed to update deployment status")
	}

	return nil
}

func (d *Deployments) updateDeviceDeploymentsStatus(
	ctx context.Context,
	deviceId string,
	status model.DeviceDeploymentStatus,
) error {
	var latestDeployment *time.Time
	// Retrieve active device deployment for the device
	deviceDeployment, err := d.db.FindOldestActiveDeviceDeployment(ctx, deviceId)
	if err != nil {
		return errors.Wrap(err, "Searching for active deployment for the device")
	} else if deviceDeployment != nil {
		now := time.Now()
		ddStatus := model.DeviceDeploymentState{
			Status:     status,
			FinishTime: &now,
		}
		if err := d.updateDeviceDeploymentStatus(
			ctx, deviceDeployment, ddStatus,
		); err != nil {
			return errors.Wrap(err, "updating device deployment status")
		}
		latestDeployment = deviceDeployment.Created
	} else {
		// get latest device deployment for the device
		deviceDeployment, err := d.db.FindLatestInactiveDeviceDeployment(ctx, deviceId)
		if err != nil {
			return errors.Wrap(err, "Searching for latest active deployment for the device")
		} else if deviceDeployment == nil {
			latestDeployment = &time.Time{}
		} else {
			latestDeployment = deviceDeployment.Created
		}
	}

	// get deployments newer then last device deployment
	// iterate over deployments and check if the device is part of the deployment or not
	// if the device is part of the deployment create new, decommisioned device deployment
	var deploy *model.Deployment
	deploy, err = d.db.FindNewerActiveDeployment(ctx, latestDeployment, deviceId)
	if err != nil {
		return errors.Wrap(err, "Failed to search for newer active deployments")
	}
	if deploy != nil {
		deviceDeployment, err = d.createDeviceDeploymentWithStatus(ctx,
			deviceId, deploy, status)
		if err != nil {
			return err
		}
		if !status.Active() {
			if err := d.reindexDeployment(ctx, deviceDeployment.DeviceId,
				deviceDeployment.DeploymentId, deviceDeployment.Id); err != nil {
				l := log.FromContext(ctx)
				l.Warn(errors.Wrap(err, "failed to trigger a deployment reindex"))
			}
		}
	}

	if err := d.reindexDevice(ctx, deviceId); err != nil {
		l := log.FromContext(ctx)
		l.Warn(errors.Wrap(err, "failed to trigger a device reindex"))
	}

	return nil
}

// DecommissionDevice updates the status of all the pending and active deployments for a device
// to decommissioned
func (d *Deployments) DecommissionDevice(ctx context.Context, deviceId string) error {
	return d.updateDeviceDeploymentsStatus(
		ctx,
		deviceId,
		model.DeviceDeploymentStatusDecommissioned,
	)
}

// AbortDeviceDeployments aborts all the pending and active deployments for a device
func (d *Deployments) AbortDeviceDeployments(ctx context.Context, deviceId string) error {
	return d.updateDeviceDeploymentsStatus(
		ctx,
		deviceId,
		model.DeviceDeploymentStatusAborted,
	)
}

// DeleteDeviceDeploymentsHistory deletes the device deployments history
func (d *Deployments) DeleteDeviceDeploymentsHistory(ctx context.Context, deviceId string) error {
	// get device deployments which will be marked as deleted
	f := false
	dd, err := d.db.GetDeviceDeployments(ctx, 0, 0, deviceId, &f, false)
	if err != nil {
		return err
	}

	// no device deployments to update
	if len(dd) <= 0 {
		return nil
	}

	// mark device deployments as deleted
	if err := d.db.DeleteDeviceDeploymentsHistory(ctx, deviceId); err != nil {
		return err
	}

	// trigger reindexing of updated device deployments
	deviceDeployments := make([]workflows.DeviceDeploymentShortInfo, len(dd))
	for i, d := range dd {
		deviceDeployments[i].ID = d.Id
		deviceDeployments[i].DeviceID = d.DeviceId
		deviceDeployments[i].DeploymentID = d.DeploymentId
	}
	if d.reportingClient != nil {
		err = d.workflowsClient.StartReindexReportingDeploymentBatch(ctx, deviceDeployments)
	}
	return err
}

// Storage settings
func (d *Deployments) GetStorageSettings(ctx context.Context) (*model.StorageSettings, error) {
	settings, err := d.db.GetStorageSettings(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Searching for settings failed")
	}

	return settings, nil
}

func (d *Deployments) SetStorageSettings(
	ctx context.Context,
	storageSettings *model.StorageSettings,
) error {
	if storageSettings != nil {
		ctx = storage.SettingsWithContext(ctx, storageSettings)
		if err := d.objectStorage.HealthCheck(ctx); err != nil {
			return errors.WithMessage(err,
				"the provided storage settings failed the health check",
			)
		}
	}
	if err := d.db.SetStorageSettings(ctx, storageSettings); err != nil {
		return errors.Wrap(err, "Failed to save settings")
	}

	return nil
}

func (d *Deployments) WithReporting(c reporting.Client) *Deployments {
	d.reportingClient = c
	return d
}

func (d *Deployments) haveReporting() bool {
	return d.reportingClient != nil
}

func (d *Deployments) search(
	ctx context.Context,
	tid string,
	parms model.SearchParams,
) ([]model.InvDevice, int, error) {
	if d.haveReporting() {
		return d.reportingClient.Search(ctx, tid, parms)
	} else {
		return d.inventoryClient.Search(ctx, tid, parms)
	}
}

func (d *Deployments) UpdateDeploymentsWithArtifactName(
	ctx context.Context,
	artifactName string,
) error {
	// first check if there are pending deployments with given artifact name
	exists, err := d.db.ExistUnfinishedByArtifactName(ctx, artifactName)
	if err != nil {
		return errors.Wrap(err, "looking for deployments with given artifact name")
	}
	if !exists {
		return nil
	}

	// Assign artifacts to the deployments with given artifact name
	artifacts, err := d.db.ImagesByName(ctx, artifactName)
	if err != nil {
		return errors.Wrap(err, "Finding artifact with given name")
	}

	if len(artifacts) == 0 {
		return ErrNoArtifact
	}
	artifactIDs := getArtifactIDs(artifacts)
	return d.db.UpdateDeploymentsWithArtifactName(ctx, artifactName, artifactIDs)
}

func (d *Deployments) reindexDevice(ctx context.Context, deviceID string) error {
	if d.reportingClient != nil {
		return d.workflowsClient.StartReindexReporting(ctx, deviceID)
	}
	return nil
}

func (d *Deployments) reindexDeployment(ctx context.Context,
	deviceID, deploymentID, ID string) error {
	if d.reportingClient != nil {
		return d.workflowsClient.StartReindexReportingDeployment(ctx, deviceID, deploymentID, ID)
	}
	return nil
}
