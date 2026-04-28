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
import { useState } from 'react';

import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { NumberField } from './NumberField';

describe('NumberField Component', () => {
  it('renders correctly', () => {
    const { baseElement } = render(<NumberField id="quantity" label="Quantity" />);
    const view = baseElement.firstChild?.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('does not render the stepper by default', () => {
    render(<NumberField id="quantity" label="Quantity" />);
    expect(screen.queryByRole('button', { name: /increase/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /decrease/i })).not.toBeInTheDocument();
  });

  it('renders the stepper when showSteps is true and increments via click', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const Harness = () => {
      const [value, setValue] = useState<number | null>(5);
      return <NumberField id="quantity" label="Quantity" value={value} onValueChange={setValue} showSteps step={10} />;
    };
    render(<Harness />);
    const increase = screen.getByRole('button', { name: /increase/i });
    const decrease = screen.getByRole('button', { name: /decrease/i });
    expect(increase).toBeInTheDocument();
    expect(decrease).toBeInTheDocument();
    await user.click(increase);
    await waitFor(() => expect(screen.getByRole('textbox', { name: /quantity/i })).toHaveValue('15'));
    await user.click(decrease);
    await user.click(decrease);
    await waitFor(() => expect(screen.getByRole('textbox', { name: /quantity/i })).toHaveValue('-5'));
  });

  it('marks the input as required when the required prop is set', () => {
    render(<NumberField id="quantity" label="Quantity" required />);
    expect(screen.getByRole('textbox', { name: /quantity/i })).toBeRequired();
  });

  it('shows helper text and reflects the error state', () => {
    render(<NumberField id="quantity" label="Quantity" error helperText="Out of range" />);
    expect(screen.getByText('Out of range')).toBeInTheDocument();
  });

  it('disables stepper buttons and input when disabled', () => {
    render(<NumberField id="quantity" label="Quantity" showSteps disabled />);
    expect(screen.getByRole('button', { name: /increase/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /decrease/i })).toBeDisabled();
    expect(screen.getByRole('textbox', { name: /quantity/i })).toBeDisabled();
  });
});
