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
import storeActions from '@northern.tech/store/actions';
import GeneralApi from '@northern.tech/store/api/general-api';
import {
  DEVICE_LIST_DEFAULTS,
  SORTING_OPTIONS,
  TIMEOUTS,
  deploymentsApiUrl,
  deploymentsApiUrlV2,
  emptyFilter,
  headerNames
} from '@northern.tech/store/constants';
import { formatReleases } from '@northern.tech/store/locationutils';
import { getSearchEndpoint } from '@northern.tech/store/selectors';
import { commonErrorFallback, commonErrorHandler } from '@northern.tech/store/store';
import { convertDeviceListStateToFilters, progress } from '@northern.tech/store/utils';
import { customSort, deepCompare, duplicateFilter, extractSoftwareItem } from '@northern.tech/utils/helpers';
import { createAsyncThunk } from '@reduxjs/toolkit';
import { isCancel } from 'axios';
import { v4 as uuid } from 'uuid';

import { actions, sliceName } from '.';
import { ARTIFACT_GENERATION_TYPE } from './constants';
import { getReleasesById } from './selectors';

const { setSnackbar, initUpload, uploadProgress, cleanUpUpload } = storeActions;
const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const sortingDefaults = { direction: SORTING_OPTIONS.desc, key: 'modified' };

const flattenRelease = (release, stateRelease) => {
  const updatedArtifacts = release.artifacts?.sort(customSort(1, 'modified')) || [];
  const { artifacts, deviceTypes, modified } = updatedArtifacts.reduce(
    (accu, item) => {
      accu.deviceTypes.push(...item.device_types_compatible);
      const stateArtifact = stateRelease.artifacts?.find(releaseArtifact => releaseArtifact.id === item.id) || {};
      accu.modified = accu.modified ? accu.modified : item.modified;
      accu.artifacts.push({
        ...stateArtifact,
        ...item
      });
      return accu;
    },
    { artifacts: [], deviceTypes: [], modified: undefined }
  );
  return { ...stateRelease, ...release, artifacts, device_types_compatible: deviceTypes.filter(duplicateFilter), modified };
};

const reduceReceivedReleases = (releases, stateReleasesById) =>
  releases.reduce((accu, release) => {
    const stateRelease = stateReleasesById[release.name] || {};
    accu[release.name] = flattenRelease(release, stateRelease);
    return accu;
  }, {});

const findArtifactIndexInRelease = (releases, id) =>
  Object.values(releases).reduce(
    (accu, item) => {
      let index = item.artifacts.findIndex(releaseArtifact => releaseArtifact.id === id);
      if (index > -1) {
        accu = { release: item, index };
      }
      return accu;
    },
    { release: null, index: -1 }
  );

/* Artifacts */
export const getArtifactInstallCount = createAsyncThunk(`${sliceName}/getArtifactInstallCount`, (id, { dispatch, getState }) => {
  let { release, index } = findArtifactIndexInRelease(getReleasesById(getState()), id);
  if (!release || index === -1) {
    return;
  }
  const releaseArtifacts = [...release.artifacts];
  const artifact = releaseArtifacts[index];
  const { key, name, version } = extractSoftwareItem(artifact.artifact_provides) ?? {};
  const attribute = `${key}${name ? `.${name}` : ''}.version`;
  const { filterTerms } = convertDeviceListStateToFilters({
    filters: [{ ...emptyFilter, key: attribute, value: version, scope: 'inventory' }]
  });
  return GeneralApi.post(getSearchEndpoint(getState()), {
    page: 1,
    per_page: 1,
    filters: filterTerms,
    attributes: [{ scope: 'identity', attribute: 'status' }]
  })
    .catch(err => commonErrorHandler(err, `Retrieving artifact installation count failed:`, dispatch, commonErrorFallback))
    .then(({ headers }) => {
      let { release, index } = findArtifactIndexInRelease(getReleasesById(getState()), id);
      if (!release || index === -1) {
        return;
      }
      const installCount = Number(headers[headerNames.total]);
      const releaseArtifacts = [...release.artifacts];
      releaseArtifacts[index] = { ...releaseArtifacts[index], installCount };
      release = {
        ...release,
        artifacts: releaseArtifacts
      };
      return dispatch(actions.receiveRelease(release));
    });
});

export const getArtifactUrl = createAsyncThunk(`${sliceName}/getArtifactUrl`, (id, { dispatch, getState }) =>
  GeneralApi.get(`${deploymentsApiUrl}/artifacts/${id}/download`).then(response => {
    let { release, index } = findArtifactIndexInRelease(getReleasesById(getState()), id);
    if (!release || index === -1) {
      return dispatch(getReleases()).unwrap();
    }
    const releaseArtifacts = [...release.artifacts];
    releaseArtifacts[index] = {
      ...releaseArtifacts[index],
      url: response.data.uri
    };
    release = {
      ...release,
      artifacts: releaseArtifacts
    };
    return dispatch(actions.receiveRelease(release));
  })
);

