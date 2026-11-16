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
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import FilterItem from './FilterItem';

describe('FilterItem Component', () => {
  const attributes = [
    { key: 'artifact_name', value: 'artifact_name', scope: 'inventory', category: 'inventory', priority: 1 },
    { key: 'test_device', value: 'test_device', scope: 'system', category: 'system', priority: 2 },
    { key: 'tier', value: 'tier', scope: 'system', category: 'system', priority: 2 },
    { key: 'created_ts', value: 'First request', scope: 'system', category: 'system', priority: 2 }
  ];

  const renderFilterItem = props =>
    render(
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <FilterItem attributes={attributes} enabledTiers={[]} onChange={vi.fn()} onSave={vi.fn()} onSelect={vi.fn()} {...props} />
      </LocalizationProvider>
    );

  const selectAttribute = async (user, name) => {
    await user.click(screen.getByRole('combobox', { name: 'Attribute' }));
    await user.click(screen.getByRole('option', { name }));
    await act(async () => vi.runOnlyPendingTimers());
  };

  it('renders correctly', async () => {
    const { baseElement } = renderFilterItem({ attributes: [{ key: 'testkey', value: 'testvalue' }] });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('implies the value of the test device attribute without showing an input', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onChange = vi.fn();
    renderFilterItem({ onChange });
    await selectAttribute(user, /test_device/);
    expect(onChange).toHaveBeenLastCalledWith({ key: 'test_device', operator: '$eq', scope: 'system', value: 'true' });
    expect(screen.queryByRole('textbox', { name: 'Value' })).not.toBeInTheDocument();
    await user.click(screen.getByRole('combobox', { name: '' }));
    expect(screen.getAllByRole('option').map(option => option.textContent)).toEqual(['equals', 'not equal']);
  });
  it('lets the tier be picked from the enabled tiers and drops it when switching to a regular attribute', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onChange = vi.fn();
    renderFilterItem({ enabledTiers: ['standard', 'micro'], onChange });
    await selectAttribute(user, /tier/);
    await user.click(screen.getByRole('combobox', { name: 'Value' }));
    await user.click(screen.getByRole('option', { name: 'micro' }));
    await act(async () => vi.runOnlyPendingTimers());
    expect(onChange).toHaveBeenLastCalledWith({ key: 'tier', operator: '$eq', scope: 'system', value: 'micro' });
    await user.click(screen.getByRole('combobox', { name: '' }));
    await user.click(screen.getByRole('option', { name: 'not equal' }));
    await act(async () => vi.runOnlyPendingTimers());
    expect(onChange).toHaveBeenLastCalledWith({ key: 'tier', operator: '$ne', scope: 'system', value: 'micro' });
    await user.tripleClick(screen.getByRole('combobox', { name: 'Attribute' }));
    await user.keyboard('artifact');
    await user.click(screen.getByRole('option', { name: /Current artifact/ }));
    await act(async () => vi.runOnlyPendingTimers());
    expect(onChange).toHaveBeenLastCalledWith({ key: 'artifact_name', operator: '$ne', scope: 'inventory', value: '' });
  });
  it('accepts dates at month, day or time precision for date attributes', async () => {
    vi.setSystemTime(new Date('2026-11-13T10:00:00'));
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onChange = vi.fn();
    renderFilterItem({ onChange });
    await selectAttribute(user, /First request/);
    expect(onChange).toHaveBeenLastCalledWith({ key: 'created_ts', operator: '$eq', scope: 'system', value: '' });
    const input = screen.getByRole('textbox', { name: 'Value' });
    await user.type(input, '2026-11-1');
    await act(async () => vi.runOnlyPendingTimers());
    // the last complete date entered was the month
    expect(onChange.mock.lastCall[0].value).toBe('2026-11');
    await user.type(input, '2 13:45');
    await act(async () => vi.runOnlyPendingTimers());
    expect(onChange).toHaveBeenLastCalledWith({ key: 'created_ts', operator: '$eq', scope: 'system', value: '2026-11-12 13:45' });

    await user.clear(input);
    await user.click(screen.getByRole('button', { name: 'Pick a date' }));
    await user.click(screen.getByRole('button', { name: /calendar view is open, switch to year view/i }));
    await user.click(screen.getByRole('radio', { name: '2026' }));
    // the current month has to be selectable on its own as well
    await user.click(screen.getByRole('radio', { name: 'November' }));
    await act(async () => vi.runOnlyPendingTimers());
    expect(input).toHaveValue('2026-11');
    expect(onChange.mock.lastCall[0].value).toBe('2026-11');
    expect(screen.queryByRole('grid')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Pick a date' }));
    await user.click(screen.getByRole('gridcell', { name: '5' }));
    await act(async () => vi.runOnlyPendingTimers());
    expect(input).toHaveValue('2026-11-05');
    expect(onChange.mock.lastCall[0].value).toBe('2026-11-05');
  });
});
