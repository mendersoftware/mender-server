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
// @ts-nocheck
import { defaultState, token } from '../../../tests/mockData';
import { DARK_MODE, LIGHT_MODE } from './constants';
import {
  generateDeploymentGroupDetails,
  groupDeploymentDevicesStats,
  groupDeploymentStats,
  isDarkMode,
  mapDeviceAttributes,
  preformatWithRequestID
} from './utils';

describe('mapDeviceAttributes function', () => {
  const defaultAttributes = {
    inventory: { device_type: [], artifact_name: '' },
    identity: {},
    monitor: {},
    system: {},
    tags: {}
  };
  it('works with empty attributes', async () => {
    expect(mapDeviceAttributes()).toEqual(defaultAttributes);
    expect(mapDeviceAttributes([])).toEqual(defaultAttributes);
  });
  it('handles unscoped attributes', async () => {
    const testAttributesObject1 = { name: 'this1', value: 'that1' };
    expect(mapDeviceAttributes([testAttributesObject1])).toEqual({
      ...defaultAttributes,
      inventory: {
        ...defaultAttributes.inventory,
        this1: 'that1'
      }
    });
    const testAttributesObject2 = { name: 'this2', value: 'that2' };
    expect(mapDeviceAttributes([testAttributesObject1, testAttributesObject2])).toEqual({
      ...defaultAttributes,
      inventory: {
        ...defaultAttributes.inventory,
        this1: 'that1',
        this2: 'that2'
      }
    });
    expect(mapDeviceAttributes([testAttributesObject1, testAttributesObject2, testAttributesObject2])).toEqual({
      ...defaultAttributes,
      inventory: {
        ...defaultAttributes.inventory,
        this1: 'that1',
        this2: 'that2'
      }
    });
  });
  it('handles scoped attributes', async () => {
    const testAttributesObject1 = { name: 'this1', value: 'that1', scope: 'inventory' };
    expect(mapDeviceAttributes([testAttributesObject1])).toEqual({
      ...defaultAttributes,
      inventory: {
        ...defaultAttributes.inventory,
        this1: 'that1'
      }
    });
    const testAttributesObject2 = { name: 'this2', value: 'that2', scope: 'identity' };
    expect(mapDeviceAttributes([testAttributesObject1, testAttributesObject2])).toEqual({
      ...defaultAttributes,
      identity: {
        ...defaultAttributes.identity,
        this2: 'that2'
      },
      inventory: {
        ...defaultAttributes.inventory,
        this1: 'that1'
      }
    });
    expect(mapDeviceAttributes([testAttributesObject1, testAttributesObject2, testAttributesObject2])).toEqual({
      ...defaultAttributes,
      identity: {
        ...defaultAttributes.identity,
        this2: 'that2'
      },
      inventory: {
        ...defaultAttributes.inventory,
        this1: 'that1'
      }
    });
  });
});

describe('preformatWithRequestID function', () => {
  it('works as expected', async () => {
    expect(preformatWithRequestID({ data: { request_id: 'someUuidLikeLongerText' } }, token)).toEqual(
      'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjZTNkMGY4Yy1hZWRlLTQwMzAtYjM5MS03ZDUwMjBlYjg3M2UiLCJzdWIiOiJhMzBhNzgwYi1iODQzLTUzNDQtODBlMy0wZmQ5NWE0ZjZmYzMiLCJleHAiOjE2MDY4MTUzNjksImlhdCI6MTYwNjIxMDU2OSwibWVuZGVyLnRlbmF... [Request ID: someUuid]'
    );
    expect(preformatWithRequestID({ data: {} }, token)).toEqual(
      'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjZTNkMGY4Yy1hZWRlLTQwMzAtYjM5MS03ZDUwMjBlYjg3M2UiLCJzdWIiOiJhMzBhNzgwYi1iODQzLTUzNDQtODBlMy0wZmQ5NWE0ZjZmYzMiLCJleHAiOjE2MDY4MTUzNjksImlhdCI6MTYwNjIxMDU2OSwibWVuZGVyLnRlbmF...'
    );
    expect(preformatWithRequestID(undefined, token)).toEqual(
      'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjZTNkMGY4Yy1hZWRlLTQwMzAtYjM5MS03ZDUwMjBlYjg3M2UiLCJzdWIiOiJhMzBhNzgwYi1iODQzLTUzNDQtODBlMy0wZmQ5NWE0ZjZmYzMiLCJleHAiOjE2MDY4MTUzNjksImlhdCI6MTYwNjIxMDU2OSwibWVuZGVyLnRlbmF...'
    );
    const expectedText = 'short text';
    expect(preformatWithRequestID({ data: { request_id: 'someUuidLikeLongerText' } }, expectedText)).toEqual('short text [Request ID: someUuid]');
    expect(preformatWithRequestID(undefined, expectedText)).toEqual(expectedText);
  });
});

describe('generateDeploymentGroupDetails function', () => {
  it('works as expected', async () => {
    expect(generateDeploymentGroupDetails({ terms: defaultState.devices.groups.byId.testGroupDynamic.filters }, 'testGroupDynamic')).toEqual(
      'testGroupDynamic (group = things)'
    );
    expect(
      generateDeploymentGroupDetails(
        {
          terms: [
            { scope: 'system', key: 'group', operator: '$eq', value: 'things' },
            { scope: 'system', key: 'group', operator: '$nexists', value: 'otherThings' },
            { scope: 'system', key: 'group', operator: '$nin', value: 'a,small,list' }
          ]
        },
        'testGroupDynamic'
      )
    ).toEqual(`testGroupDynamic (group = things, group doesn't exist otherThings, group not in a,small,list)`);
    expect(generateDeploymentGroupDetails({ terms: undefined }, 'testGroupDynamic')).toEqual('testGroupDynamic');
  });
});

describe('deployment stats grouping functions', () => {
  it('groups correctly based on deployment stats', async () => {
    let deployment = {
      statistics: {
        status: {
          aborted: 2,
          'already-installed': 1,
          decommissioned: 1,
          downloading: 3,
          failure: 1,
          installing: 1,
          noartifact: 1,
          pending: 2,
          paused: 0,
          rebooting: 1,
          success: 1
        }
      }
    };
    expect(groupDeploymentStats(deployment)).toEqual({ inprogress: 5, paused: 0, pending: 2, successes: 3, failures: 4 });
    deployment = { ...deployment, max_devices: 100, device_count: 10 };
    expect(groupDeploymentStats(deployment)).toEqual({ inprogress: 5, paused: 0, pending: 92, successes: 3, failures: 4 });
  });
  it('groups correctly based on deployment devices states', async () => {
    const deployment = {
      devices: {
        a: { status: 'aborted' },
        b: { status: 'already-installed' },
        c: { status: 'decommissioned' },
        d: { status: 'downloading' },
        e: { status: 'failure' },
        f: { status: 'installing' },
        g: { status: 'noartifact' },
        h: { status: 'pending' },
        i: { status: 'rebooting' },
        j: { status: 'success' }
      }
    };
    expect(groupDeploymentDevicesStats(deployment)).toEqual({ inprogress: 3, paused: 0, pending: 1, successes: 3, failures: 3 });
  });
});

describe('isDarkMode function', () => {
  it('should return `true` if DARK_MODE was passed in', () => {
    expect(isDarkMode(DARK_MODE)).toEqual(true);
  });
  it('should return `false` if LIGHT_MODE was passed in', () => {
    expect(isDarkMode(LIGHT_MODE)).toEqual(false);
  });
});
