// Copyright 2020 Northern.tech AS
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

import { screen, waitFor } from '@testing-library/react';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import DeviceStatus from './device-status';

describe('DeviceStatus Component', () => {
  it('renders correctly', async () => {
    let ui = <DeviceStatus device={{ auth_sets: [{ status: 'pending' }] }} />;
    const { baseElement, rerender } = render(ui);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    ui = <DeviceStatus device={{ monitor: [{ a: 'b' }] }} />;
    render(ui);
    await waitFor(() => rerender(ui));
    expect(screen.getAllByText(/monitoring/i)[0]).toBeInTheDocument();
    ui = <DeviceStatus device={{ isOffline: true }} />;
    render(ui);
    await waitFor(() => rerender(ui));
    expect(screen.getAllByText(/offline/i)[0]).toBeInTheDocument();
  });
});
