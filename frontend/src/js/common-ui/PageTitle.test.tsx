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
import { render } from '@/testUtils';
import { waitFor } from '@testing-library/react';

import PageTitle, { getPageTitle } from './PageTitle';

describe('PageTitle Component', () => {
  it('joins segments from most to least specific', () => {
    expect(getPageTitle(['Finished', 'Deployments'])).toEqual('Finished | Deployments | Mender');
    expect(getPageTitle([undefined, 'Devices', ''])).toEqual('Devices | Mender');
    expect(getPageTitle([])).toEqual('Mender');
  });

  it('sets the document title', async () => {
    render(<PageTitle segments={['Accepted', 'Devices']} />);
    await waitFor(() => expect(document.title).toEqual('Accepted | Devices | Mender'));
  });
});
