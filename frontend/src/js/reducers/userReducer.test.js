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
import { defaultState } from '../../../tests/mockData';
import * as UserConstants from '../constants/userConstants';
import reducer, { initialState } from './userReducer';

const testUser = {
  created_ts: '',
  email: 'test@example.com',
  id: '123',
  roles: ['RBAC_ROLE_PERMIT_ALL'],
  tfasecret: '',
  updated_ts: ''
};

const newDescription = 'new description';

describe('user reducer', () => {
  it('should return the initial state', async () => {
    expect(reducer(undefined, {})).toEqual(initialState);
  });

  it('should handle RECEIVED_QR_CODE', async () => {
    expect(reducer(undefined, { type: UserConstants.RECEIVED_QR_CODE, value: '123' }).qrCode).toEqual('123');
    expect(reducer(initialState, { type: UserConstants.RECEIVED_QR_CODE, value: '123' }).qrCode).toEqual('123');
  });

  it('should handle SUCCESSFULLY_LOGGED_IN', async () => {
    expect(reducer(undefined, { type: UserConstants.SUCCESSFULLY_LOGGED_IN, value: '123' }).currentSession).toEqual('123');
    expect(reducer(initialState, { type: UserConstants.SUCCESSFULLY_LOGGED_IN, value: '123' }).currentSession).toEqual('123');
  });

  it('should handle RECEIVED_USER_LIST', async () => {
    expect(reducer(undefined, { type: UserConstants.RECEIVED_USER_LIST, users: { '123': testUser } }).byId).toEqual({ '123': testUser });
    expect(reducer({ ...initialState, byId: { '123': testUser } }, { type: UserConstants.RECEIVED_USER_LIST, users: { '456': testUser } }).byId).toEqual({
      '456': testUser
    });
  });

  it('should handle RECEIVED_ACTIVATION_CODE', async () => {
    expect(reducer(undefined, { type: UserConstants.RECEIVED_ACTIVATION_CODE, code: 'code' }).activationCode).toEqual('code');
    expect(reducer({ ...initialState }, { type: UserConstants.RECEIVED_ACTIVATION_CODE, code: 'code' }).activationCode).toEqual('code');
  });

  it('should handle RECEIVED_USER', async () => {
    expect(reducer(undefined, { type: UserConstants.RECEIVED_USER, user: testUser }).byId).toEqual({ '123': testUser });
    expect(reducer({ ...initialState, byId: { '123': testUser } }, { type: UserConstants.RECEIVED_USER, user: testUser }).byId).toEqual({ '123': testUser });
  });

  it('should handle CREATED_USER', async () => {
    expect(reducer(undefined, { type: UserConstants.CREATED_USER, user: testUser }).byId).toEqual({ 0: testUser });
    expect(reducer({ ...initialState, byId: { '123': testUser } }, { type: UserConstants.CREATED_USER, user: testUser }).byId).toEqual({
      '123': testUser,
      0: testUser
    });
  });

  it('should handle REMOVED_USER', async () => {
    expect(reducer(undefined, { type: UserConstants.REMOVED_USER, userId: '123' }).byId).toEqual({});
    expect(reducer({ ...initialState, byId: { '123': testUser, '456': testUser } }, { type: UserConstants.REMOVED_USER, userId: '123' }).byId).toEqual({
      '456': testUser
    });
  });

  it('should handle UPDATED_USER', async () => {
    expect(reducer(undefined, { type: UserConstants.UPDATED_USER, userId: '123', user: testUser }).byId).toEqual({ '123': testUser });

    expect(
      reducer(
        { ...initialState, byId: { '123': testUser } },
        { type: UserConstants.UPDATED_USER, userId: '123', user: { ...testUser, email: 'test@mender.io' } }
      ).byId['123'].email
    ).toEqual('test@mender.io');
  });
  it('should handle RECEIVED_ROLES', async () => {
    const roles = reducer(undefined, { type: UserConstants.RECEIVED_ROLES, value: { ...defaultState.users.rolesById } }).rolesById;
    Object.entries(defaultState.users.rolesById).forEach(([key, role]) => expect(roles[key]).toEqual(role));
    expect(
      reducer(
        { ...initialState, rolesById: { ...defaultState.users.rolesById, thingsRole: { test: 'test' } } },
        { type: UserConstants.RECEIVED_ROLES, value: { ...defaultState.users.rolesById } }
      ).rolesById.thingsRole
    ).toBeFalsy();
  });
  it('should handle REMOVED_ROLE', async () => {
    // eslint-disable-next-line no-unused-vars
    const { [defaultState.users.rolesById.test.name]: removedRole, ...rolesById } = defaultState.users.rolesById;
    expect(reducer(undefined, { type: UserConstants.REMOVED_ROLE, value: defaultState.users.rolesById.test.name }).rolesById.test).toBeFalsy();
    expect(
      reducer({ ...initialState, rolesById: { ...defaultState.users.rolesById } }, { type: UserConstants.REMOVED_ROLE, value: rolesById }).rolesById.test
    ).toBeFalsy();
  });
  it('should handle CREATED_ROLE', async () => {
    expect(
      reducer(undefined, {
        type: UserConstants.CREATED_ROLE,
        roleId: 'newRole',
        role: { name: 'newRole', description: newDescription, groups: ['123'] }
      }).rolesById.newRole.description
    ).toEqual(newDescription);
    expect(
      reducer(
        { ...initialState },
        {
          type: UserConstants.CREATED_ROLE,
          roleId: 'newRole',
          role: { name: 'newRole', description: newDescription, groups: ['123'] }
        }
      ).rolesById.newRole.description
    ).toEqual(newDescription);
  });
  it('should handle UPDATED_ROLE', async () => {
    expect(
      reducer(undefined, { type: UserConstants.UPDATED_ROLE, roleId: 'RBAC_ROLE_CI', role: { description: newDescription } }).rolesById.RBAC_ROLE_CI.name
    ).toEqual('Releases Manager');
    expect(
      reducer({ ...initialState }, { type: UserConstants.UPDATED_ROLE, roleId: 'RBAC_ROLE_CI', role: { description: newDescription } }).rolesById.RBAC_ROLE_CI
        .name
    ).toEqual('Releases Manager');
  });
  it('should handle SET_CUSTOM_COLUMNS', async () => {
    expect(reducer(undefined, { type: UserConstants.SET_CUSTOM_COLUMNS, value: 'test' }).customColumns).toEqual('test');
    expect(reducer({ ...initialState }, { type: UserConstants.SET_CUSTOM_COLUMNS, value: 'test' }).customColumns).toEqual('test');
  });
  it('should handle SET_GLOBAL_SETTINGS', async () => {
    expect(reducer(undefined, { type: UserConstants.SET_GLOBAL_SETTINGS, settings: { newSetting: 'test' } }).globalSettings).toEqual({
      ...initialState.globalSettings,
      newSetting: 'test'
    });
    expect(reducer({ ...initialState }, { type: UserConstants.SET_GLOBAL_SETTINGS, settings: { newSetting: 'test' } }).globalSettings).toEqual({
      ...initialState.globalSettings,
      newSetting: 'test'
    });
  });
  it('should handle SET_SHOW_CONNECT_DEVICE', async () => {
    expect(reducer(undefined, { type: UserConstants.SET_SHOW_CONNECT_DEVICE, show: false }).showConnectDeviceDialog).toEqual(false);
    expect(reducer({ ...initialState }, { type: UserConstants.SET_SHOW_CONNECT_DEVICE, show: true }).showConnectDeviceDialog).toEqual(true);
  });
});
