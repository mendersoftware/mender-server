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
import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Global from './Global';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasReporting: true,
      hasMultitenancy: true,
      isEnterprise: true,
      isHosted: true
    }
  },
  deployments: {
    ...defaultState.deployments,
    config: {
      ...defaultState.deployments.config,
      binaryDelta: {
        ...defaultState.deployments.config.binaryDelta,
        timeout: 5
      },
      hasDelta: true
    }
  }
};

describe('GlobalSettings Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Global />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
