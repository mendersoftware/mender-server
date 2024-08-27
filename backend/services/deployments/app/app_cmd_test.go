// Copyright 2023 Northern.tech AS
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
	"context"
	"errors"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	mstorage "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	mstore "github.com/mendersoftware/mender-server/services/deployments/store/mocks"
)

type ArrayIterator[T interface{}] struct {
	arr    []T
	idx    int
	closed chan struct{}
}

func NewArrayIterator[T interface{}](arr []T) *ArrayIterator[T] {
	return &ArrayIterator[T]{
		arr:    arr,
		idx:    -1,
		closed: make(chan struct{}),
	}
}

func (it *ArrayIterator[T]) Next(ctx context.Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if it.idx >= len(it.arr) {
		return false, errors.New("end of iterator")
	}
	it.idx++
	return it.idx < len(it.arr), nil
}

func (it *ArrayIterator[T]) Decode(elem *T) error {
	if it.idx < 0 || it.idx >= len(it.arr) {
		return errors.New("iterator out of bounds")
	} else if elem == nil {
		return errors.New("decode nil value")
	}
	*elem = it.arr[it.idx]
	return nil
}

func (it *ArrayIterator[T]) Close(ctx context.Context) error {
	select {
	case <-it.closed:
		return errors.New("iterator already closed")
	default:
		close(it.closed)
	}
	return ctx.Err()
}

