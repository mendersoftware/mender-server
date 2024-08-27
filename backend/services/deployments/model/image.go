// Copyright 2023 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package model

import (
	"context"
	"io"
	"path"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/doc"
)

const (
	ArtifactFileSuffix = ".mender"
)

var (
	StorageKeyImageProvidesIdxKey   = "meta_artifact.provides_idx.key"
	StorageKeyImageProvidesIdxValue = "meta_artifact.provides_idx.value"
)

type ProvidesIdx map[string]string

func ImagePathFromContext(ctx context.Context, id string) string {
	imgPath := id
	if idty := identity.FromContext(ctx); idty != nil {
		imgPath = path.Join(idty.Tenant, id)
	}
	return imgPath
}

// Information provided by the user
type ImageMeta struct {
	// Image description
	Description string `json:"description,omitempty" valid:"length(1|4096),optional"`
}

// Creates new, empty ImageMeta
func NewImageMeta() *ImageMeta {
	return &ImageMeta{}
}

// Validate checks structure according to valid tags.
func (s ImageMeta) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Description, lengthLessThan4096),
	)
}

// Structure with artifact version information
type ArtifactInfo struct {
	// Mender artifact format - the only possible value is "mender"
	//Format string `json:"format" valid:"string,equal("mender"),required"`
	Format string `json:"format" valid:"required"`

	// Mender artifact format version
	//Version uint `json:"version" valid:"uint,equal(1),required"`
	Version uint `json:"version" valid:"required"`
}

func (ai ArtifactInfo) Validate() error {
	return validation.ValidateStruct(&ai,
		validation.Field(&ai.Format, validation.Required),
		validation.Field(&ai.Version, validation.In(uint(1), uint(2), uint(3))),
	)
}

// Information provided by the Mender Artifact header
type ArtifactMeta struct {
	// artifact_name from artifact file
	Name string `json:"name" bson:"name" valid:"length(1|4096),required"`

	// Compatible device types for the application
	//nolint:lll
	DeviceTypesCompatible []string `json:"device_types_compatible" bson:"device_types_compatible" valid:"length(1|4096),required"`

	// Artifact version info
	Info *ArtifactInfo `json:"info"`

	// Flag that indicates if artifact is signed or not
	Signed bool `json:"signed" bson:"signed"`

	// List of updates
	Updates []Update `json:"updates" valid:"-"`

	// Provides is a map of artifact_provides used
	// for checking artifact (version 3) dependencies.
	//nolint: lll
	Provides map[string]string `json:"artifact_provides,omitempty" bson:"provides,omitempty" valid:"-"`

	// ProvidesIdx is special representation of provides
	// which makes possible to index and query using provides.
	ProvidesIdx ProvidesIdx `json:"-" bson:"provides_idx,omitempty"`

	// Depends is a map[string]interface{} (JSON) of artifact_depends used
	// for checking/validate against artifact (version 3) provides.
	Depends map[string]interface{} `json:"artifact_depends,omitempty" bson:"depends" valid:"-"`

	// ClearsProvides is a list of strings (JSON) of clears_artifact_provides used
	// for clearing already-installed artifact (version 3) provides.
	//nolint:lll
	ClearsProvides []string `json:"clears_artifact_provides,omitempty" bson:"clears_provides,omitempty" valid:"-"`
}

// MarshalBSON transparently creates depends_idx field on bson.Marshal
func (am ArtifactMeta) MarshalBSON() ([]byte, error) {
	if err := am.Validate(); err != nil {
		return nil, err
	}
	dependsIdx, err := doc.UnwindMap(am.Depends)
	if err != nil {
		return nil, err
	}
	doc := doc.DocumentFromStruct(am, bson.E{
		Key: "depends_idx", Value: dependsIdx,
	})
	return bson.Marshal(doc)
}

// MarshalBSONValue transparently creates depends_idx field on bson.MarshalValue
// which is called if ArtifactMeta is marshaled as an embedded document.
func (am ArtifactMeta) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if err := am.Validate(); err != nil {
		return bson.TypeNull, nil, err
	}
	dependsIdx, err := doc.UnwindMap(am.Depends)
	if err != nil {
		return bson.TypeNull, nil, err
	}
	doc := doc.DocumentFromStruct(am, bson.E{
		Key: "depends_idx", Value: dependsIdx,
	})
	return bson.MarshalValue(doc)
}

