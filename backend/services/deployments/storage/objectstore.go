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

package storage

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

var (
	ErrObjectNotFound = errors.New("object not found")
)

// ObjectStorage allows to store and manage large files
//
//go:generate ../utils/mockgen.sh
type ObjectStorage interface {
	HealthCheck(ctx context.Context) error
	GetObject(ctx context.Context, path string) (io.ReadCloser, error)
	PutObject(ctx context.Context, path string, src io.Reader) error
	DeleteObject(ctx context.Context, path string) error
	StatObject(ctx context.Context, path string) (*ObjectInfo, error)

	// The following interface generates signed URLs.
	GetRequest(ctx context.Context, path string, filename string,
		duration time.Duration) (*model.Link, error)
	DeleteRequest(ctx context.Context, path string,
		duration time.Duration) (*model.Link, error)
	PutRequest(ctx context.Context, path string,
		duration time.Duration) (*model.Link, error)
}

type ObjectInfo struct {
	Path string

	Size *int64

	LastModified *time.Time
}

type ObjectReader interface {
	io.Reader

	Length() int64
}