export const createArtifact = createAsyncThunk(`${sliceName}/createArtifact`, ({ file, meta }, { dispatch }) => {
  let formData = Object.entries(meta).reduce((accu, [key, value]) => {
    if (Array.isArray(value)) {
      accu.append(key, value.join(','));
    } else if (value instanceof Object) {
      accu.append(key, JSON.stringify(value));
    } else {
      accu.append(key, value);
    }
    return accu;
  }, new FormData());
  formData.append('type', ARTIFACT_GENERATION_TYPE.SINGLE_FILE);
  formData.append('file', file);
  const uploadId = uuid();
  const cancelSource = new AbortController();
  return Promise.all([
    dispatch(setSnackbar('Generating artifact')),
    dispatch(initUpload({ id: uploadId, upload: { name: file.name, size: file.size, progress: 0, cancelSource } })),
    GeneralApi.upload(
      `${deploymentsApiUrl}/artifacts/generate`,
      formData,
      e => dispatch(uploadProgress({ id: uploadId, progress: progress(e) })),
      cancelSource.signal
    )
  ])
    .then(() => {
      setTimeout(() => {
        dispatch(getReleases()).unwrap();
        dispatch(selectRelease(file.name)).unwrap();
      }, TIMEOUTS.oneSecond);
      return Promise.resolve(dispatch(setSnackbar({ message: 'Upload successful', autoHideDuration: TIMEOUTS.fiveSeconds })));
    })
    .catch(err => {
      if (isCancel(err)) {
        return dispatch(setSnackbar({ message: 'The artifact generation has been cancelled', autoHideDuration: TIMEOUTS.fiveSeconds }));
      }
      return commonErrorHandler(err, `Artifact couldn't be generated.`, dispatch);
    })
    .finally(() => dispatch(cleanUpUpload(uploadId)));
});

export const uploadArtifact = createAsyncThunk(`${sliceName}/uploadArtifact`, ({ file, meta }, { dispatch }) => {
  let formData = new FormData();
  formData.append('size', file.size);
  formData.append('description', meta.description);
  formData.append('artifact', file);
  const uploadId = uuid();
  const cancelSource = new AbortController();
  return Promise.all([
    dispatch(setSnackbar('Uploading artifact')),
    dispatch(initUpload({ id: uploadId, upload: { name: file.name, size: file.size, progress: 0, cancelSource } })),
    GeneralApi.upload(`${deploymentsApiUrl}/artifacts`, formData, e => dispatch(uploadProgress({ id: uploadId, progress: progress(e) })), cancelSource.signal)
  ])
    .then(() => {
      const tasks = [dispatch(setSnackbar({ message: 'Upload successful', autoHideDuration: TIMEOUTS.fiveSeconds })), dispatch(getReleases()).unwrap()];
      if (meta.name) {
        tasks.push(dispatch(selectRelease(meta.name)).unwrap());
      }
      return Promise.all(tasks);
    })
    .catch(err => {
      if (isCancel(err)) {
        return dispatch(setSnackbar({ message: 'The upload has been cancelled', autoHideDuration: TIMEOUTS.fiveSeconds }));
      }
      return commonErrorHandler(err, `Artifact couldn't be uploaded.`, dispatch);
    })
    .finally(() => dispatch(cleanUpUpload(uploadId)));
});

export const cancelFileUpload = createAsyncThunk(`${sliceName}/cancelFileUpload`, (id, { dispatch, getState }) => {
  const { [id]: current } = getState().app.uploadsById;
  current.cancelSource.abort();
  return Promise.resolve(dispatch(cleanUpUpload(id)));
});

export const editArtifact = createAsyncThunk(`${sliceName}/editArtifact`, ({ id, body }, { dispatch, getState }) =>
  GeneralApi.put(`${deploymentsApiUrl}/artifacts/${id}`, body)
    .catch(err => commonErrorHandler(err, `Artifact details couldn't be updated.`, dispatch))
    .then(() => {
      const state = getState();
      let { release, index } = findArtifactIndexInRelease(getReleasesById(state), id);
      if (!release || index === -1) {
        return dispatch(getReleases()).unwrap();
      }
      const updatedRelease = {
        ...release,
        artifacts: release.artifacts.map((artifact, i) => (i === index ? { ...artifact, description: body.description } : artifact))
      };
      return Promise.all([
        dispatch(actions.receiveRelease(updatedRelease)),
        dispatch(setSnackbar({ message: 'Artifact details were updated successfully.', autoHideDuration: TIMEOUTS.fiveSeconds, action: '' })),
        dispatch(getRelease(release.name)).unwrap(),
        dispatch(selectRelease(release.name)).unwrap()
      ]);
    })
);

