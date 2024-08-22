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
import React from 'react';

import { defaultState, undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import SearchResult from './search-result';

describe('SearchResult Component', () => {
  it('renders correctly', async () => {
    const state = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasDeviceConfig: true,
          hasDeviceConnect: true,
          hasMultitenancy: true,
          isHosted: true
        },
        searchState: {
          ...defaultState.app.searchState,
          isSearching: true,
          searchTerm: 'something',
          sort: {}
        }
      }
    };
    const { baseElement } = render(<SearchResult onToggleSearchResult={jest.fn} open setSearchState={jest.fn} setSnackbar={jest.fn} />, {
      preloadedState: state
    });
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
