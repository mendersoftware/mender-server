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
import { getGroups } from '@northern.tech/store/selectors';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

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

  it('allows removing a group through the group actions menu', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onRemove = vi.fn();
    const changeGroup = vi.fn();
    render(<GroupItem canRemove changeGroup={changeGroup} name="testGroup" onRemove={onRemove} />);
    // the actions button is only visible on hover, so it has no accessible name to query by role with
    await user.click(screen.getByLabelText(/group actions/i));
    await user.click(screen.getByRole('menuitem', { name: /delete group/i }));
    expect(onRemove).toHaveBeenCalledWith('testGroup');
    expect(changeGroup).not.toHaveBeenCalled();
  });

  it('prevents removing a group without the required permissions', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onRemove = vi.fn();
    render(<GroupItem changeGroup={vi.fn} name="testGroup" onRemove={onRemove} />);
    await user.click(screen.getByLabelText(/group actions/i));
    expect(screen.getByRole('menuitem', { name: /delete group/i })).toHaveAttribute('aria-disabled', 'true');
    expect(onRemove).not.toHaveBeenCalled();
  });
});
