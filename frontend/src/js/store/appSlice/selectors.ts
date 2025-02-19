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

import { RootState } from '../store';

export const getDocsVersion = (state: RootState) => state.app.docsVersion;
export const getFeatures = (state: RootState) => state.app.features;
export const getFullVersionInformation = (state: RootState) => state.app.versionInformation;
export const getSearchState = (state: RootState) => state.app.searchState;
export const getUploads = (state: RootState) => state.app.uploadsById;
export const getSnackbar = (state: RootState) => state.app.snackbar;
export const getHostAddress = (state: RootState) => state.app.hostAddress;
export const getHostedAnnouncement = (state: RootState) => state.app.hostedAnnouncement;
export const getRecaptchaKey = (state: RootState) => state.app.recaptchaSiteKey;
export const getStripeKey = (state: RootState) => state.app.stripeAPIKey;
export const getTrackerCode = (state: RootState) => state.app.trackerCode;
export const getIsFirstLogin = (state: RootState) => state.app.firstLoginAfterSignup;
export const getFeedbackProbability = (state: RootState) => state.app.feedbackProbability;

export const getSearchedDevices = createSelector([getSearchState], ({ deviceIds }) => deviceIds);
export const getVersionInformation = createSelector([getFullVersionInformation, getFeatures], ({ Integration, ...remainder }, { isHosted }) =>
  isHosted && Integration !== 'next' ? remainder : { ...remainder, Integration }
);
export const getIsPreview = createSelector([getFullVersionInformation], ({ Integration }) => versionCompare(Integration, 'next') > -1);
