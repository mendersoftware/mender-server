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
import { formRenderWrapper, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { CheckboxAutocomplete, ControlledCheckboxAutocomplete } from './CheckboxAutocomplete';

const options = [
  { title: 'Alpha', id: 'alpha' },
  { title: 'Beta', id: 'beta' },
  { title: 'Gamma', id: 'gamma' },
  { title: 'Delta', id: 'delta' }
];

const defaultProps = {
  label: 'Select items',
  options,
  placeholder: 'Choose...'
};

const formConfig = { mode: 'onChange', defaultValues: { items: [] } };

describe('CheckboxAutocomplete Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<CheckboxAutocomplete {...defaultProps} onChange={vi.fn()} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('shows checkboxes in dropdown options', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<CheckboxAutocomplete {...defaultProps} onChange={vi.fn()} />);
    await user.click(screen.getByRole('combobox'));
    const listbox = screen.getByRole('listbox');
    const checkboxes = within(listbox).getAllByRole('checkbox');
    expect(checkboxes).toHaveLength(options.length);
    expect(checkboxes[0]).not.toBeChecked();
  });

  it('keeps dropdown open after selecting an option', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<CheckboxAutocomplete {...defaultProps} onChange={vi.fn()} />);
    await user.click(screen.getByRole('combobox'));
    await user.click(screen.getByText('Alpha'));
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });

  it('renders chips when chipDisplay is true', async () => {
    render(<CheckboxAutocomplete {...defaultProps} onChange={vi.fn()} value={[options[0], options[1]]} chipDisplay />);
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });
});

describe('ControlledCheckboxAutocomplete Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(<ControlledCheckboxAutocomplete {...defaultProps} name="items" />, formConfig);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
