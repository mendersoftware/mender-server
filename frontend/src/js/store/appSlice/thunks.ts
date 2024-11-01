// Copyright 2019 Northern.tech AS
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
// @ts-nocheck
import GeneralApi from '@northern.tech/store/api/general-api';
import { getOfflineThresholdSettings } from '@northern.tech/store/selectors';
import { searchDevices } from '@northern.tech/store/thunks';
import { extractErrorMessage, getComparisonCompatibleVersion } from '@northern.tech/store/utils';
import { deepCompare } from '@northern.tech/utils/helpers';
import { createAsyncThunk } from '@reduxjs/toolkit';
import Cookies from 'universal-cookie';

import { actions, sliceName } from '.';
import { getFeatures, getSearchState } from './selectors';

const cookies = new Cookies();

/*
  General
*/
export const setFirstLoginAfterSignup = createAsyncThunk(`${sliceName}/setFirstLoginAfterSignup`, (firstLoginAfterSignup, { dispatch }) => {
  cookies.set('firstLoginAfterSignup', !!firstLoginAfterSignup, { maxAge: 60, path: '/', domain: '.mender.io', sameSite: false });
  dispatch(actions.setFirstLoginAfterSignup(!!firstLoginAfterSignup));
});

const dateFunctionMap = {
  getDays: 'getDate',
  setDays: 'setDate'
};
export const setOfflineThreshold = createAsyncThunk(`${sliceName}/setOfflineThreshold`, (_, { dispatch, getState }) => {
  const { interval, intervalUnit } = getOfflineThresholdSettings(getState());
  const today = new Date();
  const intervalName = `${intervalUnit.charAt(0).toUpperCase()}${intervalUnit.substring(1)}`;
  const setter = dateFunctionMap[`set${intervalName}`] ?? `set${intervalName}`;
  const getter = dateFunctionMap[`get${intervalName}`] ?? `get${intervalName}`;
  today[setter](today[getter]() - interval);
  let value;
  try {
    value = today.toISOString();
  } catch {
    return Promise.resolve(dispatch(actions.setSnackbar('There was an error saving the offline threshold, please check your settings.')));
  }
  return Promise.resolve(dispatch(actions.setOfflineThreshold(value)));
});

const versionRegex = new RegExp(/\d+\.\d+/);
const getLatestRelease = thing => {
  const latestKey = Object.keys(thing)
    .filter(key => versionRegex.test(key))
    .sort()
    .reverse()[0];
  return thing[latestKey];
};

const repoKeyMap = {
  integration: 'Integration',
  mender: 'Mender-Client',
  'mender-artifact': 'Mender-Artifact'
};

const deductSaasState = (latestRelease, guiTags, saasReleases) => {
  const latestGuiTag = guiTags.length ? guiTags[0].name : '';
  const latestSaasRelease = latestGuiTag.startsWith('saas-v') ? { date: latestGuiTag.split('-v')[1].replaceAll('.', '-'), tag: latestGuiTag } : saasReleases[0];
  return latestSaasRelease.date > latestRelease.release_date ? latestSaasRelease.tag : latestRelease.release;
};

export const getLatestReleaseInfo = createAsyncThunk(`${sliceName}/getLatestReleaseInfo`, (_, { dispatch, getState }) => {
  if (!getFeatures(getState()).isHosted) {
    return Promise.resolve();
  }
  return Promise.all([GeneralApi.get('/versions.json'), GeneralApi.get('/tags.json')])
    .catch(err => {
      console.log('init error:', extractErrorMessage(err));
      return Promise.resolve([{ data: {} }, { data: [] }]);
    })
    .then(([{ data }, { data: guiTags }]) => {
      if (!guiTags.length) {
        return Promise.resolve();
      }
      const { releases, saas } = data;
      const latestRelease = getLatestRelease(getLatestRelease(releases));
      const { latestRepos, latestVersions } = latestRelease.repos.reduce(
        (accu, item) => {
          if (repoKeyMap[item.name]) {
            accu.latestVersions[repoKeyMap[item.name]] = getComparisonCompatibleVersion(item.version);
          }
          accu.latestRepos[item.name] = getComparisonCompatibleVersion(item.version);
          return accu;
        },
        { latestVersions: { ...getState().app.versionInformation }, latestRepos: {} }
      );
      const info = deductSaasState(latestRelease, guiTags, saas);
      return Promise.resolve(
        dispatch(
          actions.setVersionInformation({
            ...latestVersions,
            backend: info,
            GUI: info,
            latestRelease: {
              releaseDate: latestRelease.release_date,
              repos: latestRepos
            }
          })
        )
      );
    });
});

export const setSearchState = createAsyncThunk(`${sliceName}/setSearchState`, (searchState, { dispatch, getState }) => {
  const currentState = getSearchState(getState());
  let nextState = {
    ...currentState,
    ...searchState,
    sort: {
      ...currentState.sort,
      ...searchState.sort
    }
  };
  let tasks = [];
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { isSearching: currentSearching, deviceIds: currentDevices, searchTotal: currentTotal, ...currentRequestState } = currentState;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { isSearching: nextSearching, deviceIds: nextDevices, searchTotal: nextTotal, ...nextRequestState } = nextState;
  if (nextRequestState.searchTerm && !deepCompare(currentRequestState, nextRequestState)) {
    nextState.isSearching = true;
    tasks.push(
      dispatch(searchDevices(nextState))
        .unwrap()
        .then(results => {
          const searchResult = results[results.length - 1];
          return dispatch(actions.setSearchState({ ...searchResult, isSearching: false }));
        })
        .catch(() => dispatch(actions.setSearchState({ isSearching: false, searchTotal: 0 })))
    );
  }
  tasks.push(dispatch(actions.setSearchState(nextState)));
  return Promise.all(tasks);
});
