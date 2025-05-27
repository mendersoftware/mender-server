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
import { getGroups } from '@northern.tech/store/selectors';
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Groups, { GroupItem, GroupsSubheader } from './Groups';

describe('Groups Component', () => {
  it('renders correctly', async () => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { groupNames, ...groups } = getGroups(defaultState);
    const { baseElement } = render(<Groups groups={groups} openGroupDialog={vi.fn} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  [GroupItem, GroupsSubheader].forEach(async Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(<Component heading="test" name="test" />);
      const view = baseElement.firstChild.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});
