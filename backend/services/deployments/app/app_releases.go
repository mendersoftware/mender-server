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
	"context"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

// Errors expected from App interface
var (
	ErrReleaseNotFound = errors.New("release not found")
)

func (d *Deployments) updateReleaseEditArtifact(
	ctx context.Context,
	artifactToEdit *model.Image,
) error {

	if artifactToEdit == nil {
		return ErrEmptyArtifact
	}
	return d.db.UpdateReleaseArtifactDescription(
		ctx,
		artifactToEdit,
		artifactToEdit.ArtifactMeta.Name,
	)
}

func (d *Deployments) updateRelease(
	ctx context.Context,
	artifactToAdd *model.Image,
	artifactToRemove *model.Image,
) error {
	name := ""
	if artifactToRemove != nil {
		name = artifactToRemove.ArtifactMeta.Name
	} else if artifactToAdd != nil {
		name = artifactToAdd.ArtifactMeta.Name
	} else {
		return ErrEmptyArtifact
	}

	return d.db.UpdateReleaseArtifacts(ctx, artifactToAdd, artifactToRemove, name)
}

func (d *Deployments) ListReleaseTags(ctx context.Context) (model.Tags, error) {
	tags, err := d.db.ListReleaseTags(ctx)
	if err != nil {
		log.FromContext(ctx).
			Errorf("failed to list release tags: %s", err)
		err = ErrModelInternal
	}
	return tags, err
}

func (d *Deployments) GetReleasesUpdateTypes(ctx context.Context) ([]string, error) {
	updateTypes, err := d.db.GetUpdateTypes(ctx)
	if err != nil {
		log.FromContext(ctx).
			Errorf("failed to list release update types: %s", err)
		err = ErrModelInternal
	}
	return updateTypes, err
}

func (d *Deployments) ReplaceReleaseTags(
	ctx context.Context,
	releaseName string,
	tags model.Tags,
) error {
	err := d.db.ReplaceReleaseTags(ctx, releaseName, tags)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			err = ErrReleaseNotFound

		case model.ErrTooManyTags, model.ErrTooManyUniqueTags:
			// pass

		default:
			// Rewrite internal errors
			log.FromContext(ctx).
				Errorf("failed to replace tags in database: %s", err.Error())
			err = ErrModelInternal
		}
	}
	return err
}

func (d *Deployments) UpdateRelease(
	ctx context.Context,
	releaseName string,
	release model.ReleasePatch,
) error {
	err := d.db.UpdateRelease(ctx, releaseName, release)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			err = ErrReleaseNotFound

		default:
			// Rewrite internal errors
			log.FromContext(ctx).
				Errorf("failed to update release in the database: %s", err.Error())
			err = ErrModelInternal
		}
	}
	return err
}

func (d *Deployments) DeleteReleases(
	ctx context.Context,
	releaseNames []string,
) ([]string, error) {
	ids, err := d.db.GetActiveDeploymentIDsByArtifactNames(ctx, releaseNames)
	if err != nil || len(ids) > 0 {
		return ids, err
	}
	if err := d.db.DeleteImagesByNames(ctx, releaseNames); err != nil {
		return ids, err
	}
	err = d.db.DeleteReleasesByNames(ctx, releaseNames)
	return ids, err
}

func (d *Deployments) GetRelease(
	ctx context.Context,
	releaseName string,
) (*model.Release, error) {
	release, err := d.db.GetRelease(ctx, releaseName)
	switch err {
	case nil:
		return release, nil
	case store.ErrNotFound:
		return nil, ErrReleaseNotFound
	default:
		// Rewrite internal errors
		log.FromContext(ctx).
			Errorf("failed to get release from database: %s", err.Error())
		return nil, ErrModelInternal
	}
}
