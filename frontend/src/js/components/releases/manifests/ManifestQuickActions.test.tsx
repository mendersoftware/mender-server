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
import { defaultState, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';

import ManifestQuickActions from './ManifestQuickActions';

describe('ManifestQuickActions Component', () => {
  it('renders correctly', async () => {
    const preloadedState = {
      ...defaultState,
      releases: {
        ...defaultState.releases,
        manifestsList: {
          ...defaultState.releases.manifestsList,
          selection: [0]
        }
      }
    };
    const { baseElement } = render(<ManifestQuickActions />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders the speed dial with actions when manifests are selected', async () => {
    const preloadedState = {
      ...defaultState,
      releases: {
        ...defaultState.releases,
        manifestsList: {
          ...defaultState.releases.manifestsList,
          selection: [0]
        }
      }
    };
    render(<ManifestQuickActions />, { preloadedState });
    expect(screen.getByLabelText('manifest-actions')).toBeInTheDocument();
    expect(screen.getByText(/1 Manifest selected/i)).toBeInTheDocument();
  });
});
