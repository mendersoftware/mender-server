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
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { CustomPhaseTable } from './CustomPhases';
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

describe('CustomPhaseTable', () => {
  describe('percentage mode', () => {
    it('renders phases with percentage batch sizes', () => {
      render(
        <FormWrapper
          defaultValues={{
            phases: [{ batch_size: 30, delay: 2, delayUnit: 'hours' }, {}],
            rolloutMode: rolloutModes.percentage.key
          }}
        >
          <CustomPhaseTable deploymentDeviceCount={100} />
        </FormWrapper>
      );
      expect(screen.getByText('Phase 1')).toBeInTheDocument();
      expect(screen.getByText('Phase 2')).toBeInTheDocument();
      expect(screen.getByText('(Final step)')).toBeInTheDocument();
      expect(screen.getByText('Batch size')).toBeInTheDocument();
      expect(screen.getByText('Delay before next phase')).toBeInTheDocument();
      expect(screen.getAllByText('%').length).toBeGreaterThanOrEqual(1);
      expect(screen.getByRole('button', { name: /Add a phase/i })).toBeInTheDocument();
    });

    it('adds a phase when clicking the add button', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      render(
        <FormWrapper
          defaultValues={{
            phases: [{ batch_size: 10, delay: 2, delayUnit: 'hours' }, {}],
            rolloutMode: rolloutModes.percentage.key
          }}
        >
          <CustomPhaseTable deploymentDeviceCount={100} />
        </FormWrapper>
      );
      expect(screen.queryByText('Phase 3')).not.toBeInTheDocument();
      await user.click(screen.getByRole('button', { name: /Add a phase/i }));
      expect(screen.getByText('Phase 3')).toBeInTheDocument();
      expect(screen.getByText('(Final step)')).toBeInTheDocument();
    });

    it('shows remove and repeat buttons on non-final phases', () => {
      render(
        <FormWrapper
          defaultValues={{
            phases: [{ batch_size: 20, delay: 2, delayUnit: 'hours' }, { batch_size: 30, delay: 1, delayUnit: 'hours' }, {}],
            rolloutMode: rolloutModes.percentage.key
          }}
        >
          <CustomPhaseTable deploymentDeviceCount={100} />
        </FormWrapper>
      );
      expect(screen.getAllByTitle('Repeat phase')).toHaveLength(2);
      expect(screen.getAllByTitle('Remove phase')).toHaveLength(2);
    });
  });

  describe('device_count mode', () => {
    it('renders phases with device count batch sizes', () => {
      render(
        <FormWrapper
          defaultValues={{
            phases: [{ batch_size_devices: 30, delay: 2, delayUnit: 'hours' }, {}],
            rolloutMode: rolloutModes.device_count.key
          }}
        >
          <CustomPhaseTable deploymentDeviceCount={100} />
        </FormWrapper>
      );
      expect(screen.getByText('Phase 1')).toBeInTheDocument();
      expect(screen.getByText('Phase 2')).toBeInTheDocument();
      expect(screen.getByText('(Final step)')).toBeInTheDocument();
      // device_count mode should not render % adornment
      expect(screen.queryByText('%')).not.toBeInTheDocument();
      const rows = screen.getAllByRole('row');
      // header + 2 phase rows
      expect(rows).toHaveLength(3);
      // verify device count annotations are present
      const phase1Row = rows[1];
      expect(within(phase1Row).getByText(/device/i)).toBeInTheDocument();
    });

    it('adds a phase in device_count mode', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      render(
        <FormWrapper
          defaultValues={{
            phases: [{ batch_size_devices: 20, delay: 2, delayUnit: 'hours' }, {}],
            rolloutMode: rolloutModes.device_count.key
          }}
        >
          <CustomPhaseTable deploymentDeviceCount={100} />
        </FormWrapper>
      );
      await user.click(screen.getByRole('button', { name: /Add a phase/i }));
      expect(screen.getByText('Phase 3')).toBeInTheDocument();
      expect(screen.queryByText('%')).not.toBeInTheDocument();
    });
  });
});