export const removeArtifact = createAsyncThunk(`${sliceName}/removeArtifact`, (id, { dispatch, getState }) =>
  GeneralApi.delete(`${deploymentsApiUrl}/artifacts/${id}`)
    .then(() => {
      const state = getState();
      let { release, index } = findArtifactIndexInRelease(getReleasesById(state), id);
      const releaseArtifacts = [...release.artifacts];
      releaseArtifacts.splice(index, 1);
      if (!releaseArtifacts.length) {
        const { releasesList } = state.releases;
        const releaseIds = releasesList.releaseIds.filter(id => release.name !== id);
        return Promise.all([
          dispatch(actions.removeRelease(release.name)),
          dispatch(
            setReleasesListState({
              releaseIds,
              searchTotal: releasesList.searchTerm ? releasesList.searchTotal - 1 : releasesList.searchTotal,
              total: releasesList.total - 1
            })
          ).unwrap()
        ]);
      }
      return Promise.all([
        dispatch(setSnackbar({ message: 'Artifact was removed', autoHideDuration: TIMEOUTS.fiveSeconds, action: '' })),
        dispatch(actions.receiveRelease(release))
      ]);
    })
    .catch(err => commonErrorHandler(err, `Error removing artifact:`, dispatch))
);

export const removeRelease = createAsyncThunk(`${sliceName}/removeRelease`, (releaseId, { dispatch, getState }) =>
  Promise.all(getReleasesById(getState())[releaseId].artifacts.map(({ id }) => dispatch(removeArtifact(id)).unwrap())).then(() =>
    dispatch(selectRelease(null)).unwrap()
  )
);

export const removeReleases = createAsyncThunk(`${sliceName}/removeReleases`, (releaseIds, { dispatch, getState }) => {
  const deleteRequests = releaseIds.reduce((accu, releaseId) => {
    const releaseArtifacts = getReleasesById(getState())[releaseId].artifacts;
    accu.push(releaseArtifacts.map(({ id }) => dispatch(removeArtifact(id)).unwrap()));
    return accu;
  }, []);
  return Promise.all(deleteRequests);
});

export const selectRelease = createAsyncThunk(`${sliceName}/selectRelease`, (release, { dispatch }) => {
  const name = release ? release.name || release : null;
  let tasks = [dispatch(actions.selectedRelease(name))];
  if (name) {
    tasks.push(dispatch(getRelease(name)).unwrap());
  }
  return Promise.all(tasks);
});

export const setReleasesListState = createAsyncThunk(`${sliceName}/setReleasesListState`, (selectionState, { dispatch, getState }) => {
  const currentState = getState().releases.releasesList;
  let nextState = {
    ...currentState,
    ...selectionState,
    sort: { ...currentState.sort, ...selectionState.sort }
  };
  let tasks = [];
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { isLoading: currentLoading, ...currentRequestState } = currentState;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { isLoading: selectionLoading, ...selectionRequestState } = nextState;
  if (!deepCompare(currentRequestState, selectionRequestState)) {
    nextState.isLoading = true;
    tasks.push(
      dispatch(getReleases(nextState))
        .unwrap()
        .finally(() => dispatch(setReleasesListState({ isLoading: false })))
    );
  }
  tasks.push(dispatch(actions.setReleaseListState(nextState)));
  return Promise.all(tasks);
});

/* Releases */

const releaseListRetrieval = config => {
  const { searchTerm = '', page = defaultPage, perPage = defaultPerPage, sort = sortingDefaults, selectedTags = [], type = '' } = config;
  const { key: attribute, direction } = sort;
  const filterQuery = formatReleases({ pageState: { searchTerm, selectedTags } });
  const updateType = type ? `update_type=${type}` : '';
  const sorting = attribute ? `sort=${attribute}:${direction}`.toLowerCase() : '';
  return GeneralApi.get(
    `${deploymentsApiUrlV2}/deployments/releases?${[`page=${page}`, `per_page=${perPage}`, filterQuery, updateType, sorting].filter(i => i).join('&')}`
  );
};

const deductSearchState = (receivedReleases, config, total, state) => {
  let releaseListState = { ...state.releasesList };
  const { searchTerm, searchOnly, sort = {}, selectedTags = [], type } = config;
  const flattenedReleases = Object.values(receivedReleases).sort(customSort(sort.direction === SORTING_OPTIONS.desc, sort.key));
  const releaseIds = flattenedReleases.map(item => item.name);
  const isFiltering = !!(selectedTags.length || type || searchTerm);
  if (searchOnly) {
    releaseListState = { ...releaseListState, searchedIds: releaseIds };
  } else {
    releaseListState = {
      ...releaseListState,
      releaseIds,
      searchTotal: isFiltering ? total : state.releasesList.searchTotal,
      total: !isFiltering ? total : state.releasesList.total
    };
  }
  return releaseListState;
};

