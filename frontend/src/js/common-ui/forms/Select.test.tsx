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
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { Select } from './Select';

const options = [
  { id: 'a', title: 'Alpha' },
  { id: 'b', title: 'Beta' },
  { id: 'c', title: 'Gamma' }
];

describe('Select Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Select label="Letter" helperText="Pick one" options={options} value="a" onChange={vi.fn()} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('calls onChange with the picked option value', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onChange = vi.fn();
    render(<Select options={options} value="" onChange={onChange} />);
    await user.click(screen.getByRole('combobox'));
    await user.click(screen.getByRole('option', { name: 'Beta' }));
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange.mock.calls[0][0].target.value).toBe('b');
  });

  it('shows the placeholder while nothing is selected and offers it as an empty option', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<Select options={options} placeholder="Select a letter" value="" onChange={vi.fn()} />);
    expect(screen.getByRole('combobox')).toHaveTextContent('Select a letter');
    await user.click(screen.getByRole('combobox'));
    expect(screen.getByRole('option', { name: 'Select a letter' })).toBeInTheDocument();
  });

  it('renders checkboxes and joins the selected labels in multiple mode', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<Select multiple options={options} value={['a', 'b']} onChange={vi.fn()} />);
    expect(screen.getByRole('combobox')).toHaveTextContent('Alpha, Beta');
    await user.click(screen.getByRole('combobox'));
    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes).toHaveLength(options.length);
    expect(checkboxes[0]).toBeChecked();
    expect(checkboxes[2]).not.toBeChecked();
  });
});
