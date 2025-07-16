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
import { parseSubscriptionPreview } from './utils';

const stripePreview = {
  id: 'upcoming_in_1Rabc',
  period_start: '2025-07-28T10:39:43Z',
  period_end: '2025-07-28T10:39:43Z',
  total: 3200,
  currency: 'usd',
  lines: [
    {
      description: '50 device × Mender Basic (Tier 1 at $0.00 / month)',
      amount: 0,
      currency: 'usd',
      quantity: 50,
      price_id: 'price_1abc'
    },
    {
      description: 'Mender Basic (Tier 1 at $32.00 / month)',
      amount: 3200,
      currency: 'usd',
      quantity: 0,
      price_id: 'price_1abc'
    },
    {
      description: '50 device × Mender Configure (Tier 1 at $0.00 / month)',
      amount: 0,
      currency: 'usd',
      quantity: 50,
      price_id: 'price_1PYl8wFlFfXikjZVIlqZ6zYr'
    },
    {
      description: 'Mender Configure (Tier 1 at $10.00 / month)',
      amount: 1000,
      currency: 'usd',
      quantity: 0,
      price_id: 'price_1PYl8wFlFfXikjZVIlqZ6zYr'
    },
    {
      description: '50 device × Mender Troubleshoot (Tier 1 at $0.00 / month)',
      amount: 0,
      currency: 'usd',
      quantity: 50,
      price_id: 'price_1PYkxZFlFfXikjZVJYRaUPld'
    },
    {
      description: 'Mender Troubleshoot (Tier 1 at $20.00 / month)',
      amount: 2000,
      currency: 'usd',
      quantity: 0,
      price_id: 'price_1PYkxZFlFfXikjZVJYRaUPld'
    }
  ]
};
describe('subscription utils', () => {
  it('stripe subscription preview', async () => {
    const result = parseSubscriptionPreview(stripePreview.lines, 'basic');
    expect(result.plan).toEqual(3200);
    expect(result.addons.configure).toEqual(1000);
    expect(result.addons.troubleshoot).toEqual(2000);
  });
});
