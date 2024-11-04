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
import React from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Form from './form';
import FormCheckbox from './formcheckbox';
import PasswordInput from './passwordinput';
import TextInput from './textinput';

export const formRenderWrapper = ui => {
  const Wrapper = ({ children }) => {
    const methods = useForm();
    return <FormProvider {...methods}>{children}</FormProvider>;
  };
  return render(<Wrapper>{ui}</Wrapper>);
};

describe('Form Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <Form showButtons>
        <FormCheckbox id="testbox" label="testbox" />
        <PasswordInput id="password" create />
        <TextInput id="textbox" />
      </Form>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('works correctly with generated passwords', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    const ui = (
      <Form showButtons submitLabel="submit">
        <PasswordInput id="password" required create generate />
      </Form>
    );
    const { rerender } = render(ui);
    await user.click(screen.getByRole('button', { name: /generate/i }));
    await waitFor(() => rerender(ui));
    expect(screen.getByRole('button', { name: /submit/i })).not.toBeDisabled();
  });
});
