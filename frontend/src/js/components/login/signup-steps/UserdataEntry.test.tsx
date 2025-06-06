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
import { formRenderWrapper } from '@northern.tech/common-ui/forms/helpers';
import { act } from '@testing-library/react';
import { vi } from 'vitest';

import { undefineds } from '../../../../../tests/mockData';
import { UserDataEntry } from './UserdataEntry';

describe('Login Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(<UserDataEntry classes={{ userData: 'test' }} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(() => vi.runAllTimersAsync());
  });
});
