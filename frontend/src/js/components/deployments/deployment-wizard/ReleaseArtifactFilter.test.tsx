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
import { defaultState, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { SoftwareArtifactFilter } from './ReleaseArtifactFilter';

describe('ReleaseArtifactFilter Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <SoftwareArtifactFilter selectedSoftware={{ ...defaultState.releases.byId.r1, kind: 'release' }} open={true} onSelect={vi.fn} onClose={vi.fn} />
    );
    expect(baseElement).toMatchSnapshot();
    expect(baseElement).toEqual(expect.not.stringMatching(undefineds));
  });
  it('shows the selected software first and confirms it on click', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onClose = vi.fn();
    const onSelect = vi.fn();
    const selectedSoftware = { kind: 'release' as const, name: 'release-5', modified: '2020-09-10T12:16:04.000Z' };
    render(<SoftwareArtifactFilter selectedSoftware={selectedSoftware} open={true} onSelect={onSelect} onClose={onClose} />);
    await act(() => vi.advanceTimersByTimeAsync(1000));
    const pinnedItem = document.querySelector('#deployment-release-container')!.firstElementChild!;
    expect(pinnedItem).toHaveTextContent(/release-5(?!\d)/);
    await user.click(pinnedItem);
    expect(onSelect).toHaveBeenCalledWith(selectedSoftware);
    expect(onClose).toHaveBeenCalled();
  });
});
