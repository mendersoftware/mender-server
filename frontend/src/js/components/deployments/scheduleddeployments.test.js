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

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import ScheduledDeployments from './scheduleddeployments';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      isEnterprise: true,
      isHosted: false
    }
  },
  deployments: {
    ...defaultState.deployments,
    byStatus: {
      ...defaultState.deployments.byStatus,
      scheduled: { deploymentIds: [], total: 0 }
    },
    selectionState: {
      ...defaultState.deployments.selectionState,
      scheduled: {
        selection: []
      }
    }
  }
};

describe('ScheduledDeployments Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<ScheduledDeployments />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
