// Copyright 2020 Northern.tech AS
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
import { commonErrorHandler } from '@northern.tech/store/store';
import { searchDevices } from '@northern.tech/store/thunks';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { defaultState } from '../../../../tests/mockData';
import { actions as deviceActions } from '../devicesSlice';
import { getLatestReleaseInfo, setFirstLoginAfterSignup, setOfflineThreshold, setSearchState } from './thunks';

export const latestSaasReleaseTag = 'saas-v2023.05.02';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

/* eslint-disable sonarjs/no-identical-functions */
describe('app actions', () => {
  it('should handle different error message formats', async () => {
    const store = mockStore({ ...defaultState });
    const err = { response: { data: { error: { message: 'test' } } }, id: '123' };
    await expect(commonErrorHandler(err, 'testContext', store.dispatch)).rejects.toEqual(err);
    const expectedActions = [
      {
        type: actions.setSnackbar.type,
        payload: { message: `testContext ${err.response.data.error.message}`, action: 'Copy to clipboard' }
      }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should not get the latest release info when not hosted', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: getLatestReleaseInfo.pending.type }, { type: getLatestReleaseInfo.fulfilled.type }];
    await store.dispatch(getLatestReleaseInfo());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
  });
  it('should get the latest release info when hosted', async () => {
    const store = mockStore({
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          isHosted: true
        }
      }
    });
    const expectedActions = [
      { type: getLatestReleaseInfo.pending.type },
      {
        type: actions.setVersionInformation.type,
        payload: { backend: latestSaasReleaseTag, GUI: latestSaasReleaseTag, Integration: '1.2.3', 'Mender-Client': '3.2.1', 'Mender-Artifact': '1.3.7' }
      },
      { type: getLatestReleaseInfo.fulfilled.type }
    ];
    await store.dispatch(getLatestReleaseInfo());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should store first login after Signup', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setFirstLoginAfterSignup.pending.type },
      { type: actions.setFirstLoginAfterSignup.type, payload: true },
      { type: setFirstLoginAfterSignup.fulfilled.type }
    ];
    await store.dispatch(setFirstLoginAfterSignup(true));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should calculate yesterdays timestamp', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setOfflineThreshold.pending.type },
      { type: actions.setOfflineThreshold.type, payload: '2019-01-12T13:00:00.900Z' },
      { type: setOfflineThreshold.fulfilled.type }
    ];
    await store.dispatch(setOfflineThreshold());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should handle searching', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setSearchState.pending.type },
      { type: searchDevices.pending.type },
      { type: actions.setSearchState.type, payload: { ...defaultState.app.searchState, isSearching: true, searchTerm: 'next!' } },
      { type: deviceActions.receivedDevices.type, payload: {} },
      { type: searchDevices.fulfilled.type },
      { type: actions.setSearchState.type, payload: { deviceIds: [], isSearching: false, searchTotal: 0 } },
      { type: setSearchState.fulfilled.type }
    ];
    await store.dispatch(setSearchState({ searchTerm: 'next!' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});
