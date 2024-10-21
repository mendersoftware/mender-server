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
import { getGlobalSettings, saveGlobalSettings, setOfflineThreshold } from '@northern.tech/store/thunks';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { defaultState } from '../../../../tests/mockData';
import { actions as appActions } from '../appSlice';
import { actions as deviceActions } from '../devicesSlice';
import { actions as userActions } from '../usersSlice';
import * as DeploymentConstants from './constants';
import {
  abortDeployment,
  createDeployment,
  getDeploymentDevices,
  getDeploymentsByStatus,
  getDeploymentsConfig,
  getDeviceDeployments,
  getDeviceLog,
  getSingleDeployment,
  resetDeviceDeployments,
  saveDeltaDeploymentsConfig,
  setDeploymentsState,
  updateDeploymentControlMap
} from './thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

const createdDeployment = {
  ...defaultState.deployments.byId.d1,
  id: 'created-123'
};
const deploymentsConfig = {
  binaryDelta: {
    compressionLevel: 6,
    disableChecksum: false,
    disableDecompression: false,
    duplicatesWindow: 0,
    inputWindow: 0,
    instructionBuffer: 0,
    sourceWindow: 0,
    timeout: 0
  },
  binaryDeltaLimits: {
    duplicatesWindow: DeploymentConstants.limitDefault,
    inputWindow: DeploymentConstants.limitDefault,
    instructionBuffer: DeploymentConstants.limitDefault,
    sourceWindow: DeploymentConstants.limitDefault,
    timeout: { default: 60, max: 3600, min: 60 }
  },
  hasDelta: true
};

