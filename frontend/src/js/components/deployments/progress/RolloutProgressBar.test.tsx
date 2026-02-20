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

import { RolloutProgressBar } from './RolloutProgressBar';

describe('RolloutProgressBar Component', () => {
  it('renders correctly with list variant', async () => {
    const { baseElement } = render(<RolloutProgressBar deployment={defaultState.deployments.byId.d2} variant="list" />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly with report variant', async () => {
    const { baseElement } = render(<RolloutProgressBar deployment={defaultState.deployments.byId.d2} variant="report" />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly for phases with list variant', async () => {
    const { baseElement } = render(<RolloutProgressBar deployment={defaultState.deployments.byId.d3} variant="list" />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
