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
import { ColumnWidthProvider } from '@northern.tech/common-ui/TwoColumnData';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import ArtifactDetails, { transformArtifactCapabilities, transformArtifactMetadata } from './ArtifactDetails';

describe('ArtifactDetails Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <ColumnWidthProvider>
        <ArtifactDetails
          artifact={{
            artifact_provides: {
              artifact_name: 'myapp',
              'data-partition.myapp.version': 'v2020.10',
              list_of_fancy: ['x172']
            },
            description: 'text',
            name: 'test'
          }}
        />
      </ColumnWidthProvider>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly without software', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { baseElement } = render(
      <ColumnWidthProvider>
        <ArtifactDetails
          artifact={{
            artifact_provides: {
              list_of_fancy: ['x172']
            },
            description: 'text',
            name: 'test'
          }}
        />
      </ColumnWidthProvider>
    );
    await user.click(screen.getByText(/Provides and Depends/i));

    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});

describe('transformArtifactCapabilities', () => {
  it('works as expected', async () => {
    expect(transformArtifactCapabilities(defaultState.releases.byId.r1.artifacts[0].artifact_provides)).toEqual({
      'artifact_name': 'myapp',
      'data-partition.myapp.version': 'v2020.10',
      'list_of_fancy-1': 'qemux86-64',
      'list_of_fancy-2': 'x172'
    });
    expect(transformArtifactCapabilities(defaultState.releases.byId.r1.artifacts[0].clears_artifact_provides)).toEqual({ '0': 'data-partition.myapp.*' });
    expect(transformArtifactCapabilities(defaultState.releases.byId.r1.artifacts[0].artifact_depends)).toEqual({});
  });
});
describe('transformArtifactMetadata', () => {
  it('works as expected', async () => {
    expect(transformArtifactMetadata({ thing: 'thang', more: ['like', 'a', 'list'], or: { anObject: true }, less: undefined })).toEqual({
      thing: 'thang',
      more: 'like,a,list',
      or: '{"anObject":true}',
      less: '-'
    });
  });
});
