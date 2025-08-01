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
type Line = { amount: number; currency: string; description: string; price_id: string; quantity: number };

const MENDER_STRIPE_PREFIX = 'Mender';

export const parseSubscriptionPreview = (lines: Line[], plan: string) => {
  const idMap = {};
  lines.forEach(line => {
    if (!idMap[line.price_id]) {
      idMap[line.price_id] = { id: line.price_id, total: 0 };
    } else {
      if (line.description.startsWith(MENDER_STRIPE_PREFIX)) {
        const title = line.description.split(' ')[1];
        idMap[line.price_id] = { ...idMap[line.price_id], title, isPlan: title.toLowerCase() === plan };
      }
    }
    idMap[line.price_id].total += line.amount;
  });

  const items = Object.values(idMap);
  const planPrice = items.find(p => p.isPlan)?.total || 0;
  const addons = items.filter(p => !p.isPlan).reduce((acc, curr) => ({ ...acc, [curr.title.toLowerCase()]: curr.total }), {});
  return { plan: planPrice, addons };
};
export const currencyFormatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 });
export const formatPrice = (cents: number) => currencyFormatter.format(cents / 100);
