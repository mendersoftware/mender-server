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
import Linkify from 'react-linkify';

import { act, screen, render as testLibRender, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import 'jsdom-worker';

import { defaultState, mockDate, token, undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import * as DeviceActions from '../actions/deviceActions';
import GeneralApi from '../api/general-api';
import { getSessionInfo, maxSessionAge } from '../auth';
import { TIMEOUTS } from '../constants/appConstants';
import App, { AppProviders } from './app';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    trackerCode: 'testtracker',
    versionInformation: {
      Integration: 'next'
    }
  },
  deployments: {
    ...defaultState.deployments,
    byId: {},
    byStatus: {
      ...defaultState.deployments.byStatus,
      inprogress: {
        ...defaultState.deployments.byStatus.inprogress,
        total: 0
      }
    },
    deploymentDeviceLimit: null
  }
};

const reportsSpy = jest.spyOn(DeviceActions, 'getReportsDataWithoutBackendSupport');

jest.mock('react-linkify');

describe('App Component', () => {
  beforeAll(() => {
    Linkify.default = jest.fn();
    Linkify.default.mockReturnValue(null);
  });
  it(
    'renders correctly',
    async () => {
      jest.replaceProperty(window.mender_environment, 'integrationVersion', 'next');

      const ui = <App />;
      const { asFragment, rerender } = render(ui, {
        preloadedState: { ...preloadedState, users: { ...preloadedState.users, currentSession: getSessionInfo() } }
      });
      await waitFor(() => expect(screen.queryByText(/see all deployments/i)).toBeInTheDocument(), { timeout: TIMEOUTS.fiveSeconds });
      await act(async () => {
        jest.runOnlyPendingTimers();
        jest.runAllTicks();
        return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
      });
      await waitFor(() => expect(reportsSpy).toHaveBeenCalled(), { timeout: TIMEOUTS.fiveSeconds });
      await waitFor(() => rerender(ui));
      const view = asFragment();
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
      await act(async () => {
        jest.runOnlyPendingTimers();
        jest.runAllTicks();
      });
      reportsSpy.mockClear();
      await act(async () => jest.runAllTicks());
    },
    20 * TIMEOUTS.oneSecond
  );

  it(
    'works as intended',
    async () => {
      const currentSession = { expiresAt: new Date().toISOString(), token };
      window.localStorage.getItem.mockImplementation(name => (name === 'JWT' ? JSON.stringify(currentSession) : undefined));

      const ui = <App />;
      const { rerender } = render(ui, {
        preloadedState: { ...preloadedState, users: { ...preloadedState.users, currentSession, currentUser: 'a1' } }
      });
      await waitFor(() => expect(reportsSpy).toHaveBeenCalled(), { timeout: TIMEOUTS.threeSeconds });
      await act(async () => {
        jest.advanceTimersByTime(maxSessionAge * 1000 + 500);
        jest.runAllTicks();
        jest.runOnlyPendingTimers();
      });
      await waitFor(() => rerender(ui));
      await waitFor(() => expect(screen.queryByText(/Version:/i)).not.toBeInTheDocument(), { timeout: TIMEOUTS.fiveSeconds });
      expect(screen.queryByText(/Northern.tech/i)).toBeInTheDocument();
      expect(screen.queryByText(`© ${mockDate.getFullYear()} Northern.tech`)).toBeInTheDocument();
      await act(async () => {
        jest.runOnlyPendingTimers();
        jest.runAllTicks();
      });

      reportsSpy.mockClear();
      window.localStorage.getItem.mockReset();
    },
    20 * TIMEOUTS.oneSecond
  );

  it(
    'is embedded in working providers',
    async () => {
      window.localStorage.getItem.mockImplementation(name => (name === 'JWT' ? JSON.stringify({ token }) : undefined));
      // eslint-disable-next-line
      const ui = <AppProviders basename="" />;
      const { baseElement, rerender } = testLibRender(ui);
      await waitFor(() => screen.queryByText('Software distribution'), { timeout: TIMEOUTS.fiveSeconds });
      await waitFor(() => rerender(ui));
      const view = baseElement.lastElementChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
      await waitFor(() => expect(reportsSpy).toHaveBeenCalled(), { timeout: TIMEOUTS.fiveSeconds });
      await act(async () => {
        jest.runOnlyPendingTimers();
        jest.runAllTicks();
        return new Promise(resolve => resolve(), TIMEOUTS.fiveSeconds);
      });
      window.localStorage.getItem.mockReset();
    },
    20 * TIMEOUTS.oneSecond
  );
  it(
    'allows offline threshold migration',
    async () => {
      window.localStorage.getItem.mockImplementation(name => (name === 'JWT' ? JSON.stringify({ token }) : undefined));
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      const ui = <App />;
      render(ui, {
        preloadedState: {
          ...preloadedState,
          users: {
            ...preloadedState.users,
            currentSession: getSessionInfo(),
            settingsInitialized: false,
            globalSettings: {
              ...preloadedState.users.globalSettings,
              offlineThreshold: { interval: 15, intervalUnit: 'minutes' }
            }
          }
        }
      });
      await act(async () => {
        jest.advanceTimersByTime(TIMEOUTS.fiveSeconds);
        jest.runAllTicks();
      });
      await waitFor(() => expect(screen.queryByText(/granular device connectivity/i)).toBeInTheDocument(), { timeout: TIMEOUTS.fiveSeconds * 2 });
      const post = jest.spyOn(GeneralApi, 'post');
      await user.click(screen.getByRole('button', { name: /close/i }));
      await act(async () => {
        jest.runOnlyPendingTimers();
        jest.runAllTicks();
      });
      expect(post).toHaveBeenCalledWith(
        '/api/management/v1/useradm/settings',
        {
          id_attribute: { attribute: 'mac', scope: 'identity' },
          previousFilters: [],
          offlineThreshold: { interval: 1, intervalUnit: 'days' }
        },
        { headers: {} }
      );
    },
    15 * TIMEOUTS.oneSecond
  );
});
