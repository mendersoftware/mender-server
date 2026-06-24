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
import { formRenderWrapper } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import ChipSelect from './ChipSelect';

const defaultProps = {
  name: 'tags',
  label: 'Tags',
  options: ['alpha', 'beta', 'gamma'],
  placeholder: 'Select tags'
};

const formConfig = { mode: 'onChange', defaultValues: { tags: [] } };

describe('ChipSelect Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(<ChipSelect {...defaultProps} />, formConfig);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders with disabled state as readOnly', async () => {
    formRenderWrapper(<ChipSelect {...defaultProps} disabled />, formConfig);
    const input = screen.getByRole('combobox');
    expect(input).toHaveAttribute('readonly');
  });

  it('allows selecting an option from the dropdown', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    formRenderWrapper(<ChipSelect {...defaultProps} />, formConfig);
    const input = screen.getByRole('combobox');
    await user.click(input);
    await user.click(screen.getByText('alpha'));
    expect(screen.getByText('alpha')).toBeInTheDocument();
  });

  it('adds comma-separated values as chips', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    formRenderWrapper(<ChipSelect {...defaultProps} />, formConfig);
    const input = screen.getByRole('combobox');
    await user.type(input, 'foo,bar,');
    await waitFor(() => expect(screen.getByText('foo')).toBeInTheDocument());
    expect(screen.getByText('bar')).toBeInTheDocument();
  });

  it('adds whitespace-separated values as chips on blur', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    formRenderWrapper(<ChipSelect {...defaultProps} />, formConfig);
    const input = screen.getByRole('combobox');
    await user.type(input, 'one two');
    await user.tab();
    await waitFor(() => expect(screen.getByText('one')).toBeInTheDocument());
    expect(screen.getByText('two')).toBeInTheDocument();
  });
});
