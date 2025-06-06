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
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import RedirectionWidget from './RedirectionWidget';

describe('RedirectionWidget Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<RedirectionWidget onClick={vi.fn} content="testlocation" />);
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const content = 'test content';
    const submitCheck = vi.fn();
    render(<RedirectionWidget content={content} onClick={submitCheck} />);

    await user.click(screen.getByText(content));
    expect(screen.queryByText('redirected')).not.toBeInTheDocument();
    expect(submitCheck).toHaveBeenCalledTimes(1);
    await user.click(screen.getByText(content));
    expect(submitCheck).toHaveBeenCalledTimes(2);
  });
});
