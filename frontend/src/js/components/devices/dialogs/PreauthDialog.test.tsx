// Copyright 2020 Northern.tech AS
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
import { Provider } from 'react-redux';

import { defaultState, render, server } from '@/testUtils';
import { TIMEOUTS, deviceAuthV2 } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { mockDate } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import configureStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';
import { vi } from 'vitest';

import { PreauthDialog } from './PreauthDialog';

const mockStore = configureStore([thunk]);

const dropzone = '.dropzone input';

const publicKey = `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAGXbdqZ5RyULfSYWUETrpecTJ2AWVN6i+qEsgt0CsiO4=
-----END PUBLIC KEY-----`;

const keyFile = () => new File([publicKey], 'test.pem');

let store;

describe('PreauthDialog Component', () => {
  beforeEach(() => {
    // jsdom 29 uses setImmediate for FileReader events; fake it so file uploads work with fake timers
    vi.useFakeTimers({
      now: mockDate,
      toFake: [
        'setTimeout',
        'clearTimeout',
        'setInterval',
        'clearInterval',
        'setImmediate',
        'clearImmediate',
        'Date',
        'requestAnimationFrame',
        'cancelAnimationFrame'
      ]
    });
    store = mockStore({ ...defaultState });
  });

  it('renders correctly', async () => {
    const { baseElement } = render(
      <Provider store={store}>
        <PreauthDialog deviceLimitWarning={<div>I should not be rendered/ undefined</div>} limitMaxed={false} onSubmit={vi.fn} onCancel={vi.fn} />
      </Provider>
    );
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', { timeout: TIMEOUTS.fiveSeconds + TIMEOUTS.oneSecond }, async () => {
    const { preauthDevice: preAuthSpy } = StoreThunks;

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime, applyAccept: false });
    const submitMock = vi.fn();
    const menderFile = keyFile();
    const ui = (
      <Provider store={store}>
        <PreauthDialog limitMaxed={false} onSubmit={submitMock} onCancel={vi.fn()} />
      </Provider>
    );
    const { rerender } = render(ui);
    expect(screen.getByText(/to upload a file/i)).toBeInTheDocument();
    // container.querySelector doesn't work in this scenario for some reason -> but querying document seems to work
    const uploadInput = document.querySelector(dropzone);
    await user.upload(uploadInput, menderFile);
    await act(async () => vi.runOnlyPendingTimers());
    await waitFor(() => rerender(ui));

    expect(uploadInput.files).toHaveLength(1);
    await waitFor(() => expect(document.querySelector(dropzone)).not.toBeInTheDocument());
    expect(screen.getByText('test.pem')).toBeInTheDocument();
    expect(screen.getByTitle(/accepted/i)).toBeInTheDocument();
    const addButton = screen.getByLabelText('add-editor-line-button');
    expect(addButton).toBeDisabled();
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'testValue');
    expect(addButton).not.toBeDisabled();
    await user.click(addButton);
    await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument());
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    server.use(http.post(`${deviceAuthV2}/devices`, () => HttpResponse.json({ error: 'test conflict' }, { status: 409 }), { once: true }));
    await user.click(screen.getByRole('button', { name: 'Save' }));
    act(() => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('A device with a matching identity data set already exists'));
    await user.type(screen.getByDisplayValue('testValue'), 'testValues');
    await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: 'Save and add another' }));
    await waitFor(() => rerender(ui));
    expect(screen.queryByText('reached your limit')).toBeFalsy();
    expect(preAuthSpy).toHaveBeenCalled();
  });

  it('rejects files that do not contain a public key', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime, applyAccept: false });
    const ui = (
      <Provider store={store}>
        <PreauthDialog limitMaxed={false} onSubmit={vi.fn()} onCancel={vi.fn()} />
      </Provider>
    );
    const { rerender } = render(ui);
    await user.upload(document.querySelector(dropzone), new File(['definitely not a key'], 'test.txt'));
    await act(async () => vi.runOnlyPendingTimers());
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent(/does not contain a public key/i));
    expect(screen.queryByTitle(/accepted/i)).not.toBeInTheDocument();
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'testValue');
    await waitFor(() => rerender(ui));
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    // removing the rejected file has to make the dialog usable again
    await user.click(screen.getByRole('button', { name: /remove the selected file/i }));
    await waitFor(() => rerender(ui));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    await user.upload(document.querySelector(dropzone), keyFile());
    await act(async () => vi.runOnlyPendingTimers());
    await waitFor(() => rerender(ui));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(screen.getByTitle(/accepted/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).not.toBeDisabled();
  });

  it('prevents preauthorizations when device limit was reached', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const menderFile = keyFile();
    const ui = (
      <Provider store={store}>
        <PreauthDialog acceptedDevices={100} deviceLimit={2} limitMaxed={true} />
      </Provider>
    );

    const { rerender } = render(ui);
    // container.querySelector doesn't work in this scenario for some reason -> but querying document seems to work
    const uploadInput = document.querySelector(dropzone);
    await user.upload(uploadInput, menderFile);
    await waitFor(() => rerender(ui));
    await user.type(screen.getByPlaceholderText(/key/i), 'testKey');
    await user.type(screen.getByPlaceholderText(/value/i), 'testValue');
    await waitFor(() => rerender(ui));
    expect(screen.getByText(/You have reached your limit/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
  });
});
