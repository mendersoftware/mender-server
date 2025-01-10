// Copyright 2025 Northern.tech AS
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

import { TIMEOUTS } from '@northern.tech/store/commonConstants';
import { screen, waitFor } from '@testing-library/react';

import { render } from '../../../tests/setupTests';
import { InputErrorNotification } from './InputErrorNotification';

describe('InfoHint Component', () => {
  it('renders correctly', async () => {
    const ui = <InputErrorNotification content="test" className="some-class" />;
    const { rerender } = render(ui);
    expect(screen.getByText('test')).toBeVisible();
    await jest.advanceTimersByTimeAsync(TIMEOUTS.fiveSeconds);
    await waitFor(() => rerender(ui));
    expect(screen.getByText('test')).toHaveClass('fadeOut');
  });
});
