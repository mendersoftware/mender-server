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

import { defaultState } from '../../../tests/mockData';
import { mockAbortController } from '../../../tests/setupTests';
import * as AppConstants from '../constants/appConstants';
import * as ReleaseConstants from '../constants/releaseConstants';
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
  updateReleaseInfo,
  uploadArtifact
} from './releaseActions';

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
    const expectedActions = [{ type: ReleaseConstants.RECEIVE_RELEASE, release: defaultState.releases.byId.r1 }];
    await store.dispatch(getRelease(defaultState.releases.byId.r1.name));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve a list of releases', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: ReleaseConstants.RECEIVE_RELEASES, releases: defaultState.releases.byId },
      {
        type: ReleaseConstants.SET_RELEASES_LIST_STATE,
        value: { ...defaultState.releases.releasesList, releaseIds: ['release-1'], total: 5000 }
      }
    ];
    await store.dispatch(getReleases({ perPage: 1, sort: { direction: 'asc', key: 'name' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve a search filtered list of releases', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: ReleaseConstants.RECEIVE_RELEASES, releases: defaultState.releases.byId },
      {
        type: ReleaseConstants.SET_RELEASES_LIST_STATE,
        value: {
          ...defaultState.releases.releasesList,
          releaseIds: retrievedReleaseIds,
          searchTotal: 1234
        }
      }
    ];
    await store.dispatch(getReleases({ searchTerm: 'something' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve a deployment creation search filtered list of releases', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: ReleaseConstants.RECEIVE_RELEASES, releases: defaultState.releases.byId },
      {
        type: ReleaseConstants.SET_RELEASES_LIST_STATE,
        value: {
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
      }
    ];
    await store.dispatch(getReleases({ perPage: 10, searchOnly: true, searchTerm: 'something' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve the device installation base for an artifact', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      {
        type: ReleaseConstants.RECEIVE_RELEASE,
        release: {
          ...defaultState.releases.byId.r1,
          artifacts: [{ ...defaultState.releases.byId.r1.artifacts[0], installCount: 0 }]
        }
      }
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
      {
        type: ReleaseConstants.ARTIFACTS_SET_ARTIFACT_URL,
        release: {
          ...defaultState.releases.byId.r1,
          artifacts: [
            {
              ...defaultState.releases.byId.r1.artifacts[0],
              url: 'https://testlocation.com/artifact.mender'
            }
          ]
        }
      }
    ];
    await store.dispatch(getArtifactUrl('art1')).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should select a release by name', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(selectRelease(defaultState.releases.byId.r1.name));
    const expectedActions = [
      { type: ReleaseConstants.SELECTED_RELEASE, release: defaultState.releases.byId.r1.name },
      { type: ReleaseConstants.RECEIVE_RELEASE, release: defaultState.releases.byId.r1 }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow creating an artifact', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Generating artifact' } },
      {
        type: AppConstants.UPLOAD_PROGRESS,
        uploads: { 'mock-uuid': { cancelSource: mockAbortController, name: undefined, size: undefined, uploadProgress: 0 } }
      },
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Upload successful' } },
      { type: AppConstants.UPLOAD_PROGRESS, uploads: {} },
      { type: ReleaseConstants.SELECTED_RELEASE, release: 'createdRelease' }
    ];
    await store.dispatch(createArtifact({ name: 'createdRelease', some: 'thing', someList: ['test', 'more'], complex: { objectThing: 'yes' } }, 'filethings'));
    jest.runAllTimers();
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should support editing artifact information', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      {
        type: ReleaseConstants.UPDATED_ARTIFACT,
        release: {
          ...defaultState.releases.byId.r1,
          artifacts: [{ ...defaultState.releases.byId.r1.artifacts[0], description: 'something new' }]
        }
      },
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Artifact details were updated successfully.' } },
      { type: ReleaseConstants.SELECTED_RELEASE, release: defaultState.releases.byId.r1.name },
      { type: ReleaseConstants.RECEIVE_RELEASE, release: defaultState.releases.byId.r1 },
      { type: ReleaseConstants.RECEIVE_RELEASE, release: defaultState.releases.byId.r1 }
    ];
    await store.dispatch(editArtifact(defaultState.releases.byId.r1.artifacts[0].id, { description: 'something new' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should support uploading .mender artifact files', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Uploading artifact' } },
      {
        type: AppConstants.UPLOAD_PROGRESS,
        uploads: { 'mock-uuid': { cancelSource: mockAbortController, name: undefined, size: 1234, uploadProgress: 0 } }
      },
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Upload successful' } },
      { type: ReleaseConstants.SELECTED_RELEASE, release: defaultState.releases.byId.r1.name },
      { type: ReleaseConstants.RECEIVE_RELEASES, releases: defaultState.releases.byId },
      { type: ReleaseConstants.SET_RELEASES_LIST_STATE, value: { ...defaultState.releases.releasesList, releaseIds: retrievedReleaseIds, total: 5000 } },
      { type: ReleaseConstants.RECEIVE_RELEASE, release: defaultState.releases.byId.r1 },
      { type: AppConstants.UPLOAD_PROGRESS, uploads: {} }
    ];
    await store.dispatch(uploadArtifact({ description: 'new artifact to upload', name: defaultState.releases.byId.r1.name }, { size: 1234 }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should remove an artifact by name', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: ReleaseConstants.RELEASE_REMOVED, release: defaultState.releases.byId.r1.name },
      { type: ReleaseConstants.SET_RELEASES_LIST_STATE, value: { ...defaultState.releases.releasesList, isLoading: true, releaseIds: [], total: 0 } },
      { type: ReleaseConstants.RECEIVE_RELEASES, releases: defaultState.releases.byId },
      {
        type: ReleaseConstants.SET_RELEASES_LIST_STATE,
        value: { ...defaultState.releases.releasesList, releaseIds: retrievedReleaseIds, total: 5000 }
      },
      { type: ReleaseConstants.SET_RELEASES_LIST_STATE, value: { ...defaultState.releases.releasesList } }
    ];
    await store.dispatch(removeArtifact('art1'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should remove a release by name', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: ReleaseConstants.RELEASE_REMOVED, release: defaultState.releases.byId.r1.name },
      { type: ReleaseConstants.SET_RELEASES_LIST_STATE, value: { ...defaultState.releases.releasesList, isLoading: true, releaseIds: [], total: 0 } },
      { type: ReleaseConstants.RECEIVE_RELEASES, releases: defaultState.releases.byId },
      {
        type: ReleaseConstants.SET_RELEASES_LIST_STATE,
        value: { ...defaultState.releases.releasesList, releaseIds: retrievedReleaseIds, total: 5000 }
      },
      { type: ReleaseConstants.SET_RELEASES_LIST_STATE, value: { ...defaultState.releases.releasesList } },
      { type: ReleaseConstants.SELECTED_RELEASE, release: null }
    ];
    await store.dispatch(removeRelease(defaultState.releases.byId.r1.name));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve existing release tags', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: ReleaseConstants.RECEIVE_RELEASE_TAGS, tags: ['foo', 'bar'] }];
    await store.dispatch(getExistingReleaseTags());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should retrieve existing release tags', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: ReleaseConstants.RECEIVE_RELEASE_TYPES, types: ['single-file', 'not-this'] }];
    await store.dispatch(getUpdateTypes());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow setting new release tags', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      {
        type: ReleaseConstants.RECEIVE_RELEASE,
        release: { ...defaultState.releases.byId.r1, tags: ['foo', 'bar'] }
      },
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Release tags were set successfully.' } }
    ];
    await store.dispatch(setReleaseTags(defaultState.releases.byId.r1.name, ['foo', 'bar']));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow extending the release info', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      {
        type: ReleaseConstants.RECEIVE_RELEASE,
        release: { ...defaultState.releases.byId.r1, notes: 'this & that' }
      },
      { type: AppConstants.SET_SNACKBAR, snackbar: { message: 'Release details were updated successfully.' } }
    ];
    await store.dispatch(updateReleaseInfo(defaultState.releases.byId.r1.name, { notes: 'this & that' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});
