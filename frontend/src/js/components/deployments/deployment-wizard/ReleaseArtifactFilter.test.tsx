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
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { SoftwareArtifactFilter } from './ReleaseArtifactFilter';

describe('ReleaseArtifactFilter Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <SoftwareArtifactFilter selectedSoftware={defaultState.releases.byId.r1.name} open={true} onSelect={vi.fn} onClose={vi.fn} />
    );
    expect(baseElement).toMatchSnapshot();
    expect(baseElement).toEqual(expect.not.stringMatching(undefineds));
  });

  it('searches immediately on enter instead of submitting the form natively', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { getSoftwareList: getSoftwareListSpy } = StoreThunks;
    const submitEvents: SubmitEvent[] = [];
    const onSubmit = (event: SubmitEvent) => submitEvents.push(event);
    document.addEventListener('submit', onSubmit);
    render(<SoftwareArtifactFilter open={true} onSelect={vi.fn()} onClose={vi.fn()} />);
    await user.type(screen.getByPlaceholderText(/search software/i), 'foo');
    await user.keyboard('{Enter}');
    expect(submitEvents).toHaveLength(1);
    expect(submitEvents[0].defaultPrevented).toBeTruthy();
    expect(getSoftwareListSpy).toHaveBeenLastCalledWith(expect.objectContaining({ searchTerm: 'foo' }));
    document.removeEventListener('submit', onSubmit);
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
  });
});
