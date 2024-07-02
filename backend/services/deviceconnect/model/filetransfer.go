// Copyright 2021 Northern.tech AS
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
	"mime/multipart"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var absolutePathRegexp = regexp.MustCompile("^/")

// DownloadFileRequest stores the request to download a file
type DownloadFileRequest struct {
	// The file path to the file we are downloading
	Path *string `json:"path"`
}

// Validate validates the request
func (f DownloadFileRequest) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Path, validation.Required,
			validation.Match(absolutePathRegexp).Error("must be absolute")),
	)
}

// DownloadFileRequest stores the request to upload a file
type UploadFileRequest struct {
	// The source filename which will be appended to the
	// target path if it points to a directory
	SrcPath *string `json:"src_path"`
	// The file path to the file we are uploading
	Path *string `json:"path"`
	// The file owner
	UID *uint32 `json:"uid"`
	// The file group
	GID *uint32 `json:"gid"`
	// Mode contains the file mode and permission bits.
	Mode *uint32 `json:"mode"`
	// The file you are uploading
	File *multipart.Part `json:"file"`
}

// Validate validates the request
func (f UploadFileRequest) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Path, validation.Required,
			validation.Match(absolutePathRegexp).Error("must be absolute")),
		validation.Field(&f.File, validation.Required),
	)
}
