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
import { render } from '@/testUtils';
import { PLANS } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { vi } from 'vitest';

import { SubscriptionSummary } from './SubscriptionSummary';

describe('Subscription Summary component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <SubscriptionSummary
        addons={['configure']}
        deviceLimit={50}
        isEnabled
        isPreviewLoading={false}
        onAction={vi.fn}
        plan={PLANS.os}
        previewPrice={{ addons: { configure: 1000 }, standard: { quantity: 50, price: 3200 }, micro: { quantity: 100, price: 3200 }, total: 7400 }}
        readOnly={false}
        title={'your subscription '}
      />
    );
    const view = baseElement.lastElementChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
