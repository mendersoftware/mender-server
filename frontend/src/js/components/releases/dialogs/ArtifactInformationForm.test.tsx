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
import { FormProvider, useForm } from 'react-hook-form';

import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { ArtifactInformationForm } from './ArtifactInformationForm';

const file = { name: 'testFile.txt', size: 1024 };

const defaultValues = {
  destination: '',
  deviceTypes: [],
  fileSystem: 'rootfs-image',
  name: '',
  softwareName: 'testFile',
  softwareVersion: '1.0.0'
};

const TestForm = ({ activeStep = 0, onSubmit = vi.fn(), values = {} }) => {
  const methods = useForm({ defaultValues: { ...defaultValues, ...values } });
  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={methods.handleSubmit(onSubmit)}>
        <ArtifactInformationForm activeStep={activeStep} deviceTypes={[]} file={file} onRemove={vi.fn()} type="singleFile" />
        <button type="submit">Next</button>
      </form>
    </FormProvider>
  );
};

describe('ArtifactInformationForm Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<TestForm />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders the version information correctly', async () => {
    const { baseElement } = render(<TestForm activeStep={1} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('shows the release name help tip inline w/ the label', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<TestForm />);
    const [fileHelpTip, releaseNameHelpTip] = screen.getAllByTestId('HelpIcon');
    expect(screen.getByText(/Selected single file/i).parentNode).toContainElement(fileHelpTip);
    expect(screen.getByText(/Release name/i)).toContainElement(releaseNameHelpTip);
    await user.click(releaseNameHelpTip);
    await waitFor(() => expect(screen.getByText(/If a Release with this name exists/i)).toBeInTheDocument());
  });

  it('validates the entered information only on progressing', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const submitMock = vi.fn();
    render(<TestForm onSubmit={submitMock} />);
    expect(screen.queryByText(/is required/i)).not.toBeInTheDocument();
    // typing an invalid destination doesn't produce an error before trying to progress
    await user.type(screen.getByLabelText(/destination directory/i), 'opt/test');
    expect(screen.queryByText(/Destination has to be an absolute path/i)).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() => expect(screen.getByText(/Destination has to be an absolute path/i)).toBeInTheDocument());
    expect(screen.getByText(/Release name is required/i)).toBeInTheDocument();
    expect(screen.getByText(/Device type is required/i)).toBeInTheDocument();
    expect(submitMock).not.toHaveBeenCalled();

    // correcting an input clears the error it caused
    await user.clear(screen.getByLabelText(/destination directory/i));
    await user.type(screen.getByLabelText(/destination directory/i), '/opt/test');
    await waitFor(() => expect(screen.queryByText(/Destination has to be an absolute path/i)).not.toBeInTheDocument());
    expect(screen.getByText(/Where the file will be installed on your devices/i)).toBeInTheDocument();
  });

  it('allows progressing once all required information is entered', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const submitMock = vi.fn();
    render(<TestForm onSubmit={submitMock} values={{ destination: '/opt/test', deviceTypes: ['some-device-type'], name: 'some release' }} />);
    await user.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() => expect(submitMock).toHaveBeenCalled());
    expect(screen.queryByText(/is required/i)).not.toBeInTheDocument();
  });

  it('requires the version information', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const submitMock = vi.fn();
    render(<TestForm activeStep={1} onSubmit={submitMock} />);
    await user.clear(screen.getByLabelText(/software filesystem/i));
    expect(screen.queryByText(/is required/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() => expect(screen.getByText(/Software filesystem is required/i)).toBeInTheDocument());
    expect(submitMock).not.toHaveBeenCalled();
  });
});
