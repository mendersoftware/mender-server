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

describe('device reducer', () => {
  it('should handle ADD_TO_GROUP', async () => {
    let state = reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId });
    expect(reducer(state, { type: actions.addToGroup, payload: { group: 'testExtra', deviceIds: ['d1'] } }).groups.byId.testExtra.deviceIds).toHaveLength(1);
    expect(
      reducer(initialState, { type: actions.addToGroup, payload: { group: 'testGroup', deviceIds: ['123', '1243'] } }).groups.byId.testGroup.deviceIds
    ).toHaveLength(2);
  });
  it('should handle REMOVE_FROM_GROUP', async () => {
    let state = reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId });
    state = reducer(state, { type: actions.selectGroup, payload: 'testGroup' });
    expect(
      reducer(state, { type: actions.removeFromGroup, payload: { group: 'testGroup', deviceIds: [defaultState.devices.groups.byId.testGroup.deviceIds[0]] } })
        .groups.byId.testGroup.deviceIds
    ).toHaveLength(defaultState.devices.groups.byId.testGroup.deviceIds.length - 1);
    expect(
      reducer(state, { type: actions.removeFromGroup, payload: { group: 'testGroup', deviceIds: defaultState.devices.groups.byId.testGroup.deviceIds } }).groups
        .byId.testGroup
    ).toBeFalsy();
    expect(
      reducer(initialState, { type: actions.removeFromGroup, payload: { group: 'testExtra', deviceIds: ['123', '1243'] } }).groups.byId.testExtra
    ).toBeFalsy();
  });
  it('should handle ADD_DYNAMIC_GROUP', async () => {
    expect(
      reducer(undefined, { type: actions.addGroup, payload: { groupName: 'test', group: { something: 'test' } } }).groups.byId.test.something
    ).toBeTruthy();
    expect(
      reducer(initialState, { type: actions.addGroup, payload: { groupName: 'test', group: { something: 'test' } } }).groups.byId.test.something
    ).toBeTruthy();
  });
});
