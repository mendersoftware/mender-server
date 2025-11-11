// Copyright 2025 Northern.tech AS
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
import { TIMEOUTS } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import ArtifactGeneration from './ArtifactGeneration';

vi.mock('@northern.tech/store/thunks', { spy: true });

describe('ArtifactGeneration component', () => {
  it(`renders correctly`, async () => {
    const { baseElement } = render(<ArtifactGeneration onClose={vi.fn()} open />);
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it(`works as expected`, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { saveDeltaDeploymentsConfig: deltaConfigUpdate, getDeploymentsConfig: deltaConfigRetrieval } = StoreThunks;

    const onCloseSpy = vi.fn();
    render(<ArtifactGeneration onClose={onCloseSpy} open />);
    await waitFor(() => expect(deltaConfigRetrieval).toHaveBeenCalled());
    await act(async () => vi.advanceTimersByTimeAsync(TIMEOUTS.oneSecond));
    await user.click(await screen.findByRole('button', { name: /cancel/i }));
    expect(onCloseSpy).toHaveBeenCalled();
    onCloseSpy.mockReset();
    const checksumCheckbox = await screen.findByLabelText(/disable checksum/i);
    expect(checksumCheckbox).not.toBeChecked();
    await user.click(checksumCheckbox);
    expect(checksumCheckbox).toBeChecked();
    const sourceBufferInput = await screen.findByLabelText(/source buffer/i);
    await user.clear(sourceBufferInput);
    await user.type(sourceBufferInput, '55'); // 55 gets treated as an evil value by the MSW backend
    await act(async () => vi.advanceTimersByTime(TIMEOUTS.oneSecond));
    await user.click(await screen.findByRole('button', { name: /save/i }));
    expect(onCloseSpy).not.toHaveBeenCalled();
    await user.clear(sourceBufferInput);
    await user.type(sourceBufferInput, '66');
    await act(async () => vi.advanceTimersByTime(TIMEOUTS.oneSecond));
    await user.click(await screen.findByRole('button', { name: /save/i }));
    expect(deltaConfigUpdate).toHaveBeenCalledWith({
      compressionLevel: 0,
      disableChecksum: true,
      disableDecompression: false,
      duplicatesWindow: 5,
      inputWindow: 5,
      instructionBuffer: 5,
      sourceWindow: 66,
      timeout: -1
    });
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    expect(onCloseSpy).toHaveBeenCalled();
  });

  it('resets the form to defaults when clicking Reset to defaults', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<ArtifactGeneration onClose={vi.fn()} open />);
    await act(async () => {
      vi.advanceTimersByTimeAsync(TIMEOUTS.oneSecond);
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    const compressionInput = await screen.findByLabelText(/compression level/i);
    await user.clear(compressionInput);
    await user.type(compressionInput, '7');
    expect(compressionInput).toHaveValue('7');
    const resetButton = await screen.findByText(/reset to defaults/i);
    await user.click(resetButton);
    await act(() => vi.advanceTimersByTimeAsync(TIMEOUTS.oneSecond));
    expect(compressionInput).not.toHaveValue('7');
  });
});
