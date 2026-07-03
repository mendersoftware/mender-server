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
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { AsyncAutocomplete, ControlledAsyncAutocomplete } from './AsyncAutocomplete';

const options = [
  { title: 'Device A', id: 'device-a' },
  { title: 'Device B', id: 'device-b' }
];

const defaultProps = {
  label: 'Search devices',
  onSearch: vi.fn(),
  options,
  placeholder: 'Type to search...'
};

const formConfig = { mode: 'onChange', defaultValues: { search: null } };

describe('AsyncAutocomplete Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(<AsyncAutocomplete {...defaultProps} onChange={vi.fn()} />, formConfig);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('calls onSearch with debounced input', async () => {
    const onSearch = vi.fn();
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<AsyncAutocomplete {...defaultProps} onChange={vi.fn()} onSearch={onSearch} />);
    const input = screen.getByRole('combobox');
    await user.type(input, 'dev');
    await vi.advanceTimersByTimeAsync(500);
    await waitFor(() => expect(onSearch).toHaveBeenCalledWith('dev'));
  });

  it('shows loading indicator when isLoading is true', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<AsyncAutocomplete {...defaultProps} onChange={vi.fn()} isLoading />);
    const input = screen.getByRole('combobox');
    await user.click(input);
    expect(document.querySelector('.miniLoaderContainer')).toBeInTheDocument();
  });
});

describe('ControlledAsyncAutocomplete Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(<ControlledAsyncAutocomplete {...defaultProps} name="search" />, formConfig);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
