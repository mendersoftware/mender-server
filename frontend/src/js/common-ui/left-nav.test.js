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

import { undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import LeftNav from './left-nav';

describe('LeftNav Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <LeftNav
        sections={[
          {
            itemClass: 'test-list',
            items: [
              { exact: true, path: '/far/away', icon: <div>HEYA!</div>, style: { background: 'red' }, title: 'test1' },
              { path: '/further/away', title: 'test1' }
            ],
            title: 'test'
          }
        ]}
      />
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
