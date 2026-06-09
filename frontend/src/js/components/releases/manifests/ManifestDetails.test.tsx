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
import { paperClasses } from '@mui/material';

import { defaultState, render } from '@/testUtils';
import { ColumnWidthProvider } from '@northern.tech/common-ui/TwoColumnData';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { ComponentTypesTable } from './ComponentTypesTable';
import { ManifestDetails } from './ManifestDetails';

const preloadedState = {
  ...defaultState,
  releases: {
    ...defaultState.releases,
    selectedManifest: 'm1000'
  }
};

describe('ManifestDetails Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <ColumnWidthProvider>
        <ManifestDetails />
      </ColumnWidthProvider>,
      { preloadedState }
    );
    const view = baseElement.querySelector(`.${paperClasses.root}`);
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('does not render when no manifest is selected', () => {
    const { baseElement } = render(
      <ColumnWidthProvider>
        <ManifestDetails />
      </ColumnWidthProvider>
    );
    const view = baseElement.querySelector(`.${paperClasses.root}`);
    expect(view).toBeFalsy();
  });

  it('allows editing manifest notes', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { store } = render(
      <ColumnWidthProvider>
        <ManifestDetails />
      </ColumnWidthProvider>,
      { preloadedState }
    );
    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    await user.click(editButtons[0]);
    const textField = screen.getByRole('textbox');
    await user.clear(textField);
    await user.type(textField, 'new manifest notes');
    await user.click(screen.getByRole('button', { name: /confirm/i }));
    await waitFor(() => expect(store.getState().releases.manifestsById.m1000.notes).toBe('new manifest notes'));
  });

  it('allows editing manifest tags', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { store } = render(
      <ColumnWidthProvider>
        <ManifestDetails />
      </ColumnWidthProvider>,
      { preloadedState }
    );
    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    await user.click(editButtons[1]);
    const tagsInput = document.querySelector('#tags-chip-select') as HTMLInputElement;
    await user.type(tagsInput, 'new-tag');
    await user.keyboard('{Enter}');
    await user.click(screen.getByRole('button', { name: /confirm/i }));
    await waitFor(() => expect(store.getState().releases.manifestsById.m1000.tags).toContain('new-tag'));
  });
});

const componentTypes = {
  'comp-a': { artifact_name: 'existing-release', update_strategy: { order: 1 } },
  'comp-b': { artifact_name: 'missing-release', update_strategy: { order: 2 } },
  'comp-c': { artifact_path: '/path/to/artifact', update_strategy: { order: 3 } }
};

describe('ComponentTypesTable', () => {
  it('renders a clickable link for releases that exist', () => {
    render(<ComponentTypesTable componentTypes={componentTypes} existingReleases={{ 'existing-release': true, 'missing-release': false }} />);
    const link = screen.getByRole('button', { name: 'existing-release' });
    expect(link).toBeInTheDocument();
    expect(link.closest('a, button')).toBeTruthy();
  });

  it('renders plain text for releases when existingReleases marks them as non-existent and not in creation mode', () => {
    render(<ComponentTypesTable componentTypes={componentTypes} existingReleases={{ 'existing-release': false, 'missing-release': false }} />);
    expect(screen.getByText('existing-release')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'existing-release' })).not.toBeInTheDocument();
    expect(screen.queryByText(/not available/i)).not.toBeInTheDocument();
  });

  it('renders a warning for non-existent releases during creation', () => {
    render(<ComponentTypesTable componentTypes={componentTypes} existingReleases={{ 'existing-release': true, 'missing-release': false }} isCreation />);
    expect(screen.getByRole('button', { name: 'existing-release' })).toBeInTheDocument();
    expect(screen.getByText(/not available/i)).toBeInTheDocument();
    expect(screen.getByText('missing-release')).toBeInTheDocument();
  });

  it('falls back to artifact_path when no artifact_name is present', () => {
    render(<ComponentTypesTable componentTypes={{ 'path-only': { artifact_path: '/some/path', update_strategy: { order: 1 } } }} />);
    expect(screen.getByText('/some/path')).toBeInTheDocument();
  });
});