// Validate checks structure according to valid tags.
func (am *ArtifactMeta) Validate() error {
	if am.Depends == nil {
		am.Depends = make(map[string]interface{})
	}
	am.Depends["device_type"] = am.DeviceTypesCompatible

	return validation.ValidateStruct(am,
		validation.Field(&am.Name, validation.Required, lengthIn1To4096),
		validation.Field(&am.DeviceTypesCompatible,
			validation.Required,
			lengthIn0To200,
			validation.Each(lengthIn1To4096),
		),
		validation.Field(&am.Info),
	)
}

func NewArtifactMeta() *ArtifactMeta {
	return &ArtifactMeta{}
}

// Image YOCTO image with user application
type Image struct {
	// Image ID
	Id string `json:"id" bson:"_id" valid:"uuidv4,required"`

	// User provided field set
	*ImageMeta `bson:"meta"`

	// Field set provided with yocto image
	*ArtifactMeta `bson:"meta_artifact"`

	// Artifact total size
	Size int64 `json:"size" bson:"size" valid:"-"`

	// Last modification time, including image upload time
	Modified *time.Time `json:"modified" valid:"-"`
}

func (img Image) MarshalBSON() (b []byte, err error) {
	return bson.Marshal(doc.DocumentFromStruct(img))
}

func (img Image) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(doc.DocumentFromStruct(img))
}

// Validate checks structure according to valid tags.
func (s Image) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Id, validation.Required, is.UUID),
		validation.Field(&s.ImageMeta),
		validation.Field(&s.ArtifactMeta),
	)
}

// NewImage creates new software image object.
func NewImage(
	id string,
	metaConstructor *ImageMeta,
	metaArtifactConstructor *ArtifactMeta,
	artifactSize int64) *Image {

	now := time.Now()

	return &Image{
		ImageMeta:    metaConstructor,
		ArtifactMeta: metaArtifactConstructor,
		Modified:     &now,
		Id:           id,
		Size:         artifactSize,
	}
}

// SetModified set last modification time for the image.
func (s *Image) SetModified(time time.Time) {
	s.Modified = &time
}

type ReadCounter interface {
	io.Reader
	// Count returns the number of bytes read.
	Count() int64
}

// MultipartUploadMsg is a structure with fields extracted from the multipart/form-data form
// send in the artifact upload request
type MultipartUploadMsg struct {
	// user metadata constructor
	MetaConstructor *ImageMeta
	// ArtifactID contains the artifact ID
	ArtifactID string
	// reader pointing to the beginning of the artifact data
	ArtifactReader io.Reader
}

// MultipartGenerateImageMsg is a structure with fields extracted from the multipart/form-data
// form sent in the artifact generation request
type MultipartGenerateImageMsg struct {
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	DeviceTypesCompatible []string  `json:"device_types_compatible"`
	Type                  string    `json:"type"`
	Args                  string    `json:"args"`
	ArtifactID            string    `json:"artifact_id"`
	GetArtifactURI        string    `json:"get_artifact_uri"`
	DeleteArtifactURI     string    `json:"delete_artifact_uri"`
	TenantID              string    `json:"tenant_id"`
	Token                 string    `json:"token"`
	FileReader            io.Reader `json:"-"`
}

func (msg MultipartGenerateImageMsg) Validate() error {
	if err := validation.ValidateStruct(&msg,
		validation.Field(&msg.Name, validation.Required),
		validation.Field(&msg.DeviceTypesCompatible, validation.Required),
		validation.Field(&msg.Type, validation.Required),
	); err != nil {
		return err
	}
	// Somehow FileReader is not covered by "required" rule.
	if msg.FileReader == nil {
		return errors.New("missing 'file' section")
	}
	return nil
}

type provideInternal struct {
	Key   string
	Value string
}

func (p ProvidesIdx) MarshalBSONValue() (bsontype.Type, []byte, error) {
	attrs := make([]provideInternal, len(p))
	i := 0
	for k, v := range p {
		attrs[i].Key = k
		attrs[i].Value = v
		i++
	}
	return bson.MarshalValue(attrs)
}

func (p *ProvidesIdx) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	raw := bson.Raw(b)
	elems, err := raw.Elements()
	if err != nil {
		return err
	}
	*p = make(ProvidesIdx, len(elems))
	var tmp provideInternal
	for _, elem := range elems {
		err = elem.Value().Unmarshal(&tmp)
		if err != nil {
			return err
		}
		(*p)[tmp.Key] = tmp.Value
	}

	return nil
}
