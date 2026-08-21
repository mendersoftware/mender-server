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
import { FormProvider, useForm } from 'react-hook-form';

import { render } from '@/testUtils';
import { screen } from '@testing-library/react';

import { UniformPhaseSettings } from './UniformPhases';
import { rolloutModes } from './constants';

const FormWrapper = ({ children, defaultValues = {} }) => {
  const methods = useForm({
    defaultValues: {
      phases: [],
      rolloutMode: rolloutModes.percentage.key,
      maxDevices: 0,
      uniform_phases: undefined,
      ...defaultValues
    }
  });
  return <FormProvider {...methods}>{children}</FormProvider>;
};

describe('UniformPhaseSettings', () => {
  it('renders in percentage mode with summary', () => {
    render(
      <FormWrapper
        defaultValues={{
          rolloutMode: rolloutModes.percentage.key,
          uniform_phases: { batch_size: 10, time_interval: '7200s' }
        }}
      >
        <UniformPhaseSettings deploymentDeviceCount={100} />
      </FormWrapper>
    );
    expect(screen.getByText('Batch size')).toBeInTheDocument();
    expect(screen.getByText('Delay before next phase')).toBeInTheDocument();
    expect(screen.getByText('Summary')).toBeInTheDocument();
    expect(screen.getAllByText('%').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText(/phases with 10 devices/)).toBeInTheDocument();
    expect(screen.getByText(/100 devices total/)).toBeInTheDocument();
  });

  it('renders in device_count mode with summary', () => {
    render(
      <FormWrapper
        defaultValues={{
          rolloutMode: rolloutModes.device_count.key,
          uniform_phases: { batch_size_devices: 50, time_interval: '7200s' }
        }}
      >
        <UniformPhaseSettings deploymentDeviceCount={100} />
      </FormWrapper>
    );
    expect(screen.getByText('Batch size')).toBeInTheDocument();
    expect(screen.getByText('Summary')).toBeInTheDocument();
    expect(screen.queryByText('%')).not.toBeInTheDocument();
    expect(screen.getByText(/phases with 50 devices/)).toBeInTheDocument();
    expect(screen.getByText(/100 devices total/)).toBeInTheDocument();
  });

  it('renders with dynamic group filter showing open-ended summary', () => {
    render(
      <FormWrapper
        defaultValues={{
          rolloutMode: rolloutModes.percentage.key,
          uniform_phases: { batch_size: 25, time_interval: '3600s' }
        }}
      >
        <UniformPhaseSettings deploymentDeviceCount={200} filter={{ id: 'f1', name: 'dynamic-group' }} />
      </FormWrapper>
    );
    expect(screen.getByText('Summary')).toBeInTheDocument();
    expect(screen.getByText(/Deploy in phases of/)).toBeInTheDocument();
    expect(screen.getByText(/delay between phases/)).toBeInTheDocument();
  });
});
