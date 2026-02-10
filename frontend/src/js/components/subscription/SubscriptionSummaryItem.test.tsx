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

import { SubscriptionSummaryItem } from './SubscriptionSummaryItem';

describe('Subscription Summary item component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <SubscriptionSummaryItem
        addons={['configure', 'monitor', 'troubleshoot']}
        isPreviewLoading={false}
        previewPriceItem={{
          addons: { configure: 1000, monitor: 1000, troubleshoot: 2000 },
          price: 2000,
          quantity: 50
        }}
        summaryLabel="Micro Devices"
      />,
      {
        preloadedState: {
          ...defaultState,
          app: {
            ...defaultState.app,
            features: {
              ...defaultState.app.features,
              isHosted: true
            }
          }
        }
      }
    );
    const view = baseElement.lastElementChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
