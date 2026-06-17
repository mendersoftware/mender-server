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
import * as ReleasesThunks from '@northern.tech/store/releasesSlice/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, vi } from 'vitest';

import { AddManifestDrawer } from './ManifestDrawer';

vi.mock('@northern.tech/store/releasesSlice/thunks', { spy: true });

const yamlContent = `api_version: v1
kind: manifest
name: my-manifest
system_types_compatible:
  - device-type-a
  - device-type-b
component_types:
  rootfs:
    artifact_name: something
    update_strategy:
      order: 1
`;

describe('AddManifestDrawer Component', () => {
  beforeEach(() => {
    vi.mocked(ReleasesThunks.uploadManifest).mockClear();
    vi.mocked(ReleasesThunks.generateManifest).mockClear();
  });

  it('renders correctly', async () => {
    const { baseElement } = render(<AddManifestDrawer open onClose={vi.fn()} />);
    const view = baseElement.querySelector('.MuiDrawer-paper');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('dispatches uploadManifest with the selected file, description, and tags', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(<AddManifestDrawer open onClose={vi.fn()} />);

    const file = new File(['manifest content'], 'test.mender', { type: 'application/octet-stream' });
    const dropzoneInput = baseElement.querySelector('.dropzone input') as HTMLInputElement;
    await user.upload(dropzoneInput, file);

    await user.type(screen.getByPlaceholderText(/add notes here/i), 'unit test note');
    await user.type(screen.getByPlaceholderText(/add tags/i), 'unit-tag{enter}');

    await user.click(screen.getByRole('button', { name: /^upload$/i }));

    expect(ReleasesThunks.uploadManifest).toHaveBeenCalledWith({
      file: expect.objectContaining({ name: 'test.mender' }),
      meta: { description: 'unit test note', name: '', tags: ['unit-tag'] }
    });
  });

  it('parses a valid YAML manifest, shows the extracted metadata, and uploads it', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(<AddManifestDrawer open onClose={vi.fn()} />);

    File.prototype.text = vi.fn().mockResolvedValue(yamlContent);
    const file = new File([yamlContent], 'test.yaml', { type: 'application/yaml' });
    const dropzoneInput = baseElement.querySelector('.dropzone input') as HTMLInputElement;
    await user.upload(dropzoneInput, file);

    expect(await screen.findByText('my-manifest')).toBeInTheDocument();
    expect(screen.getByText('device-type-a, device-type-b')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /^upload$/i }));

    expect(ReleasesThunks.generateManifest).toHaveBeenCalledWith({
      file: expect.objectContaining({ name: 'test.yaml' }),
      meta: { description: '', name: '', tags: [] }
    });
  });

  it('shows an error when a multi-document YAML is uploaded', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(<AddManifestDrawer open onClose={vi.fn()} />);

    const multiDocYaml = `---
${yamlContent}
---
kind: manifest
name: missing-required-fields
`;
    File.prototype.text = vi.fn().mockResolvedValue(multiDocYaml);
    const file = new File([multiDocYaml], 'multi.yaml', { type: 'application/yaml' });
    const dropzoneInput = baseElement.querySelector('.dropzone input') as HTMLInputElement;
    await user.upload(dropzoneInput, file);

    expect(await screen.findByDisplayValue('multi.yaml')).toBeInTheDocument();
    expect(screen.getByText(/only single-document/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^upload$/i })).toBeDisabled();
  });

  it('shows an error when an empty YAML manifest is uploaded', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(<AddManifestDrawer open onClose={vi.fn()} />);

    File.prototype.text = vi.fn().mockResolvedValue('');
    const file = new File([''], 'empty.yaml', { type: 'application/yaml' });
    const dropzoneInput = baseElement.querySelector('.dropzone input') as HTMLInputElement;
    await user.upload(dropzoneInput, file);

    expect(await screen.findByDisplayValue('empty.yaml')).toBeInTheDocument();
    expect(screen.getByText(/the manifest file is empty/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^upload$/i })).toBeDisabled();
  });

  it('shows a validation error when an invalid YAML manifest is uploaded', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(<AddManifestDrawer open onClose={vi.fn()} />);

    const invalidYaml = `kind: manifest
name: missing-required-fields
`;
    File.prototype.text = vi.fn().mockResolvedValue(invalidYaml);
    const file = new File([invalidYaml], 'broken.yaml', { type: 'application/yaml' });
    const dropzoneInput = baseElement.querySelector('.dropzone input') as HTMLInputElement;
    await user.upload(dropzoneInput, file);

    expect(await screen.findByDisplayValue('broken.yaml')).toBeInTheDocument();
    expect(screen.getByText(/at api_version/)).toBeInTheDocument();
    expect(screen.getByText(/at system_types_compatible/)).toBeInTheDocument();
    expect(screen.getByText(/at component_types/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^upload$/i })).toBeDisabled();
    expect(ReleasesThunks.uploadManifest).not.toHaveBeenCalled();
  });
});
