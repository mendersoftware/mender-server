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

import { defaultState, token, undefineds } from '../../tests/mockData';
import { render } from '../../tests/setupTests';
import { DARK_MODE, LIGHT_MODE } from './constants/appConstants.js';
import {
  FileSize,
  customSort,
  deepCompare,
  detectOsIdentifier,
  duplicateFilter,
  extractSoftware,
  formatTime,
  fullyDecodeURI,
  generateDeploymentGroupDetails,
  getDebConfigurationCode,
  getDemoDeviceAddress,
  getFormattedSize,
  getPhaseDeviceCount,
  getRemainderPercent,
  groupDeploymentDevicesStats,
  groupDeploymentStats,
  isDarkMode,
  isEmpty,
  mapDeviceAttributes,
  preformatWithRequestID,
  standardizePhases,
  stringToBoolean,
  unionizeStrings,
  validatePhases,
  versionCompare
} from './helpers';

const deploymentCreationTime = defaultState.deployments.byId.d1.created;

/* eslint-disable sonarjs/no-duplicate-string */
describe('FileSize Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<FileSize fileSize={1000} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});

describe('getFormattedSize function', () => {
  it('converts correctly', async () => {
    expect(getFormattedSize()).toEqual('0 Bytes');
    expect(getFormattedSize(null)).toEqual('0 Bytes');
    expect(getFormattedSize(0)).toEqual('0 Bytes');
    expect(getFormattedSize(31)).toEqual('31.00 Bytes');
    expect(getFormattedSize(1024)).toEqual('1.00 KB');
    expect(getFormattedSize(1024 * 1024)).toEqual('1.00 MB');
    expect(getFormattedSize(1024 * 1024 * 2.5)).toEqual('2.50 MB');
    expect(getFormattedSize(1024 * 1024 * 1024 * 1.2345)).toEqual('1.23 GB');
  });
});

describe('isEmpty function', () => {
  it('should identify empty objects', async () => {
    expect(isEmpty({})).toEqual(true);
  });
  it('should identify non-empty objects', async () => {
    expect(isEmpty({ a: 1 })).toEqual(false);
  });
  it('should identify an object with nested empty objects as non-empty', async () => {
    expect(isEmpty({ a: {} })).toEqual(false);
  });
});

describe('stringToBoolean function', () => {
  it('should convert truthy objects', async () => {
    expect(stringToBoolean(1)).toEqual(true);
    expect(stringToBoolean('1')).toEqual(true);
    expect(stringToBoolean(true)).toEqual(true);
    expect(stringToBoolean('yes')).toEqual(true);
    expect(stringToBoolean('TRUE')).toEqual(true);
    expect(stringToBoolean('something')).toEqual(true);
  });
  it('should convert truthy objects', async () => {
    expect(stringToBoolean(0)).toEqual(false);
    expect(stringToBoolean('0')).toEqual(false);
    expect(stringToBoolean(false)).toEqual(false);
    expect(stringToBoolean('no')).toEqual(false);
    expect(stringToBoolean('FALSE')).toEqual(false);
  });
});

describe('versionCompare function', () => {
  it('should work as intended', async () => {
    expect(versionCompare('2.5.1', '2.6.0').toString()).toEqual('-1');
    expect(versionCompare('2.6.0', '2.6.0').toString()).toEqual('0');
    expect(versionCompare('2.6.x', '2.6.0').toString()).toEqual('1');
    expect(versionCompare('next', '2.6').toString()).toEqual('1');
    expect(versionCompare('', '2.6.0').toString()).toEqual('-1');
  });
});

const oldHostname = window.location.hostname;
const postTestCleanUp = () => {
  window.location = {
    ...window.location,
    hostname: oldHostname
  };
};
describe('getDebConfigurationCode function', () => {
  let code;
  describe('configuring devices for hosted mender', () => {
    beforeEach(() => {
      code = getDebConfigurationCode({ ipAddress: '192.168.7.41', isDemoMode: true, deviceType: 'raspberrypi3', token: 'omnomnom' });
    });
    afterEach(postTestCleanUp);
    it('should not contain any template string leftovers', async () => {
      expect(code).not.toMatch(/\$\{([^}]+)\}/);
    });
    it('should return a sane result', async () => {
      expect(code).toMatch(
        `wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --demo --server-ip 192.168.7.41`
      );
    });
    it('should not contain tenant information for OS calls', async () => {
      expect(code).not.toMatch(/tenant/);
      expect(code).not.toMatch(/token/);
      expect(code).not.toMatch(/TENANT/);
      expect(code).not.toMatch(/TOKEN/);
    });
  });
  describe('configuring devices for hosted mender', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'hosted.mender.io'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for hosted calls', async () => {
      code = getDebConfigurationCode({
        deviceType: 'raspberrypi3',
        hasMonitor: true,
        isHosted: true,
        isOnboarding: true,
        tenantToken: 'token',
        token: 'omnomnom'
      });
      expect(code).toMatch(
        `JWT_TOKEN="omnomnom"
TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --commercial --jwt-token $JWT_TOKEN --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-url https://hosted.mender.io --server-cert=""`
      );
    });
    it('should contain sane information for hosted calls by users without monitor access', async () => {
      code = getDebConfigurationCode({
        deviceType: 'raspberrypi3',
        isHosted: true,
        isOnboarding: true,
        tenantToken: 'token',
        token: 'omnomnom'
      });
      expect(code).toMatch(
        `JWT_TOKEN="omnomnom"
TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --jwt-token $JWT_TOKEN --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-url https://hosted.mender.io --server-cert=""`
      );
    });
  });
  describe('configuring devices for staging.hosted.mender', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'staging.hosted.mender.io'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for staging preview calls', async () => {
      code = getDebConfigurationCode({
        deviceType: 'raspberrypi3',
        hasMonitor: true,
        isHosted: true,
        isOnboarding: true,
        isPreRelease: true,
        tenantToken: 'token',
        token: 'omnomnom'
      });
      expect(code).toMatch(
        `JWT_TOKEN="omnomnom"
TENANT_TOKEN="token"
wget -O- https://get.mender.io/staging | sudo bash -s -- --demo -c experimental --commercial --jwt-token $JWT_TOKEN --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-url https://staging.hosted.mender.io --server-cert=""`
      );
    });
  });
  describe('configuring devices for fancy.enterprise.on.prem', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'fancy.enterprise.on.prem'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for enterprise demo on-prem calls', async () => {
      code = getDebConfigurationCode({ ipAddress: '1.2.3.4', isDemoMode: true, tenantToken: 'token', token: 'omnomnom', deviceType: 'raspberrypi3' });
      expect(code).toMatch(
        `TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-ip 1.2.3.4`
      );
    });
    it('should contain sane information for enterprise production on-prem calls', async () => {
      code = getDebConfigurationCode({ ipAddress: '1.2.3.4', isDemoMode: false, tenantToken: 'token', deviceType: 'raspberrypi3' });
      expect(code).toMatch(
        `TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --retry-poll 300 --update-poll 1800 --inventory-poll 28800 --server-url https://fancy.enterprise.on.prem --server-cert=""`
      );
    });
  });
  describe('configuring devices for fancy.opensource.on.prem', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'fancy.opensource.on.prem'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for OS demo on-prem calls', async () => {
      code = getDebConfigurationCode({ ipAddress: '1.2.3.4', isDemoMode: true, tenantToken: 'token', token: 'omnomnom', deviceType: 'raspberrypi3' });
      expect(code).toMatch(
        `wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-ip 1.2.3.4`
      );
    });
    it('should contain sane information for OS production on-prem calls', async () => {
      code = getDebConfigurationCode({ ipAddress: '1.2.3.4', isDemoMode: false, tenantToken: 'token', token: 'omnomnom', deviceType: 'raspberrypi3' });
      expect(code).toMatch(
        `wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --retry-poll 300 --update-poll 1800 --inventory-poll 28800 --server-url https://fancy.opensource.on.prem --server-cert=""`
      );
    });
  });
});

