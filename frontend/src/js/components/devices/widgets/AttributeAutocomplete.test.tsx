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
import { render } from '@/testUtils';
import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import AttributeAutoComplete from './AttributeAutocomplete';

describe('AttributeAutoComplete Component', () => {
  const attributes = [
    { key: 'mender-orchestrator-manifest.version', value: 'version', scope: 'inventory', category: 'recently used', priority: 0 },
    { key: 'updated_ts', value: 'Last inventory update', scope: 'system', category: 'recently used', priority: 0 },
    { key: 'mac', value: 'mac', scope: 'identity', category: 'identity', priority: 1 },
    { key: 'artifact_name', value: 'artifact_name', scope: 'inventory', category: 'inventory', priority: 2 },
    { key: 'mender-orchestrator-manifest.component_type', value: 'component_type', scope: 'inventory', category: 'system', priority: 3 },
    { key: 'created_ts', value: 'First request', scope: 'system', category: 'default', priority: 4 }
  ];
  it('groups attributes under the relabeled scope groups', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(<AttributeAutoComplete attributes={attributes} onRemove={vi.fn()} onSelect={vi.fn()} />);
    await user.click(screen.getByRole('combobox'));
    const groupLabels = [...baseElement.querySelectorAll('.MuiAutocomplete-groupLabel')].map(item => item.textContent);
    expect(groupLabels).toEqual(['recently used', 'identity', 'inventory', 'system', 'default']);
    expect(screen.getByRole('option', { name: /component_type/i })).toBeVisible();
  });
  it('shows relabeled scopes on recently used entries', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<AttributeAutoComplete attributes={attributes} onRemove={vi.fn()} onSelect={vi.fn()} />);
    await user.click(screen.getByRole('combobox'));
    expect(screen.getByText('(system)')).toBeVisible();
    expect(screen.getByText('(default)')).toBeVisible();
  });
  it('selects orchestrator manifest attributes with the full key and inventory scope', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onSelect = vi.fn();
    render(<AttributeAutoComplete attributes={attributes} onRemove={vi.fn()} onSelect={onSelect} />);
    await user.click(screen.getByRole('combobox'));
    await user.click(screen.getByRole('option', { name: /component_type/i }));
    await act(async () => vi.runOnlyPendingTimers());
    expect(onSelect).toHaveBeenLastCalledWith({ key: 'mender-orchestrator-manifest.component_type', scope: 'inventory' });
  });
});
