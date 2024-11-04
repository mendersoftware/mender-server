// Copyright 2021 Northern.tech AS
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

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import KeyValueEditor from './keyvalueeditor';

describe('KeyValueEditor Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <KeyValueEditor deviceLimitWarning={<div>I should not be rendered/ undefined</div>} limitMaxed={false} onSubmit={jest.fn} onCancel={jest.fn} />
    );
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  const fabSelector = '.MuiFab-root';
  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const submitMock = jest.fn();

    const ui = <KeyValueEditor onInputChange={submitMock} />;
    const { rerender } = render(ui);
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'testValue');
    expect(document.querySelector(fabSelector)).not.toBeDisabled();
    await user.click(document.querySelector(fabSelector));
    expect(submitMock).toHaveBeenLastCalledWith({ testKey: 'testValue' });
    await waitFor(() => rerender(ui));

    act(() => {
      jest.runAllTimers();
      jest.runAllTicks();
    });
    await user.type(screen.getByDisplayValue('testValue'), 's');
    expect(submitMock).toHaveBeenLastCalledWith({ testKey: 'testValues' });
  });

  it('warns of duplicate keys', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <KeyValueEditor onInputChange={jest.fn} />;
    const { rerender } = render(ui);
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'testValue');
    await user.click(document.querySelector(fabSelector));
    await waitFor(() => rerender(ui));
    await user.type(screen.getAllByPlaceholderText(/key/i)[1], 'testKey');
    await user.type(screen.getAllByPlaceholderText(/value/i)[1], 'testValue2');
    expect(screen.getByText(/Duplicate keys exist/i)).toBeInTheDocument();
  });

  it('forwards a warning', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <KeyValueEditor errortext="I should be rendered" onInputChange={jest.fn} />;
    render(ui);
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'testValue');
    expect(screen.getByText(/I should be rendered/i)).toBeInTheDocument();
  });

  it('displays tooltips when keys match', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const TestComponent = () => <div>testthing</div>;
    const helptipsMap = {
      timezone: {
        component: TestComponent
      }
    };

    const ui = <KeyValueEditor inputHelpTipsMap={helptipsMap} onInputChange={jest.fn} />;
    render(ui);
    await user.type(screen.getByPlaceholderText(/key/i), 'timezon');
    expect(screen.queryByText(/testthing/i)).not.toBeInTheDocument();
    await user.type(screen.getByPlaceholderText(/key/i), 'e');
    expect(screen.getByText(/testthing/i)).toBeInTheDocument();
  });
});
