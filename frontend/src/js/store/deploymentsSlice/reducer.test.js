/*eslint import/namespace: ['error', { allowComputed: true }]*/
import { defaultState } from '../../../tests/mockData';
import * as DeploymentConstants from '../constants/deploymentConstants';
import reducer, { initialState } from './deploymentReducer';

const {
  RECEIVE_DEPLOYMENT,
  RECEIVE_DEPLOYMENTS,
  RECEIVE_DEPLOYMENT_DEVICE_LOG,
  RECEIVE_DEPLOYMENT_DEVICES,
  DEPLOYMENT_STATES,
  SET_DEPLOYMENTS_CONFIG,
  SET_DEPLOYMENTS_STATE,
  REMOVE_DEPLOYMENT,
  CREATE_DEPLOYMENT
} = DeploymentConstants;

describe('deployment reducer', () => {
  it('should return the initial state', async () => {
    expect(reducer(undefined, {})).toEqual(initialState);
  });

  it('should handle RECEIVE_DEPLOYMENT', async () => {
    expect(reducer(undefined, { type: RECEIVE_DEPLOYMENT, deployment: defaultState.deployments.byId.d1 }).byId.d1).toEqual(defaultState.deployments.byId.d1);
    expect(reducer(initialState, { type: RECEIVE_DEPLOYMENT, deployment: defaultState.deployments.byId.d1 }).byId.d1).toEqual(defaultState.deployments.byId.d1);
  });
  it('should handle RECEIVE_DEPLOYMENTS', async () => {
    const { statistics } = defaultState.deployments.byId.d1;
    expect(reducer(undefined, { type: RECEIVE_DEPLOYMENTS, deployments: { plain: 'passing' } }).byId.plain).toBeTruthy();
    expect(
      reducer(initialState, { type: RECEIVE_DEPLOYMENTS, deployments: { [defaultState.deployments.byId.d1.id]: { statistics } } }).byId.d1.statistics
    ).toBeTruthy();
  });
  it('should handle RECEIVE_DEPLOYMENT_DEVICE_LOG', async () => {
    const { devices } = defaultState.deployments.byId.d1;
    expect(reducer(undefined, { type: RECEIVE_DEPLOYMENT_DEVICE_LOG, deployment: defaultState.deployments.byId.d1 }).byId.d1.devices.a1.id).toEqual(
      devices.a1.id
    );
    expect(reducer(initialState, { type: RECEIVE_DEPLOYMENT_DEVICE_LOG, deployment: defaultState.deployments.byId.d1 }).byId.d1.devices.a1.id).toEqual(
      devices.a1.id
    );
  });
  it('should handle RECEIVE_DEPLOYMENT_DEVICES', async () => {
    const { devices, id } = defaultState.deployments.byId.d1;
    expect(
      reducer(undefined, {
        type: RECEIVE_DEPLOYMENT_DEVICES,
        deploymentId: id,
        devices,
        selectedDeviceIds: [devices.a1.id],
        totalDeviceCount: 500
      }).byId.d1.totalDeviceCount
    ).toEqual(500);
    expect(
      reducer(defaultState.deployments, {
        type: RECEIVE_DEPLOYMENT_DEVICES,
        deploymentId: id,
        devices,
        selectedDeviceIds: [devices.a1.id],
        totalDeviceCount: 500
      }).byId.d1.statistics
    ).toEqual(defaultState.deployments.byId.d1.statistics);
  });
  it('should handle RECEIVE_<deploymentstatus>_DEPLOYMENTS', async () => {
    Object.values(DEPLOYMENT_STATES).forEach(status => {
      expect(
        reducer(undefined, { type: DeploymentConstants[`RECEIVE_${status.toUpperCase()}_DEPLOYMENTS`], deploymentIds: ['a1'], total: 1, status }).byStatus[
          status
        ]
      ).toEqual({ deploymentIds: ['a1'], total: 1 });
      expect(
        reducer(initialState, { type: DeploymentConstants[`RECEIVE_${status.toUpperCase()}_DEPLOYMENTS`], deploymentIds: ['a1'], total: 1, status }).byStatus[
          status
        ]
      ).toEqual({ deploymentIds: ['a1'], total: 1 });
    });
  });
  it('should handle SELECT_<deploymentstatus>_DEPLOYMENTS', async () => {
    Object.values(DEPLOYMENT_STATES).forEach(status => {
      expect(
        reducer(undefined, { type: DeploymentConstants[`SELECT_${status.toUpperCase()}_DEPLOYMENTS`], deploymentIds: ['a1'], status }).selectionState[status]
          .selection
      ).toEqual(['a1']);
      expect(
        reducer(initialState, { type: DeploymentConstants[`SELECT_${status.toUpperCase()}_DEPLOYMENTS`], deploymentIds: ['a1'], status }).selectionState[status]
          .selection
      ).toEqual(['a1']);
    });
  });
  it('should handle SET_DEPLOYMENTS_STATE', async () => {
    const newState = { something: 'new' };
    expect(reducer(undefined, { type: SET_DEPLOYMENTS_STATE, state: newState }).selectionState).toEqual(newState);
    expect(reducer(initialState, { type: SET_DEPLOYMENTS_STATE, state: newState }).selectionState).toEqual(newState);
  });
  it('should handle REMOVE_DEPLOYMENT', async () => {
    let state = reducer(undefined, { type: RECEIVE_DEPLOYMENT, deployment: defaultState.deployments.byId.d1 });
    expect(reducer(state, { type: REMOVE_DEPLOYMENT, deploymentId: defaultState.deployments.byId.d1.id }).byId).toEqual({});
    expect(reducer(initialState, { type: REMOVE_DEPLOYMENT, deploymentId: 'a1' }).byId).toEqual({});
  });
  it('should handle CREATE_DEPLOYMENT', async () => {
    expect(reducer(undefined, { type: CREATE_DEPLOYMENT, deployment: { name: 'test' }, deploymentId: 'test' }).byId.test.devices).toEqual({});
    expect(reducer(initialState, { type: CREATE_DEPLOYMENT, deployment: { name: 'test' }, deploymentId: 'a1' }).byStatus.pending.deploymentIds).toContain('a1');
  });
  it('should handle SET_DEPLOYMENTS_CONFIG', async () => {
    expect(reducer(undefined, { type: SET_DEPLOYMENTS_CONFIG, config: { name: 'test' } }).config).toEqual({ name: 'test' });
    expect(reducer(initialState, { type: SET_DEPLOYMENTS_CONFIG, config: { name: 'test' } }).config).toEqual({ name: 'test' });
  });
});
