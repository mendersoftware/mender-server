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
	"path"
	"time"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func (d *Deployments) cleanupExpiredLink(
	ctx context.Context,
	link model.UploadLink,
	now time.Time,
) (err error) {
	switch link.Status {
	case model.LinkStatusProcessing:
		if link.UpdatedTS.Before(now.Add(-inprogressIdleTime)) {
			err = d.db.UpdateUploadIntentStatus(
				ctx,
				link.ArtifactID,
				model.LinkStatusProcessing,
				model.LinkStatusPending,
			)
			if err == store.ErrNotFound {
				err = nil
			}
		}
		// TODO: Call deployments API to restart processing
		// TODO: Increment retry counter to avoid infinite loop

	case model.LinkStatusAborted,
		model.LinkStatusCompleted,
		model.LinkStatusPending:
		objectPath := link.ArtifactID + fileSuffixTmp
		if link.TenantID != "" {
			objectPath = path.Join(link.TenantID, objectPath)
		}
		err = d.objectStorage.DeleteObject(ctx, objectPath)
		if err != nil && err != storage.ErrObjectNotFound {
			break
		}
		statusNew := link.Status
		if statusNew == model.LinkStatusPending {
			statusNew = model.LinkStatusAborted
		}
		statusNew |= model.LinkStatusProcessedBit
		err = d.db.UpdateUploadIntentStatus(
			ctx,
			link.ArtifactID,
			link.Status,
			statusNew,
		)
	}
	return err
}

func (d *Deployments) CleanupExpiredUploads(
	ctx context.Context, interval, jitter time.Duration,
) error {
	var (
		err error
		tc  <-chan time.Time
		run bool = true
	)
	if interval > 0 {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		tc = ticker.C
	} else {
		c := make(chan time.Time)
		close(c)
		tc = c
	}
	var it store.Iterator[model.UploadLink]
	defer func() {
		if it != nil {
			it.Close(ctx)
		}
	}()

	for run && err == nil {
		now := time.Now().Add(-jitter)
		it, err = d.db.FindUploadLinks(ctx, now)
		if err != nil {
			break
		}
		for run && err == nil {
			var link model.UploadLink
			run, err = it.Next(ctx)
			if !run {
				break
			}
			err = it.Decode(&link)
			if err != nil {
				break
			}
			err = d.cleanupExpiredLink(ctx, link, now)
		}
		if err != nil && err != store.ErrNotFound {
			break
		}
		err = it.Close(ctx)
		it = nil
		select {
		case <-ctx.Done():
			err = ctx.Err()

		case _, run = <-tc:
		}
	}
	return err
}
