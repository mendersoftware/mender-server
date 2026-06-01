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
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { SettingsItem, ToggleSettingsItem } from './SettingsItem';

describe('SettingsItem Component', () => {
  it('renders correctly with string title', async () => {
    const { baseElement } = render(<SettingsItem title="Current plan" secondary="Trial" notification="upgrade now!" />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly without secondary', async () => {
    const { baseElement } = render(<SettingsItem title="Title only" description="Just a description" />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});

describe('ToggleSettingsItem Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<ToggleSettingsItem title="Test toggle" checked={false} onClick={vi.fn()} description="A test description" />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('calls onClick when switch is clicked', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onClick = vi.fn();
    render(<ToggleSettingsItem title="Clickable toggle" checked={false} onClick={onClick} />);
    await user.click(screen.getByText('Clickable toggle'));
    expect(onClick).toHaveBeenCalled();
  });
});
