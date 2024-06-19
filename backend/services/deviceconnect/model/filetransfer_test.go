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
	"errors"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func str2pointer(val string) *string {
	return &val
}

func TestDownloadFileRequestValidation(t *testing.T) {
	testCases := []struct {
		Name    string
		Request *DownloadFileRequest
		Error   error
	}{
		{
			Name: "validation ok",
			Request: &DownloadFileRequest{
				Path: str2pointer("/path"),
			},
		},
		{
			Name: "validation failed, path is relative",
			Request: &DownloadFileRequest{
				Path: str2pointer("path/to/file"),
			},
			Error: errors.New("path: must be absolute."),
		},
		{
			Name: "validation failed, path empty",
			Request: &DownloadFileRequest{
				Path: str2pointer(""),
			},
			Error: errors.New("path: cannot be blank."),
		},
		{
			Name: "validation failed, missing path",
			Request: &DownloadFileRequest{
				Path: nil,
			},
			Error: errors.New("path: cannot be blank."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Request.Validate()
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadFileRequestValidation(t *testing.T) {
	testCases := []struct {
		Name    string
		Request *UploadFileRequest
		Error   error
	}{
		{
			Name: "validation ok",
			Request: &UploadFileRequest{
				Path: str2pointer("/path"),
				File: &multipart.Part{},
			},
		},
		{
			Name: "validation failed, path is relative",
			Request: &UploadFileRequest{
				Path: str2pointer("path/to/file"),
				File: &multipart.Part{},
			},
			Error: errors.New("path: must be absolute."),
		},
		{
			Name: "validation failed, path empty",
			Request: &UploadFileRequest{
				Path: str2pointer(""),
				File: &multipart.Part{},
			},
			Error: errors.New("path: cannot be blank."),
		},
		{
			Name: "validation failed, missing path",
			Request: &UploadFileRequest{
				Path: nil,
				File: &multipart.Part{},
			},
			Error: errors.New("path: cannot be blank."),
		},
		{
			Name: "validation failed, missing file",
			Request: &UploadFileRequest{
				Path: str2pointer("/absolute/path"),
			},
			Error: errors.New("file: cannot be blank."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Request.Validate()
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