func TestCleanupExpiredUploads(t *testing.T) {
	t.Parallel()

	t.Run("single-shot/ok", func(t *testing.T) {
		const (
			jitter = time.Second
		)
		ctx := context.Background()
		links := []model.UploadLink{{
			ArtifactID: "94a89c91-a905-4c3a-8bfa-62a362851c1f",
			Link: model.Link{
				Uri:      "http://localhost:8080",
				TenantID: "123456789012345678901234",
				Expire:   time.Now().Add(-time.Hour * 24),
			},
			UpdatedTS: time.Now().Add(-time.Hour),
			Status:    model.LinkStatusCompleted,
		}, {
			ArtifactID: "624836fd-29f5-474e-b101-5482b67c9204",
			Link: model.Link{
				Uri:    "http://localhost:8080",
				Expire: time.Now().Add(-time.Hour * 12),
			},
			UpdatedTS: time.Now().Add(-time.Hour * 2),
			Status:    model.LinkStatusPending,
		}, {
			ArtifactID: "1ea293ad-c94b-44b7-a137-af1dd9d6b126",
			Link: model.Link{
				Uri:    "http://localhost:8080",
				Expire: time.Now().Add(-time.Hour * 12),
			},
			UpdatedTS: time.Now().Add(-inprogressIdleTime * 3),
			Status:    model.LinkStatusProcessing,
		}}

		database := new(mstore.DataStore)
		objectStore := new(mstorage.ObjectStorage)
		defer database.AssertExpectations(t)
		defer objectStore.AssertExpectations(t)

		database.On("FindUploadLinks", ctx, mock.Anything).
			Run(func(args mock.Arguments) {
				exp := args.Get(1).(time.Time)
				assert.WithinDuration(t, time.Now().Add(-jitter), exp, time.Minute)
			}).
			Return(NewArrayIterator[model.UploadLink](links), nil).
			Once()

		for _, link := range links {
			switch status := link.Status; status {
			case model.LinkStatusProcessing:
				if link.UpdatedTS.Before(time.Now().Add(-inprogressIdleTime)) {
					database.On("UpdateUploadIntentStatus",
						ctx, link.ArtifactID,
						model.LinkStatusProcessing, model.LinkStatusPending).
						Return(store.ErrNotFound).
						Once()
				}

			case model.LinkStatusAborted, model.LinkStatusCompleted, model.LinkStatusPending:
				var errDelete error
				statusNew := status | model.LinkStatusProcessedBit
				if status == model.LinkStatusPending {
					statusNew = model.LinkStatusAborted | model.LinkStatusProcessedBit
					errDelete = storage.ErrObjectNotFound
				}
				objectStore.On("DeleteObject",
					ctx,
					path.Join(link.TenantID, link.ArtifactID)+fileSuffixTmp).
					Return(errDelete).
					Once()
				database.On("UpdateUploadIntentStatus",
					ctx, link.ArtifactID,
					link.Status, statusNew).
					Return(nil).
					Once()
			}
		}

		app := NewDeployments(database, objectStore, 0, false)

		err := app.CleanupExpiredUploads(ctx, 0, jitter)
		assert.NoError(t, err)
	})
	t.Run("periodic/context canceled", func(t *testing.T) {
		const (
			jitter = time.Second
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		links := []model.UploadLink{}
		iterator := NewArrayIterator[model.UploadLink](links)

		database := new(mstore.DataStore)
		objectStore := new(mstorage.ObjectStorage)
		defer database.AssertExpectations(t)
		defer objectStore.AssertExpectations(t)

		database.On("FindUploadLinks", ctx, mock.Anything).
			Run(func(args mock.Arguments) {
				exp := args.Get(1).(time.Time)
				assert.WithinDuration(t, time.Now().Add(-jitter), exp, time.Minute)
			}).
			Return(iterator, nil).
			Once()

		app := NewDeployments(database, objectStore, 0, false)

		go func() {
			select {
			case <-iterator.closed:
			case <-time.After(time.Second * 10):
			}
			cancel()
		}()
		err := app.CleanupExpiredUploads(ctx, time.Hour, jitter)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("error/delete object internal error", func(t *testing.T) {
		const (
			jitter = time.Second
		)
		ctx := context.Background()
		links := []model.UploadLink{{
			ArtifactID: "94a89c91-a905-4c3a-8bfa-62a362851c1f",
			Link: model.Link{
				Uri:      "http://localhost:8080",
				TenantID: "123456789012345678901234",
				Expire:   time.Now().Add(-time.Hour * 24),
			},
			UpdatedTS: time.Now().Add(-time.Hour),
			Status:    model.LinkStatusCompleted,
		}}

		database := new(mstore.DataStore)
		objectStore := new(mstorage.ObjectStorage)
		defer database.AssertExpectations(t)
		defer objectStore.AssertExpectations(t)

		database.On("FindUploadLinks", ctx, mock.Anything).
			Run(func(args mock.Arguments) {
				exp := args.Get(1).(time.Time)
				assert.WithinDuration(t, time.Now().Add(-jitter), exp, time.Minute)
			}).
			Return(NewArrayIterator[model.UploadLink](links), nil).
			Once()

		errInternal := errors.New("internal error")
		for _, link := range links {
			objectStore.On("DeleteObject",
				ctx,
				path.Join(link.TenantID, link.ArtifactID)+fileSuffixTmp).
				Return(errInternal).
				Once()
		}

		app := NewDeployments(database, objectStore, 0, false)

		err := app.CleanupExpiredUploads(ctx, 0, jitter)
		assert.ErrorIs(t, err, errInternal)
	})
	t.Run("error/database find upload links", func(t *testing.T) {
		const (
			jitter = time.Second
		)
		ctx := context.Background()
		database := new(mstore.DataStore)
		objectStore := new(mstorage.ObjectStorage)
		defer database.AssertExpectations(t)
		defer objectStore.AssertExpectations(t)

		errInternal := errors.New("internal error")
		database.On("FindUploadLinks", ctx, mock.Anything).
			Run(func(args mock.Arguments) {
				exp := args.Get(1).(time.Time)
				assert.WithinDuration(t, time.Now().Add(-jitter), exp, time.Minute)
			}).
			Return(nil, errInternal).
			Once()

		app := NewDeployments(database, objectStore, 0, false)

		err := app.CleanupExpiredUploads(ctx, 0, jitter)
		assert.ErrorIs(t, err, errInternal)
	})
}
