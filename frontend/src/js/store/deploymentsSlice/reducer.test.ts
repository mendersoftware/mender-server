// Copyright 2024 Northern.tech AS
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

/*eslint import/namespace: ['error', { allowComputed: true }]*/
import reducer, { actions, initialState } from '.';
import { defaultState } from '../../../../tests/mockData';

describe('deployment reducer', () => {
  it('should handle CREATE_DEPLOYMENT', async () => {
    expect(reducer(undefined, { type: actions.createdDeployment, payload: { name: 'test', id: 'test' } }).byId.test.devices).toEqual({});
    expect(reducer(initialState, { type: actions.createdDeployment, payload: { name: 'test', id: 'a1' } }).byStatus.pending.deploymentIds).toContain('a1');
  });
  it('should handle RECEIVE_DEPLOYMENT_DEVICE_LOG', async () => {
    expect(
      reducer(undefined, {
        type: actions.receivedDeploymentDeviceLog,
        payload: { id: defaultState.deployments.byId.d1.id, deviceId: defaultState.deployments.byId.d1.devices.a1.id, log: 'foo' }
      }).byId.d1.devices.a1.log
    ).toEqual('foo');
    expect(
      reducer(initialState, {
        type: actions.receivedDeploymentDeviceLog,
        payload: { id: defaultState.deployments.byId.d1.id, deviceId: defaultState.deployments.byId.d1.devices.a1.id, log: 'bar' }
      }).byId.d1.devices.a1.log
    ).toEqual('bar');
  });
  it('should handle RECEIVE_DEPLOYMENT_DEVICES', async () => {
    const { devices, id } = defaultState.deployments.byId.d1;
    expect(
      reducer(undefined, {
        type: actions.receivedDeploymentDevices,
        payload: { id, devices, selectedDeviceIds: [devices.a1.id], totalDeviceCount: 500 }
      }).byId.d1.totalDeviceCount
    ).toEqual(500);
    expect(
      reducer(defaultState.deployments, {
        type: actions.receivedDeploymentDevices,
        payload: { id, devices, selectedDeviceIds: [devices.a1.id], totalDeviceCount: 500 }
      }).byId.d1.statistics
    ).toEqual(defaultState.deployments.byId.d1.statistics);
  });
});
