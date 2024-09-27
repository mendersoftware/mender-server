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
import { listItemMapper } from '@northern.tech/store/utils';
import { createSelector } from '@reduxjs/toolkit';

const getSelectedReleaseId = state => state.releases.selectedRelease;
export const getReleasesById = state => state.releases.byId;
export const getReleaseTags = state => state.releases.tags;
export const getReleaseListState = state => state.releases.releasesList;
const getListedReleases = state => state.releases.releasesList.releaseIds;
export const getUpdateTypes = state => state.releases.updateTypes;

const getReleaseMappingDefaults = () => ({});
export const getReleasesList = createSelector([getReleasesById, getListedReleases, getReleaseMappingDefaults], listItemMapper);

export const getReleaseTagsById = createSelector([getReleaseTags], releaseTags => releaseTags.reduce((accu, key) => ({ ...accu, [key]: key }), {}));
export const getHasReleases = createSelector(
  [getReleaseListState, getReleasesById],
  ({ searchTotal, total }, byId) => !!(Object.keys(byId).length || total || searchTotal)
);

export const getSelectedRelease = createSelector([getReleasesById, getSelectedReleaseId], (byId, id) => byId[id] ?? {});
