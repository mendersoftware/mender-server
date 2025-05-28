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
import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Artifact from './Artifact';
import { columns } from './ReleaseDetails';

describe('Artifact Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <Artifact
        artifact={{
          artifact_provides: {
            artifact_name: 'myapp',
            'data-partition.myapp.version': 'v2020.10',
            list_of_fancy: ['x172']
          },
          device_types_compatible: ['test-type'],
          updates: [],
          modified: '2019-01-01'
        }}
        index={0}
        columns={columns}
      />
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
