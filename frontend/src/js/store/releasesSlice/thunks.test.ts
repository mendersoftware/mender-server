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
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { defaultState } from '../../../../tests/mockData';
import { mockAbortController } from '../../../../tests/setupTests';
import { actions as appActions } from '../appSlice';
import {
  createArtifact,
  editArtifact,
  getArtifactInstallCount,
  getArtifactUrl,
  getExistingReleaseTags,
  getRelease,
  getReleases,
  getUpdateTypes,
  removeArtifact,
  removeRelease,
  selectRelease,
  setReleaseTags,
  setReleasesListState,
  updateReleaseInfo,
  uploadArtifact
} from './thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

const retrievedReleaseIds = [
  'release-999',
  'release-998',
  'release-997',
  'release-996',
  'release-995',
  'release-994',
  'release-993',
  'release-992',
  'release-991',
  'release-990',
  'release-99',
  'release-989',
  'release-988',
  'release-987',
  'release-986',
  'release-985',
  'release-984',
  'release-983',
  'release-982',
  'release-981'
];

describe('release actions', () => {
  it('should retrieve a single release by name', async () => {
    const store = mockStore({ ...defaultState });
    store.clearActions();
    const expectedActions = [
      { type: getRelease.pending.type },
      { type: actions.receiveRelease.type, payload: defaultState.releases.byId.r1 },
      { type: getRelease.fulfilled.type }
    ];
    await store.dispatch(getRelease(defaultState.releases.byId.r1.name));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve a list of releases', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getReleases.pending.type },
      { type: actions.receiveReleases.type, payload: defaultState.releases.byId },
      {
        type: actions.setReleaseListState.type,
        payload: { ...defaultState.releases.releasesList, releaseIds: ['release-1'], total: 5000 }
      },
      { type: getReleases.fulfilled.type }
    ];
    await store.dispatch(getReleases({ perPage: 1, sort: { direction: 'asc', key: 'name' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve a search filtered list of releases', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getReleases.pending.type },
      { type: actions.receiveReleases.type, payload: defaultState.releases.byId },
      {
        type: actions.setReleaseListState.type,
        payload: {
          ...defaultState.releases.releasesList,
          releaseIds: retrievedReleaseIds,
          searchTotal: 1234
        }
      },
      { type: getReleases.fulfilled.type }
    ];
    await store.dispatch(getReleases({ searchTerm: 'something' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve a deployment creation search filtered list of releases', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getReleases.pending.type },
      { type: actions.receiveReleases.type, payload: defaultState.releases.byId },
      {
        type: actions.setReleaseListState.type,
        payload: {
          ...defaultState.releases.releasesList,
          searchedIds: [
            'release-999',
            'release-998',
            'release-997',
            'release-996',
            'release-995',
            'release-994',
            'release-993',
            'release-992',
            'release-991',
            'release-990'
          ]
        }
      },
      { type: getReleases.fulfilled.type }
    ];
    await store.dispatch(getReleases({ perPage: 10, searchOnly: true, searchTerm: 'something' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve the device installation base for an artifact', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getArtifactInstallCount.pending.type },
      {
        type: actions.receiveRelease.type,
        payload: {
          ...defaultState.releases.byId.r1,
          artifacts: [{ ...defaultState.releases.byId.r1.artifacts[0], installCount: 0 }]
        }
      },
      { type: getArtifactInstallCount.fulfilled.type }
    ];
    await store.dispatch(getArtifactInstallCount('art1')).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should retrieve the download url for an artifact', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getArtifactUrl.pending.type },
      {
        type: actions.receiveRelease.type,
        payload: {
          ...defaultState.releases.byId.r1,
          artifacts: [
            {
              ...defaultState.releases.byId.r1.artifacts[0],
              url: 'https://testlocation.com/artifact.mender'
            }
          ]
        }
      },
      { type: getArtifactUrl.fulfilled.type }
    ];
    await store.dispatch(getArtifactUrl('art1')).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should select a release by name', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: selectRelease.pending.type },
      { type: actions.selectedRelease.type, payload: defaultState.releases.byId.r1.name },
      { type: getRelease.pending.type },
      { type: actions.receiveRelease.type, payload: defaultState.releases.byId.r1 },
      { type: getRelease.fulfilled.type },
      { type: selectRelease.fulfilled.type }
    ];
    await store.dispatch(selectRelease(defaultState.releases.byId.r1.name));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow creating an artifact', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: createArtifact.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Generating artifact' },
      {
        type: appActions.initUpload.type,
        payload: {
          id: 'mock-uuid',
          upload: { cancelSource: mockAbortController, name: 'createdRelease', size: undefined, uploadProgress: 0 }
        }
      },
      { type: appActions.uploadProgress.type, payload: { id: 'mock-uuid', progress: 100 } },
      { type: appActions.setSnackbar.type, payload: 'Upload successful' },
      { type: appActions.cleanUpUpload.type, payload: 'mock-uuid' },
      { type: createArtifact.fulfilled.type },
      { type: getReleases.pending.type },
      { type: selectRelease.pending.type },
      { type: actions.selectedRelease.type, payload: 'createdRelease' },
      { type: getReleases.pending.type }
    ];
    await store.dispatch(
      createArtifact({ file: { name: 'createdRelease', some: 'thing', someList: ['test', 'more'], complex: { objectThing: 'yes' } }, meta: 'filethings' })
    );
    jest.runAllTimers();
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should support editing artifact information', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: editArtifact.pending.type },
      {
        type: actions.receiveRelease.type,
        payload: {
          ...defaultState.releases.byId.r1,
          artifacts: [{ ...defaultState.releases.byId.r1.artifacts[0], description: 'something new' }]
        }
      },
      { type: appActions.setSnackbar.type, payload: 'Artifact details were updated successfully.' },
      { type: getReleases.pending.type },
      { type: selectRelease.pending.type },
      { type: actions.selectedRelease.type, payload: defaultState.releases.byId.r1.name },
      { type: getReleases.pending.type },
      { type: actions.receiveRelease.type, payload: defaultState.releases.byId.r1 },
      { type: actions.receiveRelease.type, payload: defaultState.releases.byId.r1 },
      { type: getReleases.fulfilled.type },
      { type: getReleases.fulfilled.type },
      { type: selectRelease.fulfilled.type },
      { type: editArtifact.fulfilled.type }
    ];
    await store.dispatch(editArtifact({ id: defaultState.releases.byId.r1.artifacts[0].id, body: { description: 'something new' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should support uploading .mender artifact files', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: uploadArtifact.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Uploading artifact' },
      {
        type: appActions.initUpload.type,
        payload: { id: 'mock-uuid', upload: { cancelSource: mockAbortController, name: defaultState.releases.byId.r1.name, size: 1234, uploadProgress: 0 } }
      },
      { type: appActions.uploadProgress.type, payload: { id: 'mock-uuid', progress: 100 } },
      { type: appActions.setSnackbar.type, payload: 'Upload successful' },
      { type: getReleases.pending.type },
      { type: actions.receiveReleases.type, payload: defaultState.releases.byId },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList, releaseIds: retrievedReleaseIds, total: 5000 } },
      { type: getReleases.fulfilled.type },
      { type: appActions.cleanUpUpload.type, payload: 'mock-uuid' },
      { type: uploadArtifact.fulfilled.type }
    ];
    await store.dispatch(uploadArtifact({ file: { name: defaultState.releases.byId.r1.name, size: 1234 }, meta: { description: 'new artifact to upload' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should remove an artifact by name', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: removeArtifact.pending.type },
      { type: actions.removeRelease.type, payload: defaultState.releases.byId.r1.name },
      { type: setReleasesListState.pending.type },
      { type: getReleases.pending.type },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList, isLoading: true, releaseIds: [], total: 0 } },
      { type: actions.receiveReleases.type, payload: defaultState.releases.byId },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList, releaseIds: retrievedReleaseIds, total: 5000 } },
      { type: getReleases.fulfilled.type },
      { type: setReleasesListState.pending.type },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList } },
      { type: setReleasesListState.fulfilled.type },
      { type: setReleasesListState.fulfilled.type },
      { type: removeArtifact.fulfilled.type }
    ];
    await store.dispatch(removeArtifact('art1'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should remove a release by name', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: removeRelease.pending.type },
      { type: removeArtifact.pending.type },
      { type: actions.removeRelease.type, payload: defaultState.releases.byId.r1.name },
      { type: setReleasesListState.pending.type },
      { type: getReleases.pending.type },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList, isLoading: true, releaseIds: [], total: 0 } },
      { type: actions.receiveReleases.type, payload: defaultState.releases.byId },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList, releaseIds: retrievedReleaseIds, total: 5000 } },
      { type: getReleases.fulfilled.type },
      { type: setReleasesListState.pending.type },
      { type: actions.setReleaseListState.type, payload: { ...defaultState.releases.releasesList } },
      { type: setReleasesListState.fulfilled.type },
      { type: setReleasesListState.fulfilled.type },
      { type: removeArtifact.fulfilled.type },
      { type: selectRelease.pending.type },
      { type: actions.selectedRelease.type, payload: null },
      { type: selectRelease.fulfilled.type },
      { type: removeRelease.fulfilled.type }
    ];
    await store.dispatch(removeRelease(defaultState.releases.byId.r1.name));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve existing release tags', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getExistingReleaseTags.pending.type },
      { type: actions.receiveReleaseTags.type, payload: ['foo', 'bar'] },
      { type: getExistingReleaseTags.fulfilled.type }
    ];
    await store.dispatch(getExistingReleaseTags());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve existing update types', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getUpdateTypes.pending.type },
      { type: actions.receiveReleaseTypes.type, payload: ['single-file', 'not-this'] },
      { type: getUpdateTypes.fulfilled.type }
    ];
    await store.dispatch(getUpdateTypes());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow setting new release tags', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setReleaseTags.pending.type },
      {
        type: actions.receiveRelease.type,
        payload: { ...defaultState.releases.byId.r1, tags: ['foo', 'bar'] }
      },
      { type: appActions.setSnackbar.type, payload: 'Release tags were set successfully.' },
      { type: setReleaseTags.fulfilled.type }
    ];
    await store.dispatch(setReleaseTags({ name: defaultState.releases.byId.r1.name, tags: ['foo', 'bar'] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow extending the release info', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: updateReleaseInfo.pending.type },
      {
        type: actions.receiveRelease.type,
        payload: { ...defaultState.releases.byId.r1, notes: 'this & that' }
      },
      { type: appActions.setSnackbar.type, payload: 'Release details were updated successfully.' },
      { type: updateReleaseInfo.fulfilled.type }
    ];
    await store.dispatch(updateReleaseInfo({ name: defaultState.releases.byId.r1.name, info: { notes: 'this & that' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});
