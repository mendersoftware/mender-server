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
// @ts-nocheck
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS } from '@northern.tech/store/constants';
import { createSlice } from '@reduxjs/toolkit';

export const sliceName = 'releases';

export const initialState = {
  /*
   * Return list of saved artifacts objects
   */
  /*
   * return list of artifacts where duplicate names are collated with device compatibility lists combined
   */
  // artifacts: AppStore.getCollatedArtifacts(AppStore.getArtifactsRepo()),
  artifacts: [],
  /*
   * Return list of saved release objects
   */
  byId: {
    /*
    [releaseName]: {
      artifacts: [
        {
          id: '',
          name: '',
          description: '',
          device_types_compatible: [],
          ...
          updates: [{
            files: [
              { size: 123, name: '' }
            ],
            type_info: { type: '' }
          }],
          url: '' // optional
        }
      ],
      modified: ''
      device_types_compatible,
      name: '',
      tags: ['something'],
      notes: ''
    }
    */
  },
  releasesList: {
    ...DEVICE_LIST_DEFAULTS,
    searchedIds: [],
    releaseIds: [],
    selection: [],
    sort: {
      direction: SORTING_OPTIONS.desc,
      key: 'modified'
    },
    isLoading: undefined,
    searchTerm: '',
    searchTotal: 0,
    selectedTags: [],
    total: 0,
    type: ''
  },
  tags: [],
  updateTypes: [],
  /*
   * Return single release with corresponding Artifacts
   */
  selectedRelease: null
};

export const releaseSlice = createSlice({
  name: sliceName,
  initialState,
  reducers: {
    receiveRelease: (state, action) => {
      const { name } = action.payload;
      state.byId[name] = action.payload;
    },
    receiveReleases: (state, action) => {
      state.byId = action.payload;
    },
    receiveReleaseTags: (state, action) => {
      state.tags = action.payload;
    },
    receiveReleaseTypes: (state, action) => {
      state.updateTypes = action.payload;
    },
    removeRelease: (state, action) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [action.payload]: toBeRemoved, ...byId } = state.byId;
      state.byId = byId;
      state.selectedRelease = action.payload === state.selectedRelease ? null : state.selectedRelease;
    },
    selectedRelease: (state, action) => {
      state.selectedRelease = action.payload;
    },
    setReleaseListState: (state, action) => {
      state.releasesList = action.payload;
    }
  }
});

export const actions = releaseSlice.actions;
export default releaseSlice.reducer;
