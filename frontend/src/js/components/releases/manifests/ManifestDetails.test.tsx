// Copyright 2026 Northern.tech AS
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
import { paperClasses } from '@mui/material';

import { defaultState, render } from '@/testUtils';
import { ColumnWidthProvider } from '@northern.tech/common-ui/TwoColumnData';
import { undefineds } from '@northern.tech/testing/mockData';

import { ManifestDetails } from './ManifestDetails';

const preloadedState = {
  ...defaultState,
  releases: {
    ...defaultState.releases,
    selectedManifest: 'm1000'
  }
};

describe('ManifestDetails Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <ColumnWidthProvider>
        <ManifestDetails />
      </ColumnWidthProvider>,
      { preloadedState }
    );
    const view = baseElement.querySelector(`.${paperClasses.root}`);
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('does not render when no manifest is selected', () => {
    const { baseElement } = render(
      <ColumnWidthProvider>
        <ManifestDetails />
      </ColumnWidthProvider>
    );
    const view = baseElement.querySelector(`.${paperClasses.root}`);
    expect(view).toBeFalsy();
  });
});
