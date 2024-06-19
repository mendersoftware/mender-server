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
import React from 'react';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import ArtifactDetails, { transformArtifactCapabilities, transformArtifactMetadata } from './artifactdetails';

describe('ArtifactDetails Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<ArtifactDetails artifact={{ description: 'text', name: 'test' }} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});

describe('transformArtifactCapabilities', () => {
  it('works as expected', async () => {
    expect(transformArtifactCapabilities(defaultState.releases.byId.r1.artifacts[0].artifact_provides)).toEqual([
      { key: 'artifact_name', primary: 'artifact_name', secondary: 'myapp' },
      { key: 'data-partition.myapp.version', primary: 'data-partition.myapp.version', secondary: 'v2020.10' },
      { key: 'list_of_fancy-1', primary: 'list_of_fancy-1', secondary: 'qemux86-64' },
      { key: 'list_of_fancy-2', primary: 'list_of_fancy-2', secondary: 'x172' }
    ]);
    expect(transformArtifactCapabilities(defaultState.releases.byId.r1.artifacts[0].clears_artifact_provides)).toEqual([
      { key: '0', primary: '0', secondary: 'data-partition.myapp.*' }
    ]);
    expect(transformArtifactCapabilities(defaultState.releases.byId.r1.artifacts[0].artifact_depends)).toEqual([]);
  });
});
describe('transformArtifactMetadata', () => {
  it('works as expected', async () => {
    expect(transformArtifactMetadata({ thing: 'thang', more: ['like', 'a', 'list'], or: { anObject: true }, less: undefined })).toEqual([
      { key: 'thing', primary: 'thing', secondary: 'thang', secondaryTypographyProps: { component: 'div' } },
      { key: 'more', primary: 'more', secondary: 'like,a,list', secondaryTypographyProps: { component: 'div' } },
      { key: 'or', primary: 'or', secondary: '{"anObject":true}', secondaryTypographyProps: { component: 'div' } },
      { key: 'less', primary: 'less', secondary: '-', secondaryTypographyProps: { component: 'div' } }
    ]);
  });
});