export const getReleases = createAsyncThunk(`${sliceName}/getReleases`, (passedConfig = {}, { dispatch, getState }) => {
  const config = { ...getState().releases.releasesList, ...passedConfig };
  return releaseListRetrieval(config)
    .then(({ data: receivedReleases = [], headers = {} }) => {
      const total = headers[headerNames.total] ? Number(headers[headerNames.total]) : 0;
      const state = getState().releases;
      const flatReleases = reduceReceivedReleases(receivedReleases, state.byId);
      const combinedReleases = { ...state.byId, ...flatReleases };
      let tasks = [dispatch(actions.receiveReleases(combinedReleases))];
      const releaseListState = deductSearchState(receivedReleases, config, total, state);
      tasks.push(dispatch(actions.setReleaseListState(releaseListState)));
      return Promise.all(tasks);
    })
    .catch(err => commonErrorHandler(err, `Please check your connection`, dispatch));
});

export const getRelease = createAsyncThunk(`${sliceName}/getReleases`, async (name, { dispatch, getState }) => {
  const releaseResponse = await GeneralApi.get(`${deploymentsApiUrlV2}/deployments/releases/${name}`);
  const { data: release } = releaseResponse;
  if (release) {
    const stateRelease = getReleasesById(getState())[release.name] || {};
    await dispatch(actions.receiveRelease(flattenRelease(release, stateRelease)));
  }
  return Promise.resolve(null);
});

export const updateReleaseInfo = createAsyncThunk(`${sliceName}/updateReleaseInfo`, ({ name, info }, { dispatch, getState }) =>
  GeneralApi.patch(`${deploymentsApiUrlV2}/deployments/releases/${name}`, info)
    .catch(err => commonErrorHandler(err, `Release details couldn't be updated.`, dispatch))
    .then(() =>
      Promise.all([
        dispatch(actions.receiveRelease({ ...getReleasesById(getState())[name], ...info, name })),
        dispatch(setSnackbar({ message: 'Release details were updated successfully.', autoHideDuration: TIMEOUTS.fiveSeconds, action: '' }))
      ])
    )
);

export const setSingleReleaseTags = createAsyncThunk(`${sliceName}/setSingleReleaseTags`, ({ name, tags }, { dispatch, getState }) =>
  GeneralApi.put(`${deploymentsApiUrlV2}/deployments/releases/${name}/tags`, tags).then(() =>
    Promise.resolve(dispatch(actions.receiveRelease({ ...getReleasesById(getState())[name], name, tags })))
  )
);

export const setReleaseTags = createAsyncThunk(`${sliceName}/setReleaseTags`, ({ name, tags = [] }, { dispatch }) =>
  dispatch(setSingleReleaseTags({ name, tags }))
    .unwrap()
    .catch(err => commonErrorHandler(err, `Release tags couldn't be set.`, dispatch))
    .then(() => Promise.resolve(dispatch(setSnackbar({ message: 'Release tags were set successfully.', autoHideDuration: TIMEOUTS.fiveSeconds, action: '' }))))
);

export const setReleasesTags = createAsyncThunk(`${sliceName}/setReleasesTags`, ({ releases, tags = [] }, { dispatch }) => {
  const addRequests = releases.reduce((accu, release) => {
    accu.push(dispatch(setSingleReleaseTags({ name: release.name, tags: [...new Set([...release.tags, ...tags])] })).unwrap());
    return accu;
  }, []);
  return Promise.all(addRequests)
    .catch(err => commonErrorHandler(err, `Releases couldn't be tagged.`, dispatch))
    .then(() => Promise.resolve(dispatch(setSnackbar({ message: 'Releases were tagged successfully.', autoHideDuration: TIMEOUTS.fiveSeconds, action: '' }))));
});

export const getExistingReleaseTags = createAsyncThunk(`${sliceName}/getReleaseTags`, (_, { dispatch }) =>
  GeneralApi.get(`${deploymentsApiUrlV2}/releases/all/tags`)
    .catch(err => commonErrorHandler(err, `Existing release tags couldn't be retrieved.`, dispatch))
    .then(({ data: tags }) => Promise.resolve(dispatch(actions.receiveReleaseTags(tags))))
);

export const getUpdateTypes = createAsyncThunk(`${sliceName}/getReleaseTypes`, (_, { dispatch }) =>
  GeneralApi.get(`${deploymentsApiUrlV2}/releases/all/types`)
    .catch(err => commonErrorHandler(err, `Existing update types couldn't be retrieved.`, dispatch))
    .then(({ data: types }) => Promise.resolve(dispatch(actions.receiveReleaseTypes(types))))
);