const defaultResponseActions = {
  creation: {
    type: actions.createdDeployment.type,
    isImportant: true,
    payload: { id: createdDeployment.id, devices: [{ id: Object.keys(defaultState.devices.byId)[0], status: 'pending' }], statistics: { status: {} } }
  },
  devices: {
    type: actions.receivedDeploymentDevices.type,
    payload: {
      id: defaultState.deployments.byId.d1.id,
      devices: defaultState.deployments.byId.d1.devices,
      selectedDeviceIds: [defaultState.deployments.byId.d1.devices.a1.id],
      totalDeviceCount: 1
    }
  },
  log: {
    type: actions.receivedDeploymentDeviceLog.type,
    payload: {
      id: defaultState.deployments.byId.d1.id,
      deviceId: defaultState.deployments.byId.d1.devices.a1.id,
      log: 'test'
    }
  },
  snackbar: { type: appActions.setSnackbar.type, payload: 'Deployment created successfully' },
  receive: { type: actions.receivedDeployment.type, payload: createdDeployment },
  receiveMultiple: { type: actions.receivedDeployments.type, payload: {} },
  receiveInprogress: { type: actions.receivedDeploymentsForStatus.type, payload: { deploymentIds: [], status: 'inprogress', total: 0 } },
  remove: { type: actions.removedDeployment.type, payload: defaultState.deployments.byId.d1.id },
  selectMultiple: {
    type: actions.selectDeploymentsForStatus.type,
    payload: { deploymentIds: Object.keys(defaultState.deployments.byId), status: 'inprogress' }
  },
  setOfflineThreshold: { type: appActions.setOfflineThreshold.type, payload: '2019-01-12T13:00:00.900Z' }
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { id_attribute, ...retrievedSettings } = defaultState.users.globalSettings;

const assertionFunction =
  storeActions =>
  ({ type, isImportant, payload }, index) => {
    expect(storeActions[index].type).toEqual(type);
    if (isImportant) {
      expect(storeActions[index].payload).toEqual(payload);
    }
  };

/* eslint-disable sonarjs/no-identical-functions */
describe('deployment actions', () => {
  it('should allow aborting deployments', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: abortDeployment.pending.type },
      defaultResponseActions.receiveMultiple,
      defaultResponseActions.receiveInprogress,
      defaultResponseActions.remove,
      { ...defaultResponseActions.snackbar, payload: 'The deployment was successfully aborted' },
      { type: abortDeployment.fulfilled.type }
    ];
    return store
      .dispatch(abortDeployment(defaultState.deployments.byId.d1.id))
      .unwrap()
      .then(() => {
        const storeActions = store.getActions();
        expect(storeActions.length).toEqual(expectedActions.length);
        expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
      });
  });
  it(`should reject aborting deployments that don't exist`, () => {
    const store = mockStore({ ...defaultState });
    const abortedDeployment = store.dispatch(abortDeployment(`${defaultState.deployments.byId.d1.id}-invalid`)).unwrap();
    expect(typeof abortedDeployment === Promise);
    expect(abortedDeployment).rejects.toBeTruthy();
  });
  it('should allow creating deployments without filter or group', async () => {
    const store = mockStore({
      ...defaultState,
      deployments: {
        ...defaultState.deployments,
        byStatus: {
          ...defaultState.deployments.byStatus,
          finished: { ...defaultState.deployments.byStatus.finished, total: 0 },
          inprogress: { ...defaultState.deployments.byStatus.inprogress, total: 0 },
          pending: { ...defaultState.deployments.byStatus.pending, total: 0 },
          scheduled: { ...defaultState.deployments.byStatus.scheduled, total: 0 }
        }
      }
    });
    const expectedActions = [
      { type: createDeployment.pending.type },
      defaultResponseActions.creation,
      { type: getSingleDeployment.pending.type },
      defaultResponseActions.snackbar,
      { type: saveGlobalSettings.pending.type },
      { type: getGlobalSettings.pending.type },
      defaultResponseActions.receive,
      { type: getSingleDeployment.fulfilled.type },
      { type: userActions.setGlobalSettings.type },
      { type: setOfflineThreshold.pending.type },
      defaultResponseActions.setOfflineThreshold,
      { type: setOfflineThreshold.fulfilled.type },
      { type: getGlobalSettings.fulfilled.type },
      { type: userActions.setGlobalSettings.type },
      { type: saveGlobalSettings.fulfilled.type },
      { type: createDeployment.fulfilled.type }
    ];
    return store.dispatch(createDeployment({ newDeployment: { devices: [Object.keys(defaultState.devices.byId)[0]] } })).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map(assertionFunction(storeActions));
    });
  });
  it('should allow creating deployments with a filter', async () => {
    const store = mockStore({ ...defaultState });
    const filter_id = '1234';
    const expectedActions = [
      { type: createDeployment.pending.type },
      { ...defaultResponseActions.creation, payload: { ...defaultResponseActions.creation.payload, devices: [], filter_id, statistics: { status: {} } } },
      { type: getSingleDeployment.pending.type },
      defaultResponseActions.snackbar,
      { type: saveGlobalSettings.pending.type },
      { type: getGlobalSettings.pending.type },
      defaultResponseActions.receive,
      { type: getSingleDeployment.fulfilled.type },
      { type: userActions.setGlobalSettings.type },
      { type: setOfflineThreshold.pending.type },
      defaultResponseActions.setOfflineThreshold,
      { type: setOfflineThreshold.fulfilled.type },
      { type: getGlobalSettings.fulfilled.type },
      { type: userActions.setGlobalSettings.type },
      { type: saveGlobalSettings.fulfilled.type },
      { type: createDeployment.fulfilled.type }
    ];
    return store.dispatch(createDeployment({ newDeployment: { filter_id } })).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map(assertionFunction(storeActions));
    });
  });
  it('should allow creating deployments with a group', async () => {
    const store = mockStore({ ...defaultState });
    const group = Object.keys(defaultState.devices.groups.byId)[0];
    const expectedActions = [
      { type: createDeployment.pending.type },
      { ...defaultResponseActions.creation, payload: { ...defaultResponseActions.creation.payload, devices: [], group, statistics: { status: {} } } },
      { type: getSingleDeployment.pending.type },
      defaultResponseActions.snackbar,
      { type: saveGlobalSettings.pending.type },
      { type: getGlobalSettings.pending.type },
      defaultResponseActions.receive,
      { type: getSingleDeployment.fulfilled.type },
      { type: userActions.setGlobalSettings.type },
      { type: setOfflineThreshold.pending.type },
      defaultResponseActions.setOfflineThreshold,
      { type: setOfflineThreshold.fulfilled.type },
      { type: getGlobalSettings.fulfilled.type },
      { type: userActions.setGlobalSettings.type },
      { type: saveGlobalSettings.fulfilled.type },
      { type: createDeployment.fulfilled.type }
    ];
    return store.dispatch(createDeployment({ newDeployment: { group } })).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map(assertionFunction(storeActions));
    });
  });
  it('should allow deployments retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeploymentsByStatus.pending.type },
      { ...defaultResponseActions.receiveMultiple, payload: defaultState.deployments.byId },
      {
        ...defaultResponseActions.receiveInprogress,
        payload: {
          deploymentIds: Object.keys(defaultState.deployments.byId),
          total: defaultState.deployments.byStatus.inprogress.total
        }
      },
      defaultResponseActions.selectMultiple,
      { type: getDeploymentsByStatus.fulfilled.type }
    ];
    return store
      .dispatch(getDeploymentsByStatus({ status: 'inprogress', group: Object.keys(defaultState.devices.groups.byId)[0], type: 'configuration' }))
      .then(() => {
        const storeActions = store.getActions();
        expect(storeActions.length).toEqual(expectedActions.length);
        expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
      });
  });
  it('should allow deployment device log retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: getDeviceLog.pending.type }, defaultResponseActions.log, { type: getDeviceLog.fulfilled.type }];
    return store
      .dispatch(getDeviceLog({ deploymentId: Object.keys(defaultState.deployments.byId)[0], deviceId: defaultState.deployments.byId.d1.devices.a1.id }))
      .then(() => {
        const storeActions = store.getActions();
        expect(storeActions.length).toEqual(expectedActions.length);
        expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
      });
  });
  it('should allow deployment device list retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: getDeploymentDevices.pending.type }, defaultResponseActions.devices, { type: getDeploymentDevices.fulfilled.type }];
    return store.dispatch(getDeploymentDevices({ id: Object.keys(defaultState.deployments.byId)[0] })).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow device deployment history retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceDeployments.pending.type },
      {
        type: deviceActions.receivedDevice.type,
        payload: {
          id: defaultState.devices.byId.a1.id,
          deploymentsCount: 34,
          deviceDeployments: [
            {
              id: defaultState.deployments.byId.d1.id,
              release: defaultState.deployments.byId.d1.artifact_name,
              created: '2019-01-01T12:35:00.000Z',
              finished: '2019-01-01T12:40:00.000Z',
              status: 'noartifact',
              route: DeploymentConstants.DEPLOYMENT_ROUTES.active.key
            }
          ]
        }
      },
      { type: getDeviceDeployments.fulfilled.type }
    ];
    await store.dispatch(getDeviceDeployments({ deviceId: defaultState.devices.byId.a1.id }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow device deployment history deletion', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: resetDeviceDeployments.pending.type },
      { type: getDeviceDeployments.pending.type },
      {
        type: deviceActions.receivedDevice.type,
        payload: {
          id: defaultState.devices.byId.a1.id,
          deploymentsCount: 34,
          deviceDeployments: [
            {
              id: defaultState.deployments.byId.d1.id,
              release: defaultState.deployments.byId.d1.artifact_name,
              created: '2019-01-01T12:35:00.000Z',
              finished: '2019-01-01T12:40:00.000Z',
              status: 'noartifact',
              route: DeploymentConstants.DEPLOYMENT_ROUTES.active.key
            }
          ]
        }
      },
      { type: getDeviceDeployments.fulfilled.type },
      { type: resetDeviceDeployments.fulfilled.type }
    ];
    await store.dispatch(resetDeviceDeployments(defaultState.devices.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow updating a deployment to continue the execution', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: updateDeploymentControlMap.pending.type },
      { type: getSingleDeployment.pending.type },
      defaultResponseActions.receive,
      { type: getSingleDeployment.fulfilled.type },
      { type: updateDeploymentControlMap.fulfilled.type }
    ];
    return store.dispatch(updateDeploymentControlMap({ deploymentId: createdDeployment.id, updateControlMap: { something: 'continue' } })).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow deployment state tracking', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(
      setDeploymentsState({
        general: { showCreationDialog: true },
        [DeploymentConstants.DEPLOYMENT_STATES.finished]: { something: 'new' },
        selectedId: createdDeployment.id
      })
    );
    const expectedActions = [
      { type: setDeploymentsState.pending.type },
      {
        type: actions.setDeploymentsState.type,
        payload: {
          ...defaultState.deployments.selectionState,
          finished: {
            ...defaultState.deployments.selectionState.finished,
            something: 'new'
          },
          general: {
            ...defaultState.deployments.selectionState.general,
            showCreationDialog: true
          },
          selectedId: createdDeployment.id
        }
      },
      { type: getSingleDeployment.pending.type },
      defaultResponseActions.receive,
      { type: getSingleDeployment.fulfilled.type },
      { type: setDeploymentsState.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should allow retrieving config for deployments', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeploymentsConfig.pending.type },
      { type: actions.setDeploymentsConfig.type, payload: deploymentsConfig },
      { type: getDeploymentsConfig.fulfilled.type }
    ];
    return store.dispatch(getDeploymentsConfig()).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow storing delta deployments settings', async () => {
    const store = mockStore({ ...defaultState });
    const changedConfig = {
      timeout: 100,
      duplicatesWindow: 734,
      compressionLevel: 5,
      disableChecksum: true,
      disableDecompression: false,
      inputWindow: 1253,
      instructionBuffer: 123,
      sourceWindow: 13
    };
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { hasDelta, ...expectedConfig } = deploymentsConfig;
    const expectedActions = [
      { type: saveDeltaDeploymentsConfig.pending.type },
      { type: actions.setDeploymentsConfig.type, payload: { ...expectedConfig, binaryDelta: { ...expectedConfig.binaryDelta, ...changedConfig } } },
      { ...defaultResponseActions.setSnackbar, payload: 'Settings saved successfully' },
      { type: saveDeltaDeploymentsConfig.fulfilled.type }
    ];
    return store.dispatch(saveDeltaDeploymentsConfig(changedConfig)).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
});
