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
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import DeviceNotifications from './DeviceNotifications';

const preloadedState = {
  ...defaultState,
  devices: {
    ...defaultState.devices,
    limits: {
      standard: 250,
      micro: 0,
      system: 0
    },
    byStatus: {
      ...defaultState.devices.byStatus,
      accepted: {
        ...defaultState.devices.byStatus,
        counts: {
          standard: 100,
          micro: 0,
          system: 0,
          total: 100
        }
      }
    }
  }
};
const preloadedState240Devices = {
  ...preloadedState,
  devices: {
    ...preloadedState.devices,
    byStatus: {
      ...preloadedState.devices.byStatus,
      accepted: {
        ...preloadedState.devices.byStatus,
        counts: {
          standard: 240,
          micro: 0,
          system: 0,
          total: 240
        }
      }
    }
  }
};
const preloadedState250Devices = {
  ...preloadedState,
  devices: {
    ...preloadedState.devices,
    byStatus: {
      ...preloadedState.devices.byStatus,
      accepted: {
        ...preloadedState.devices.byStatus,
        counts: {
          standard: 250,
          micro: 0,
          system: 0,
          total: 250
        }
      }
    }
  }
};
describe('DeviceNotifications Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceNotifications pending={10} total={100} />, {
      preloadedState: { ...preloadedState, devices: { ...preloadedState.devices, limits: { ...defaultState.devices.limits, standard: 1000 } } }
    });
    const view = baseElement.firstChild?.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly with limits', async () => {
    const { baseElement } = render(<DeviceNotifications total={40} pending={5} />, {
      preloadedState: { ...preloadedState, devices: { ...preloadedState.devices } }
    });
    const view = baseElement.firstChild?.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly close to limits', async () => {
    const { baseElement } = render(<DeviceNotifications total={240} pending={5} />, { preloadedState: preloadedState240Devices });
    const view = baseElement.firstChild?.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly at limit', async () => {
    const { baseElement } = render(<DeviceNotifications total={250} pending={5} />, { preloadedState: preloadedState250Devices });
    const view = baseElement.firstChild?.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly without limits', async () => {
    const { baseElement } = render(<DeviceNotifications total={240} pending={5} />);
    const view = baseElement.firstChild?.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('popover renders correctly when opened', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });

    render(<DeviceNotifications total={240} pending={5} />, { preloadedState });
    await user.click(screen.getByRole('button', { name: '240' }));

    const popover = await screen.findByRole('presentation');
    expect(popover).toMatchSnapshot();
  });
});
