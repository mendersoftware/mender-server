// Copyright 2020 Northern.tech AS
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

import "testing"

const (
	validUUIDv4  = "d50eda0d-2cea-4de1-8d42-9cd3e7e8670d"
	artifactSize = 10000
)

func TestValidateEmptyImageMeta(t *testing.T) {
	image := NewImageMeta()

	if err := image.Validate(); err != nil {
		t.FailNow()
	}
}

func TestValidateEmptyImageMetaArtifact(t *testing.T) {
	image := NewArtifactMeta()

	if err := image.Validate(); err == nil {
		t.FailNow()
	}
}

func TestValidateCorrectImageMeta(t *testing.T) {
	image := NewImageMeta()

	if err := image.Validate(); err != nil {
		t.FailNow()
	}
}

func TestValidateCorrectImageMetaYocot(t *testing.T) {
	image := NewArtifactMeta()
	required := "required"

	image.Name = required
	image.DeviceTypesCompatible = []string{"required"}
	image.Info = &ArtifactInfo{
		Format:  required,
		Version: 1,
	}

	if err := image.Validate(); err != nil {
		t.FailNow()
	}
}

func TestValidateCorrectImage(t *testing.T) {
	required := "required"
	imageMeta := NewImageMeta()
	imageMetaArtifact := NewArtifactMeta()

	imageMetaArtifact.Name = required
	imageMetaArtifact.DeviceTypesCompatible = []string{"required"}

	image := NewImage(
		validUUIDv4, imageMeta, imageMetaArtifact, artifactSize)

	if err := image.Validate(); err != nil {
		t.Errorf("%v", err)
	}
}
