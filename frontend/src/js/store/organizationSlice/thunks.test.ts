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
// @ts-nocheck
import { EXTERNAL_PROVIDER } from '@northern.tech/store/commonConstants';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { defaultState, webhookEvents } from '../../../../tests/mockData';
import { actions as appActions } from '../appSlice';
import { locations } from '../appSlice/constants';
import { getSessionInfo } from '../auth';
import { TIMEOUTS } from '../commonConstants';
import { SSO_TYPES } from './constants';
import {
  cancelRequest,
  cancelUpgrade,
  changeIntegration,
  changeSsoConfig,
  completeUpgrade,
  confirmCardUpdate,
  createIntegration,
  createOrganizationTrial,
  deleteIntegration,
  deleteSsoConfig,
  downloadLicenseReport,
  getAuditLogs,
  getAuditLogsCsvLink,
  getCurrentCard,
  getIntegrations,
  getSsoConfigById,
  getSsoConfigs,
  getTargetLocation,
  getUserOrganization,
  getWebhookEvents,
  requestPlanChange,
  sendSupportMessage,
  setAuditlogsState,
  startCardUpdate,
  startUpgrade,
  storeSsoConfig,
  tenantDataDivergedMessage
} from './thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

const expectedDeviceProviders = [
  { id: 1, provider: EXTERNAL_PROVIDER['iot-hub'].provider, something: 'something', connection_string: 'something_else' },
  { id: 2, provider: EXTERNAL_PROVIDER['iot-core'].provider, something: 'new' }
];

const expectedSsoConfigs = [
  { id: '1', issuer: 'https://samltest.id/saml/idp', valid_until: '2038-08-24T21:14:09Z' },
  { id: '2', issuer: 'https://samltest2.id/saml/idp', valid_until: '2030-10-24T21:14:09Z' }
];

const oldHostname = window.location.hostname;

