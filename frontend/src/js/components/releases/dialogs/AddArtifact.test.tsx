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
import { defaultState, render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import AddArtifact from './AddArtifact';

const destinationPlaceholder = 'Example: /opt/installed-by-single-file';

describe('AddArtifact Component', () => {
  const preloadedState = { ...defaultState, onboarding: { ...defaultState.onboarding, complete: true } };
  it('renders correctly', async () => {
    const { baseElement } = render(<AddArtifact />, { preloadedState });
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows uploading a mender artifact', async () => {
    const { uploadArtifact: uploadSpy } = StoreThunks;
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const menderFile = new File(['testContent'], 'test.mender');

    const ui = <AddArtifact onUploadStarted={vi.fn} />;
    const { rerender } = render(ui, { preloadedState });
    expect(screen.getByText(/Upload a premade/i)).toBeInTheDocument();
    // container.querySelector doesn't work in this scenario for some reason -> but querying document seems to work
    const uploadInput = document.querySelector('.dropzone input');
    await user.upload(uploadInput, menderFile);
    expect(uploadInput.files).toHaveLength(1);
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByRole('button', { name: /upload/i })).toBeInTheDocument());
    expect(screen.getByText('test.mender')).toBeInTheDocument();
    // FileSize component is not an input based component -> query text only
    expect(screen.getByText('11.00 Bytes')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /upload/i }));
    expect(uploadSpy).toHaveBeenCalledWith({ file: menderFile, meta: { description: '' } });
  });

  it('validates the artifact information only when trying to progress', { timeout: TIMEOUTS.refreshDefault }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const singleFile = new File(['testContent plain'], 'testFile.txt');
    render(<AddArtifact onUploadStarted={vi.fn} />, { preloadedState });
    await user.upload(document.querySelector('.dropzone input'), singleFile);
    await waitFor(() => expect(screen.getByPlaceholderText(destinationPlaceholder)).toBeInTheDocument());
    expect(screen.queryByText(/is required/i)).not.toBeInTheDocument();

    const nextButton = screen.getByRole('button', { name: /next/i });
    expect(nextButton).toBeEnabled();
    await user.click(nextButton);
    await waitFor(() => expect(screen.getByText(/Destination directory is required/i)).toBeInTheDocument());
    expect(screen.getByText(/Device type is required/i)).toBeInTheDocument();
    // the release name is prefilled from the selected file, so it doesn't need to be entered again
    expect(screen.queryByText(/Release name is required/i)).not.toBeInTheDocument();
    // the dialog stays on the first step until the entered information is valid
    expect(screen.queryByText(/Version information/i)).not.toBeInTheDocument();

    await user.type(screen.getByPlaceholderText(destinationPlaceholder), 'some/path');
    await waitFor(() => expect(screen.getByText(/Destination has to be an absolute path/i)).toBeInTheDocument());
    await user.clear(screen.getByPlaceholderText(destinationPlaceholder));
    await user.type(screen.getByPlaceholderText(destinationPlaceholder), '/some/path');
    await waitFor(() => expect(screen.queryByText(/Destination has to be an absolute path/i)).not.toBeInTheDocument());
  });

  it('keeps a modified software name when going back to change the release name', { timeout: TIMEOUTS.refreshDefault }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const singleFile = new File(['testContent plain'], 'testFile.txt');
    render(<AddArtifact onUploadStarted={vi.fn} />, { preloadedState });
    await user.upload(document.querySelector('.dropzone input'), singleFile);
    await waitFor(() => expect(screen.getByPlaceholderText(destinationPlaceholder)).toBeInTheDocument());
    await user.type(screen.getByPlaceholderText(destinationPlaceholder), '/some/path');
    await user.type(screen.getByRole('combobox', { name: /device types compatible/i }), 'something');
    await user.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() => expect(screen.getByText(/Version information/i)).toBeInTheDocument());

    await user.clear(screen.getByLabelText(/software name/i));
    await user.type(screen.getByLabelText(/software name/i), 'custom software name');
    await user.click(screen.getByRole('button', { name: /back/i }));
    await waitFor(() => expect(screen.getByLabelText(/release name/i)).toBeInTheDocument());
    await user.clear(screen.getByLabelText(/release name/i));
    await user.type(screen.getByLabelText(/release name/i), 'some release');
    await user.click(screen.getByRole('button', { name: /next/i }));

    await waitFor(() => expect(screen.getByText(/Version information/i)).toBeInTheDocument());
    expect(screen.getByLabelText(/software name/i)).toHaveValue('custom software name');
  });

  it('allows creating a mender artifact', { timeout: TIMEOUTS.refreshDefault }, async () => {
    const { createArtifact: creationSpy } = StoreThunks;

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });

    const singleFile = new File(['testContent plain'], 'testFile.txt');
    const ui = <AddArtifact onUploadStarted={vi.fn} />;
    const { rerender } = render(ui, { preloadedState });
    expect(screen.getByText(/Upload a premade/i)).toBeInTheDocument();
    // container.querySelector doesn't work in this scenario for some reason -> but querying document seems to work
    const uploadInput = document.querySelector('.dropzone input');
    await user.upload(uploadInput, singleFile);
    expect(uploadInput.files).toHaveLength(1);
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByPlaceholderText(destinationPlaceholder)).toBeInTheDocument());
    expect(screen.getByText('testFile.txt')).toBeInTheDocument();
    // FileSize component is not an input based component -> query text only
    expect(screen.getByText('17.00 Bytes')).toBeInTheDocument();
    await user.type(screen.getByPlaceholderText(destinationPlaceholder), '/some/path');
    await user.type(screen.getByRole('combobox', { name: /device types compatible/i }), 'something');
    await user.clear(screen.getByLabelText(/release name/i));
    await user.type(screen.getByLabelText(/release name/i), 'some release');
    await user.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() => expect(screen.getByText(/Version information/i)).toBeInTheDocument());
    // the software name follows the release name, unless it was modified by the user
    expect(screen.getByLabelText(/software name/i)).toHaveValue('some release');

    // the version information is validated on progressing only, too
    await user.clear(screen.getByLabelText(/software filesystem/i));
    expect(screen.queryByText(/is required/i)).not.toBeInTheDocument();
    await user.type(screen.getByLabelText(/software filesystem/i), 'rootfs-image');

    await user.click(screen.getByRole('button', { name: /upload/i }));
    await waitFor(() => expect(creationSpy).toHaveBeenCalled());
    expect(creationSpy).toHaveBeenCalledWith({
      file: singleFile,
      meta: {
        args: {
          dest_dir: '/some/path',
          filename: 'testFile.txt',
          software_filesystem: 'rootfs-image',
          software_name: 'some release',
          software_version: '1.0.0'
        },
        description: '',
        device_types_compatible: ['something'],
        name: 'some release'
      }
    });
  });
});