describe('duplicateFilter function', () => {
  it('removes duplicastes from an array', async () => {
    expect([].filter(duplicateFilter)).toEqual([]);
    expect([1, 1, 2, 3, 4, 5].filter(duplicateFilter)).toEqual([1, 2, 3, 4, 5]);
    expect(['hey', 'hey', 'ho', 'ho', 'ho', 'heyho'].filter(duplicateFilter)).toEqual(['hey', 'ho', 'heyho']);
  });
});

describe('unionizeStrings function', () => {
  it('joins arrays of strings to a list of distinct strings', async () => {
    expect(unionizeStrings([], [])).toEqual([]);
    expect(unionizeStrings(['hey', 'hey', 'ho', 'ho', 'ho', 'heyho'], ['hohoho'])).toEqual(['hey', 'ho', 'heyho', 'hohoho']);
    expect(unionizeStrings(['hohoho'], ['hey', 'hey', 'ho', 'ho', 'ho', 'heyho'])).toEqual(['hohoho', 'hey', 'ho', 'heyho']);
    expect(unionizeStrings(['hohoho', 'hohoho'], ['hey', 'hey', 'ho', 'ho', 'ho', 'heyho'])).toEqual(['hohoho', 'hey', 'ho', 'heyho']);
  });
});

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

describe('getPhaseDeviceCount function', () => {
  it('works with empty attributes', async () => {
    expect(getPhaseDeviceCount(120, 10, 20, false)).toEqual(12);
    expect(getPhaseDeviceCount(120, 10, 20, true)).toEqual(12);
    expect(getPhaseDeviceCount(120, null, 20, true)).toEqual(24);
    expect(getPhaseDeviceCount(120, null, 20, false)).toEqual(24);
    expect(getPhaseDeviceCount(undefined, null, 20, false)).toEqual(0);
  });
});
describe('customSort function', () => {
  it('works as expected', async () => {
    const creationSortedUp = Object.values(defaultState.deployments.byId).sort(customSort(false, 'created'));
    expect(creationSortedUp[1].id).toEqual(defaultState.deployments.byId.d1.id);
    expect(creationSortedUp[0].id).toEqual(defaultState.deployments.byId.d2.id);
    const creationSortedDown = Object.values(defaultState.deployments.byId).sort(customSort(true, 'created'));
    expect(creationSortedDown[0].id).toEqual(defaultState.deployments.byId.d1.id);
    expect(creationSortedDown[1].id).toEqual(defaultState.deployments.byId.d2.id);
    const idSortedUp = Object.values(defaultState.deployments.byId).sort(customSort(false, 'id'));
    expect(idSortedUp[0].id).toEqual(defaultState.deployments.byId.d1.id);
    expect(idSortedUp[1].id).toEqual(defaultState.deployments.byId.d2.id);
    const idSortedDown = Object.values(defaultState.deployments.byId).sort(customSort(true, 'id'));
    expect(idSortedDown[1].id).toEqual(defaultState.deployments.byId.d1.id);
    expect(idSortedDown[0].id).toEqual(defaultState.deployments.byId.d2.id);
  });
});
describe('deepCompare function', () => {
  it('works as expected', async () => {
    expect(deepCompare(false, 12)).toBeFalsy();
    expect(deepCompare(defaultState, {})).toBeFalsy();
    expect(
      deepCompare(defaultState, {
        ...defaultState,
        devices: { ...defaultState.devices, byId: { ...defaultState.devices.byId, a1: { ...defaultState.devices.byId.a1, id: 'test' } } }
      })
    ).toBeFalsy();
    expect(deepCompare({}, {})).toBeTruthy();
    expect(deepCompare({}, {}, {})).toBeTruthy();
    expect(deepCompare(defaultState.devices.byId, { ...defaultState.devices.byId }, { ...defaultState.devices.byId })).toBeTruthy();
    expect(deepCompare(['test', { test: 'test' }, 1], ['test', { test: 'test' }, 1])).toBeTruthy();
    expect(deepCompare(undefined, null)).toBeFalsy();
    expect(deepCompare(1, 2)).toBeFalsy();
    expect(deepCompare(1, 1)).toBeTruthy();
    const date = new Date();
    expect(deepCompare(date, date)).toBeTruthy();
    expect(deepCompare(date, new Date().setDate(date.getDate() - 1))).toBeFalsy();
    expect(deepCompare(<FileSize />, <FileSize />)).toBeTruthy();
    expect(deepCompare(defaultState, {})).toBeFalsy();
  });
});
describe('detectOsIdentifier function', () => {
  it('works as expected', async () => {
    expect(detectOsIdentifier()).toEqual('Linux');
    navigator.appVersion = '5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36';
    expect(detectOsIdentifier()).toEqual('MacOs');
  });
});
describe('formatTime function', () => {
  it('works as expected', async () => {
    expect(formatTime(new Date('2020-11-30T18:36:38.258Z'))).toEqual('2020-11-30T18:36:38.258');
    expect(formatTime('2020-11-30 18 : 36 : 38 . 258 UTC')).toEqual('2020-11-30T18:36:38.258');
  });
});
describe('fullyDecodeURI function', () => {
  it('works as expected', async () => {
    expect(fullyDecodeURI('http%3A%2F%2Ftest%20encoded%20%2520http%253A%252F%252Ftest%20%2520%20twice%20%C3%B8%C3%A6%C3%A5%C3%9F%2F%60%C2%B4')).toEqual(
      'http://test encoded  http://test   twice øæåß/`´'
    );
  });
});
describe('getDemoDeviceAddress function', () => {
  it('works as expected', async () => {
    expect(getDemoDeviceAddress(Object.values(defaultState.devices.byId), 'virtual')).toEqual('localhost');
    expect(getDemoDeviceAddress(Object.values(defaultState.devices.byId), 'physical')).toEqual('192.168.10.141');
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

describe('extractSoftware function', () => {
  it('works as expected', async () => {
    expect(
      extractSoftware({
        artifact_name: 'myapp',
        'rootfs-image.version': 'stablev1-beta-final-v0',
        'rootfs-image.checksum': '12341143',
        'test.version': 'test-2',
        'a.whole.lot.of.dots.version': 'test-3'
      })
    ).toEqual({
      nonSoftware: [['artifact_name', 'myapp']],
      software: [
        ['rootfs-image.version', 'stablev1-beta-final-v0'],
        ['rootfs-image.checksum', '12341143'],
        ['test.version', 'test-2'],
        ['a.whole.lot.of.dots.version', 'test-3']
      ]
    });
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

describe('standardizePhases function', () => {
  it('works as expected', async () => {
    const phases = [
      { batch_size: 10, delay: 2, delayUnit: 'hours', start_ts: deploymentCreationTime },
      { batch_size: 10, delay: 2, start_ts: deploymentCreationTime },
      { batch_size: 10, start_ts: deploymentCreationTime }
    ];
    expect(standardizePhases(phases)).toEqual([
      { batch_size: 10, delay: 2, delayUnit: 'hours' },
      { batch_size: 10, delay: 2, delayUnit: 'hours', start_ts: 1 },
      { batch_size: 10, start_ts: 2 }
    ]);
  });
});

describe('getRemainderPercent function', () => {
  const phases = [
    { batch_size: 10, not: 'interested' },
    { batch_size: 10, not: 'interested' },
    { batch_size: 10, not: 'interested' }
  ];
  expect(getRemainderPercent(phases)).toEqual(80);
  expect(
    getRemainderPercent([
      { batch_size: 10, not: 'interested' },
      { batch_size: 90, not: 'interested' }
    ])
  ).toEqual(90);
  expect(
    getRemainderPercent([
      { batch_size: 10, not: 'interested' },
      { batch_size: 95, not: 'interested' }
    ])
  ).toEqual(90);
  // this will be caught in the phase validation - should still be good to be fixed in the future
  expect(
    getRemainderPercent([
      { batch_size: 50, not: 'interested' },
      { batch_size: 55, not: 'interested' },
      { batch_size: 95, not: 'interested' }
    ])
  ).toEqual(-5);
});

describe('validatePhases function', () => {
  it('works as expected', async () => {
    const phases = [
      { batch_size: 10, delay: 2, delayUnit: 'hours', start_ts: deploymentCreationTime },
      { batch_size: 10, delay: 2, start_ts: deploymentCreationTime },
      { batch_size: 10, start_ts: deploymentCreationTime }
    ];
    expect(validatePhases(undefined, 10000, false)).toEqual(true);
    expect(validatePhases(undefined, 10000, true)).toEqual(true);
    expect(validatePhases(phases, 10, true)).toEqual(true);
    expect(validatePhases(phases, 10, true)).toEqual(true);
    expect(validatePhases([], 10, true)).toEqual(true);
    expect(
      validatePhases(
        [
          { batch_size: 50, not: 'interested' },
          { batch_size: 55, not: 'interested' },
          { batch_size: 95, not: 'interested' }
        ],
        10,
        false
      )
    ).toEqual(false);
    expect(
      validatePhases(
        [
          { batch_size: 50, not: 'interested' },
          { batch_size: 55, not: 'interested' },
          { batch_size: 95, not: 'interested' }
        ],
        100,
        true
      )
    ).toEqual(true);
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
