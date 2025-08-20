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
import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import Form from './Form';
import FormCheckbox from './FormCheckbox';
import PasswordInput from './PasswordInput';
import TextInput from './TextInput';

describe('Form Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <Form showButtons submitLabel="submit">
        <FormCheckbox id="testbox" label="testbox" />
        <PasswordInput id="password" create />
        <TextInput id="textbox" />
      </Form>
    );
    expect(await screen.findByText('submit')).toBeInTheDocument();
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  window.prompt = vi.fn();
  it('works correctly with generated passwords', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });

    const ui = (
      <Form showButtons submitLabel="submit">
        <PasswordInput id="password" required create generate />
      </Form>
    );
    const { rerender } = render(ui);
    await user.click(screen.getByRole('button', { name: /generate/i }));
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByRole('button', { name: /submit/i })).not.toBeDisabled());
    window.prompt.mockClear();
  });
});
