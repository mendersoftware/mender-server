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
import { versionCompare } from '@northern.tech/utils/helpers';
import { createSelector } from '@reduxjs/toolkit';

const getAppDocsVersion = state => state.app.docsVersion;
export const getFeatures = state => state.app.features;
export const getFullVersionInformation = state => state.app.versionInformation;

export const getDocsVersion = createSelector([getAppDocsVersion, getFeatures], (appDocsVersion, { isHosted }) => {
  // if hosted, use latest docs version
  const docsVersion = appDocsVersion ? `${appDocsVersion}/` : 'development/';
  return isHosted ? '' : docsVersion;
});

export const getSearchState = state => state.app.searchState;

export const getSearchedDevices = createSelector([getSearchState], ({ deviceIds }) => deviceIds);
export const getVersionInformation = createSelector([getFullVersionInformation, getFeatures], ({ Integration, ...remainder }, { isHosted }) =>
  isHosted && Integration !== 'next' ? remainder : { ...remainder, Integration }
);
export const getIsPreview = createSelector([getFullVersionInformation], ({ Integration }) => versionCompare(Integration, 'next') > -1);