/* eslint-disable sonarjs/no-identical-functions */
describe('organization actions', () => {
  it('should handle different error message formats', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: cancelRequest.pending.type },
      { type: appActions.setSnackbar.type, payload: { message: 'Deactivation request was sent successfully', autoHideDuration: TIMEOUTS.fiveSeconds } },
      { type: cancelRequest.fulfilled.type }
    ];
    await store.dispatch(cancelRequest(defaultState.organization.organization.id, 'testReason')).then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });

  it('should point to proper target locations from different circumstances', () => {
    const expectations = [
      { hostname: locations.us.location, location: locations.us.key, result: locations.us.location },
      { hostname: locations.us.location, location: locations.eu.key, result: locations.eu.location },
      { hostname: locations.eu.location, location: locations.us.key, result: locations.us.location },
      { hostname: locations.eu.location, location: locations.eu.key, result: locations.eu.location },
      { hostname: `staging.${locations.us.location}`, location: locations.us.key, result: `staging.${locations.us.location}` },
      { hostname: `staging.${locations.us.location}`, location: locations.eu.key, result: `staging.${locations.eu.location}` },
      { hostname: `testing.staging.${locations.us.location}`, location: locations.us.key, result: `testing.staging.${locations.us.location}` },
      { hostname: `testing.staging.${locations.us.location}`, location: locations.eu.key, result: `testing.staging.${locations.eu.location}` },
      { hostname: 'docker.mender.io', location: locations.us.key, result: '' },
      { hostname: 'docker.mender.io', location: locations.eu.key, result: '' },
      { hostname: 'localhost', location: locations.us.key, result: '' },
      { hostname: 'localhost', location: locations.eu.key, result: '' }
    ];

    expectations.map(({ hostname, location, result }) => {
      window.location = { ...window.location, hostname };
      let targetLocation = getTargetLocation(location);
      expect(targetLocation).toBe(result ? `https://${result}` : result);
    });

    window.location = { ...window.location, hostname: oldHostname };
  });

  it('should handle trial creation', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [{ type: appActions.setFirstLoginAfterSignup.type, payload: true }];
    const result = store.dispatch(
      createOrganizationTrial({
        'g-recaptcha-response': 'test',
        email: 'test@test.com',
        location: 'us',
        marketing: true,
        organization: 'test',
        plan: 'os',
        tos: true
      })
    );
    jest.advanceTimersByTime(6000);
    result.then(token => {
      expect(token).toBeTruthy();
      expect(store.getActions()).toHaveLength(expectedActions.length);
    });
  });

  it('should handle credit card details retrieval', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getCurrentCard.pending.type },
      { type: actions.receiveCurrentCard.type, payload: defaultState.organization.card },
      { type: getCurrentCard.fulfilled.type }
    ];
    await store.dispatch(getCurrentCard()).then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });

  it('should handle organization retrieval', async () => {
    const store = mockStore({ ...defaultState, users: { ...defaultState.users, currentSession: getSessionInfo() } });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getUserOrganization.pending.type },
      { type: actions.setOrganization.type, payload: defaultState.organization.organization },
      { type: appActions.setAnnouncement.type, payload: tenantDataDivergedMessage },
      { type: getUserOrganization.fulfilled.type }
    ];
    await store.dispatch(getUserOrganization()).then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });

  it('should handle support request sending', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: sendSupportMessage.pending.type },
      { type: appActions.setSnackbar.type, payload: { message: 'Your request was sent successfully', autoHideDuration: TIMEOUTS.fiveSeconds } },
      { type: sendSupportMessage.fulfilled.type }
    ];
    await store.dispatch(sendSupportMessage({ body: 'test', subject: 'testsubject' })).then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });

  it('should handle schema based support request sending', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: requestPlanChange.pending.type },
      { type: appActions.setSnackbar.type, payload: { message: 'Your request was sent successfully', autoHideDuration: TIMEOUTS.fiveSeconds } },
      { type: requestPlanChange.fulfilled.type }
    ];
    await store
      .dispatch(
        requestPlanChange({
          tenantId: defaultState.organization.organization.id,
          content: {
            current_plan: 'Basic',
            requested_plan: 'Enterprise',
            current_addons: 'something,extra',
            requested_addons: 'something,extra,special',
            user_message: 'more please'
          }
        })
      )
      .then(() => {
        const storeActions = store.getActions();
        expect(storeActions).toHaveLength(expectedActions.length);
        expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
      });
  });

  it('should handle license report downloads', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [{ type: downloadLicenseReport.pending.type }, { type: downloadLicenseReport.fulfilled.type }];
    const result = await store.dispatch(downloadLicenseReport()).unwrap();
    const storeActions = store.getActions();
    expect(storeActions).toHaveLength(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    expect(result).toEqual('test,report');
  });

  it('should handle account upgrade init', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: startUpgrade.pending.type }, { type: startUpgrade.fulfilled.type }];
    const secret = await store.dispatch(startUpgrade(defaultState.organization.organization.id)).unwrap();
    const storeActions = store.getActions();
    expect(storeActions).toHaveLength(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    expect(secret).toEqual('testSecret');
  });

  it('should handle account upgrade cancelling', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: cancelUpgrade.pending.type }, { type: cancelUpgrade.fulfilled.type }];
    await store.dispatch(cancelUpgrade(defaultState.organization.organization.id));
    const storeActions = store.getActions();
    expect(storeActions).toHaveLength(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should handle account upgrade completion', async () => {
    const store = mockStore({ ...defaultState, users: { ...defaultState.users, currentSession: getSessionInfo() } });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: completeUpgrade.pending.type },
      { type: getUserOrganization.pending.type },
      { type: actions.setOrganization.type, payload: defaultState.organization.organization },
      { type: appActions.setAnnouncement.type, payload: tenantDataDivergedMessage },
      { type: getUserOrganization.fulfilled.type },
      { type: completeUpgrade.fulfilled.type }
    ];
    await store.dispatch(completeUpgrade({ tenantId: defaultState.organization.organization.id, plan: 'enterprise' }));
    const storeActions = store.getActions();
    expect(storeActions).toHaveLength(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should handle confirm card update initialization', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: startCardUpdate.pending.type },
      { type: actions.receiveSetupIntent.type, payload: 'testIntent' },
      { type: startCardUpdate.fulfilled.type }
    ];
    const secret = await store.dispatch(startCardUpdate()).unwrap();
    const storeActions = store.getActions();
    expect(secret).toEqual('testSecret');
    expect(storeActions).toHaveLength(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should handle confirm card update confirmation', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: confirmCardUpdate.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Payment card was updated successfully' },
      { type: actions.receiveSetupIntent.type, payload: null },
      { type: confirmCardUpdate.fulfilled.type }
    ];
    const request = store.dispatch(confirmCardUpdate());
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });

  it('should handle auditlog retrieval', async () => {
    const store = mockStore({
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasAuditlogs: true,
          isEnterprise: true
        }
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getAuditLogs.pending.type },
      {
        type: actions.receiveAuditLogs.type,
        payload: { events: defaultState.organization.auditlog.events, total: defaultState.organization.auditlog.selectionState.total }
      },
      { type: getAuditLogs.fulfilled.type }
    ];
    const request = store.dispatch(getAuditLogs({ page: 1, perPage: 20 }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow auditlog state tracking', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(setAuditlogsState({ page: 1, sort: { direction: 'something' } }));
    const expectedActions = [
      { type: setAuditlogsState.pending.type },
      { type: getAuditLogs.pending.type },
      {
        type: actions.setAuditLogState.type,
        payload: { ...defaultState.organization.auditlog.selectionState, isLoading: true, sort: { direction: 'something' } }
      },
      { type: getAuditLogs.fulfilled.type },
      { type: actions.setAuditLogState.type, payload: { isLoading: false } },
      { type: setAuditlogsState.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should handle csv information download', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [{ type: getAuditLogsCsvLink.pending.type }, { type: getAuditLogsCsvLink.fulfilled.type }];
    const request = store.dispatch(getAuditLogsCsvLink()).unwrap();
    expect(request).resolves.toBeTruthy();
    await request.then(link => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
      expect(link).toEqual('http://localhost/api/management/v1/auditlogs/logs/export?limit=20000&sort=desc');
    });
  });
  it('should allow initializing external device providers', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        externalDeviceIntegrations: [
          { id: 1, something: 'something' },
          { id: 2, provider: 'aws', something: 'new' }
        ]
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: createIntegration.pending.type },
      { type: appActions.setSnackbar.type, payload: 'The integration was set up successfully' },
      { type: getIntegrations.pending.type },
      { type: actions.receiveExternalDeviceIntegrations.type, payload: expectedDeviceProviders },
      { type: getIntegrations.fulfilled.type },
      { type: createIntegration.fulfilled.type }
    ];
    const request = store.dispatch(createIntegration({ connection_string: 'testString', provider: 'iot-hub' }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow configuring external device providers', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        externalDeviceIntegrations: [
          { id: 1, something: 'something' },
          { id: 2, provider: 'iot-core', something: 'new' }
        ]
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: changeIntegration.pending.type },
      { type: appActions.setSnackbar.type, payload: 'The integration was updated successfully' },
      { type: getIntegrations.pending.type },
      { type: actions.receiveExternalDeviceIntegrations.type, payload: expectedDeviceProviders },
      { type: getIntegrations.fulfilled.type },
      { type: changeIntegration.fulfilled.type }
    ];
    const request = store.dispatch(changeIntegration({ connection_string: 'testString2', id: 1, provider: 'iot-hub' }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow retrieving external device providers', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        externalDeviceIntegrations: [
          { id: 1, something: 'something' },
          { id: 2, provider: 'iot-core', something: 'new' }
        ]
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getIntegrations.pending.type },
      { type: actions.receiveExternalDeviceIntegrations.type, payload: expectedDeviceProviders },
      { type: getIntegrations.fulfilled.type }
    ];
    const request = store.dispatch(getIntegrations());
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow deleting external device provider configurations', async () => {
    const store = mockStore({ ...defaultState, externalDeviceIntegrations: [{ id: 1, something: 'something' }] });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: deleteIntegration.pending.type },
      { type: appActions.setSnackbar.type, payload: 'The integration was removed successfully' },
      { type: actions.receiveExternalDeviceIntegrations.type, payload: [] },
      { type: deleteIntegration.fulfilled.type }
    ];
    const request = store.dispatch(deleteIntegration({ id: 1 }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow retrieving webhook events', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        webhooks: {
          ...defaultState.organization.webhooks,
          events: [
            { id: 1, something: 'something' },
            { id: 2, provider: 'aws', something: 'new' }
          ]
        }
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getWebhookEvents.pending.type },
      { type: actions.receiveWebhookEvents.type, payload: { value: webhookEvents, total: 2 } },
      { type: getWebhookEvents.fulfilled.type }
    ];
    const request = store.dispatch(getWebhookEvents());
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should auto check for more webhook events', async () => {
    const existingEvents = [
      { id: 1, something: 'something' },
      { id: 2, provider: 'aws', something: 'new' }
    ];
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        webhooks: {
          ...defaultState.organization.webhooks,
          events: existingEvents,
          eventTotal: 2
        }
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const defaultEvent = webhookEvents[0];
    const expectedActions = [
      { type: getWebhookEvents.pending.type },
      { type: actions.receiveWebhookEvents.type, payload: { value: [defaultEvent], total: 1 } },
      { type: getWebhookEvents.pending.type },
      { type: actions.receiveWebhookEvents.type, payload: { value: existingEvents, total: 2 } },
      { type: getWebhookEvents.fulfilled.type },
      { type: getWebhookEvents.fulfilled.type }
    ];
    const request = store.dispatch(getWebhookEvents({ page: 1, perPage: 1 }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow configuring external identity providers', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        ssoConfigs: [{ id: 1, something: 'something' }]
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: storeSsoConfig.pending.type },
      { type: appActions.setSnackbar.type, payload: 'The SSO configuration was stored successfully' },
      { type: getSsoConfigs.pending.type },
      { type: getSsoConfigById.pending.type },
      { type: getSsoConfigById.pending.type },
      { type: getSsoConfigById.fulfilled.type },
      { type: getSsoConfigById.fulfilled.type },
      { type: actions.receiveSsoConfigs.type },
      { type: getSsoConfigs.fulfilled.type },
      { type: storeSsoConfig.fulfilled.type }
    ];
    const request = store.dispatch(
      storeSsoConfig({ config: { connection_string: 'testString', provider: 'iot-hub' }, contentType: SSO_TYPES.oidc.contentType })
    );
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow updating external identity providers', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        ssoConfigs: [
          { id: 1, something: 'something' },
          { id: 2, provider: 'aws', something: 'new' }
        ]
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: changeSsoConfig.pending.type },
      { type: appActions.setSnackbar.type, payload: 'The SSO configuration was updated successfully' },
      { type: getSsoConfigs.pending.type },
      { type: getSsoConfigById.pending.type },
      { type: getSsoConfigById.pending.type },
      { type: getSsoConfigById.fulfilled.type },
      { type: getSsoConfigById.fulfilled.type },
      { type: actions.receiveSsoConfigs.type },
      { type: getSsoConfigs.fulfilled.type },
      { type: changeSsoConfig.fulfilled.type }
    ];
    const request = store.dispatch(
      changeSsoConfig({ config: { connection_string: 'testString2', id: 1, provider: 'iot-hub' }, contentType: SSO_TYPES.oidc.contentType })
    );
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow retrieving external identity providers', async () => {
    const store = mockStore({
      ...defaultState,
      organization: {
        ...defaultState.organization,
        ssoConfigs: [
          { id: 1, something: 'something' },
          { id: 2, provider: 'aws', something: 'new' }
        ]
      }
    });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getSsoConfigs.pending.type },
      { type: getSsoConfigById.pending.type },
      { type: getSsoConfigById.pending.type },
      { type: getSsoConfigById.fulfilled.type },
      { type: getSsoConfigById.fulfilled.type },
      { type: actions.receiveSsoConfigs.type, payload: expectedSsoConfigs },
      { type: getSsoConfigs.fulfilled.type }
    ];
    const request = store.dispatch(getSsoConfigs());
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow deleting external identity providers', async () => {
    const store = mockStore({ ...defaultState, organization: { ...defaultState.organization, ssoConfigs: [...expectedSsoConfigs] } });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: deleteSsoConfig.pending.type },
      { type: appActions.setSnackbar.type, payload: 'The SSO configuration was removed successfully' },
      { type: actions.receiveSsoConfigs.type, payload: [expectedSsoConfigs[1]] },
      { type: deleteSsoConfig.fulfilled.type }
    ];
    const request = store.dispatch(deleteSsoConfig({ id: '1' }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
});
