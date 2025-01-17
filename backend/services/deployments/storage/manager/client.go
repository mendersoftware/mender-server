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

package manager

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	"github.com/mendersoftware/mender-server/services/deployments/storage/azblob"
	"github.com/mendersoftware/mender-server/services/deployments/storage/s3"
)

var (
	ErrInvalidProvider = errors.New("manager: invalid storage provider")
)

type client struct {
	defaultStorage storage.ObjectStorage
	providerMap    map[model.StorageType]storage.ObjectStorage
}

func New(
	ctx context.Context,
	defaultStore storage.ObjectStorage,
	s3Options *s3.Options,
	azOptions *azblob.Options,
) (storage.ObjectStorage, error) {
	var err error
	providerMap := make(map[model.StorageType]storage.ObjectStorage, 2)
	providerMap[model.StorageTypeAzure], err = azblob.NewEmpty(ctx, azOptions)
	if err != nil {
		return nil, err
	}
	providerMap[model.StorageTypeS3], err = s3.NewEmpty(ctx, s3Options)
	if err != nil {
		return nil, err
	}

	return &client{
		defaultStorage: defaultStore,
		providerMap:    providerMap,
	}, nil
}

func (c *client) clientFromContext(
	ctx context.Context,
) (objStore storage.ObjectStorage, err error) {
	var ok bool
	if settings, _ := storage.SettingsFromContext(ctx); settings != nil {
		if objStore, ok = c.providerMap[settings.Type]; !ok {
			err = ErrInvalidProvider
		}
	} else {
		objStore = c.defaultStorage
	}
	return objStore, err
}

func (c *client) HealthCheck(ctx context.Context) (err error) {
	var objStore storage.ObjectStorage
	objStore, err = c.clientFromContext(ctx)
	if err != nil {
		return err
	}
	return objStore.HealthCheck(ctx)
}

func (c *client) GetObject(ctx context.Context, path string) (io.ReadCloser, error) {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return objStore.GetObject(ctx, path)
}

func (c *client) PutObject(ctx context.Context, path string, src io.Reader) error {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return err
	}

	return objStore.PutObject(ctx, path, src)
}

func (c *client) DeleteObject(ctx context.Context, path string) error {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return err
	}
	return objStore.DeleteObject(ctx, path)
}

func (c *client) StatObject(ctx context.Context, path string) (*storage.ObjectInfo, error) {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return objStore.StatObject(ctx, path)
}

func (c *client) GetRequest(
	ctx context.Context,
	path string,
	filename string,
	duration time.Duration,
	public bool,
) (*model.Link, error) {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return objStore.GetRequest(ctx, path, filename, duration, public)
}

func (c *client) DeleteRequest(
	ctx context.Context,
	path string,
	duration time.Duration,
	public bool,
) (*model.Link, error) {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return objStore.DeleteRequest(ctx, path, duration, public)
}

func (c *client) PutRequest(
	ctx context.Context,
	path string,
	duration time.Duration,
	public bool,
) (*model.Link, error) {
	objStore, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return objStore.PutRequest(ctx, path, duration, public)
}
