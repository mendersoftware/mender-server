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
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { InventoryTable } from './InventoryTable';

const testConfig = {
  hostname: 'device-001',
  ipAddress: '192.168.1.100',
  macAddress: 'aa:bb:cc:dd:ee:ff'
};

describe('InventoryTable Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<InventoryTable config={testConfig} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('toggles sort direction when clicking the same column header', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<InventoryTable config={testConfig} />);
    await user.click(screen.getByText('Attribute'));

    const rows = screen.getAllByRole('row');
    expect(within(rows[1]).getByText('macAddress')).toBeInTheDocument();
    expect(within(rows[3]).getByText('hostname')).toBeInTheDocument();
  });
});
