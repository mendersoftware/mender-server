// Copyright 2025 Northern.tech AS
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

import { ListProgress, SimpleProgress } from './RolloutProgressBar';

describe('ListProgress Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<ListProgress deployment={defaultState.deployments.byId.d2} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly for phases', async () => {
    const { baseElement } = render(<ListProgress deployment={{ ...defaultState.deployments.byId.d3, current_phase: 2 }} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    expect(screen.getByText('Current phase: 2 of 5')).toBeInTheDocument();
  });
});

describe('SimpleProgress Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<SimpleProgress deployment={defaultState.deployments.byId.d3} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
