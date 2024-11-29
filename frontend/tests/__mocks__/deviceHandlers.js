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
import {
  DEVICE_FILTERING_OPTIONS,
  DEVICE_STATES,
  deviceAuthV2,
  deviceConfig,
  deviceConnect,
  headerNames,
  inventoryApiUrl,
  inventoryApiUrlV2,
  iotManagerBaseURL,
  reportingApiUrl
} from '@northern.tech/store/constants';
import { HttpResponse, http } from 'msw';

import { defaultCreationDate, defaultMacAddress, defaultState } from '../mockData';

const deviceAuthDevice = {
  id: defaultState.devices.byId.a1.id,
  identity_data: { mac: defaultMacAddress },
  status: 'accepted',
  decommissioning: false,
  created_ts: defaultCreationDate,
  updated_ts: '2019-01-01T09:25:00.000Z',
  auth_sets: [
    {
      id: 'auth1',
      identity_data: { mac: defaultMacAddress },
      pubkey: '-----BEGIN PUBLIC KEY-----\nMIIBojWELzgJ62hcXIhAfqfoNiaB1326XZByZwcnHr5BuSPAgMBAAE=\n-----END PUBLIC KEY-----\n',
      ts: defaultCreationDate,
      status: 'accepted'
    }
  ]
};

export const inventoryDevice = {
  id: defaultState.devices.byId.a1.id,
  attributes: [
    { name: 'network_interfaces', value: ['eth0', 'wlan0'], scope: 'inventory' },
    { name: 'kernel', value: 'Linux version 4.19.75-v7l+', scope: 'inventory' },
    { name: 'artifact_name', value: 'raspotifyInstaller', scope: 'inventory' },
    { name: 'ipv6_wlan0', value: 'fe80::68d9:8453:3e9c:7c80/64', scope: 'inventory' },
    { name: 'mac_eth0', value: defaultMacAddress, scope: 'inventory' },
    { name: 'rootfs_type', value: 'ext4', scope: 'inventory' },
    { name: 'mender_bootloader_integration', value: 'unknown', scope: 'inventory' },
    { name: 'cpu_model', value: 'ARMv7 Processor rev 3 (v7l)', scope: 'inventory' },
    { name: 'hostname', value: 'raspberrypi', scope: 'inventory' },
    { name: 'mender_client_version', value: '2.1.0', scope: 'inventory' },
    { name: 'ipv4_wlan0', value: '192.168.10.141/24', scope: 'inventory' },
    { name: 'device_type', value: 'raspberrypi4', scope: 'inventory' },
    { name: 'mem_total_kB', value: '1986024', scope: 'inventory' },
    { name: 'mac_wlan0', value: 'dc:a6:32:12:ad:c0', scope: 'inventory' },
    { name: 'os', value: 'Raspbian GNU/Linux 10 (buster)', scope: 'inventory' },
    { name: 'created_ts', value: defaultCreationDate, scope: 'system' },
    { name: 'updated_ts', value: defaultState.devices.byId.a1.updated_ts, scope: 'system' },
    { name: 'status', value: 'accepted', scope: 'identity' },
    { name: 'mac', value: defaultMacAddress, scope: 'identity' },
    { name: 'group', value: 'test', scope: 'system' }
  ],
  check_in_time: defaultState.devices.byId.a1.check_in_time,
  updated_ts: defaultState.devices.byId.a1.updated_ts
};

const deviceAttributes = [
  { name: 'status', scope: 'identity', count: 5 },
  { name: 'created_ts', scope: 'system', count: 5 },
  { name: 'updated_ts', scope: 'system', count: 5 },
  { name: 'mac', scope: 'identity', count: 1 },
  { name: 'artifact_name', scope: 'inventory', count: 1 },
  { name: 'cpu_model', scope: 'inventory', count: 1 },
  { name: 'device_type', scope: 'inventory', count: 1 },
  { name: 'hostname', scope: 'inventory', count: 1 },
  { name: 'ipv4_wlan0', scope: 'inventory', count: 1 },
  { name: 'ipv6_wlan0', scope: 'inventory', count: 1 },
  { name: 'kernel', scope: 'inventory', count: 1 },
  { name: 'mac_eth0', scope: 'inventory', count: 1 },
  { name: 'mac_wlan0', scope: 'inventory', count: 1 },
  { name: 'mem_total_kB', scope: 'inventory', count: 1 },
  { name: 'mender_bootloader_integration', scope: 'inventory', count: 1 },
  { name: 'mender_client_version', scope: 'inventory', count: 1 },
  { name: 'network_interfaces', scope: 'inventory', count: 1 },
  { name: 'os', scope: 'inventory', count: 1 },
  { name: 'rootfs_type', scope: 'inventory', count: 1 },
  { name: 'group', scope: 'system', count: 1 }
];

const searchHandler = async ({ request }) => {
  const { page, per_page, filters } = await request.json();
  if ([page, per_page, filters].some(item => !item)) {
    return new HttpResponse(null, { status: 509 });
  }
  const filter = filters.find(filter => filter.scope === 'identity' && filter.attribute === 'status' && Object.values(DEVICE_STATES).includes(filter.value));
  const status = filter?.value || '';
  if (!status || filters.length > 1) {
    if (filters.find(filter => filter.attribute === 'group' && filter.value.includes(Object.keys(defaultState.devices.groups.byId)[0]))) {
      return new HttpResponse(JSON.stringify([inventoryDevice, { ...inventoryDevice, id: defaultState.devices.byId.b1.id }]), {
        headers: { [headerNames.total]: 2 }
      });
    }
    if (filters.find(filter => filter.scope === 'monitor' && ['failed_last_update', 'alerts', 'auth_request'].includes(filter.attribute))) {
      return new HttpResponse(JSON.stringify([inventoryDevice]), { headers: { [headerNames.total]: 4 } });
    }
    return new HttpResponse(JSON.stringify([]), { headers: { [headerNames.total]: 0 } });
  }
  let deviceList = Array.from({ length: defaultState.devices.byStatus[status].total }, (_, index) => ({
    ...inventoryDevice,
    attributes: [...inventoryDevice.attributes, { name: 'test-count', value: index, scope: 'system' }],
    id: `${String.fromCharCode('a'.charCodeAt(0) + index)}1`
  }));
  deviceList = deviceList.slice((page - 1) * per_page, page * per_page);
  return new HttpResponse(JSON.stringify(deviceList), { headers: { [headerNames.total]: defaultState.devices.byStatus[status].total } });
};

export const deviceHandlers = [
  http.delete(`${deviceAuthV2}/devices/:deviceId/auth/:authId`, ({ params: { authId, deviceId } }) => {
    if (defaultState.devices.byId[deviceId].auth_sets.find(authSet => authSet.id === authId)) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 501 });
  }),
  http.delete(`${deviceAuthV2}/devices/:deviceId`, ({ params: { deviceId } }) => {
    if (defaultState.devices.byId[deviceId]) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 502 });
  }),
  http.delete(`${inventoryApiUrl}/groups/:group`, ({ params: { group } }) => {
    if (defaultState.devices.groups.byId[group]) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 515 });
  }),
  http.delete(`${inventoryApiUrl}/groups/:group/devices`, async ({ params: { group }, request }) => {
    const deviceIds = await request.json();
    if (defaultState.devices.groups.byId[group] && deviceIds.every(id => !!defaultState.devices.byId[id])) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 503 });
  }),
  http.delete(`${inventoryApiUrlV2}/filters/:filterId`, ({ params: { filterId } }) => {
    if (Object.values(defaultState.devices.groups.byId).find(group => group.id === filterId)) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 504 });
  }),
  http.get(`${deviceAuthV2}/devices`, ({ request }) => {
    const { searchParams } = new URL(request.url);
    const deviceIds = searchParams.getAll('id');
    if (deviceIds.every(id => !!defaultState.devices.byId[id])) {
      return HttpResponse.json(deviceIds.map(id => ({ ...deviceAuthDevice, id })));
    }
    return new HttpResponse(null, { status: 505 });
  }),
  http.get(`${deviceAuthV2}/limits/max_devices`, () => HttpResponse.json({ limit: defaultState.devices.limit })),
  http.get(`${inventoryApiUrl}/devices/:deviceId`, ({ params: { deviceId } }) => {
    if (defaultState.devices.byId[deviceId]) {
      return HttpResponse.json(inventoryDevice);
    }
    return new HttpResponse(null, { status: 506 });
  }),
  http.put(`${inventoryApiUrl}/devices/:deviceId/tags`, async ({ params: { deviceId }, request }) => {
    const tags = await request.json();
    if (!defaultState.devices.byId[deviceId] && !Array.isArray(tags) && !tags.every(item => item.name && item.value)) {
      return new HttpResponse(null, { status: 506 });
    }
    return HttpResponse.json(tags);
  }),
  http.get(`${inventoryApiUrl}/groups`, () => {
    const groups = Object.entries(defaultState.devices.groups.byId).reduce((accu, [groupName, group]) => {
      if (!group.id) {
        accu.push(groupName);
      }
      return accu;
    }, []);
    return HttpResponse.json(groups);
  }),
  http.get(`${inventoryApiUrlV2}/filters/attributes`, () => HttpResponse.json(deviceAttributes)),
  http.get(`${reportingApiUrl}/devices/attributes`, () => HttpResponse.json({ attributes: deviceAttributes, count: deviceAttributes.length, limit: 100 })),
  http.get(`${reportingApiUrl}/devices/search/attributes`, () => HttpResponse.json(deviceAttributes)),
  http.post(`${reportingApiUrl}/devices/aggregate`, () =>
    HttpResponse.json([
      {
        name: '*',
        items: [
          { key: 'test', count: 6 },
          { key: 'original', count: 1 }
        ],
        other_count: 42
      }
    ])
  ),
  http.get(`${inventoryApiUrlV2}/filters`, () =>
    HttpResponse.json([
      {
        id: 'filter1',
        name: 'testGroupDynamic',
        terms: [
          { scope: 'identity', attribute: 'id', type: '$in', value: ['a1'] },
          { scope: 'identity', attribute: 'mac', type: '$exists', value: false },
          { scope: 'identity', attribute: 'kernel', type: '$exists', value: true }
        ]
      }
    ])
  ),
  http.patch(`${inventoryApiUrl}/groups/:group/devices`, async ({ params: { group }, request }) => {
    const deviceIds = await request.json();
    if (!!group && deviceIds.every(id => !!defaultState.devices.byId[id])) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 508 });
  }),
  http.post(`${deviceAuthV2}/devices`, async ({ request }) => {
    const authset = await request.json();
    if (
      Object.values(defaultState.devices.byId).some(device =>
        device.auth_sets.some(deviceAuthSet => deviceAuthSet.pubkey == authset.pubkey || deviceAuthSet.identity_data == authset.identity_data)
      )
    ) {
      return new HttpResponse(null, { status: 409 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.post(`${inventoryApiUrlV2}/filters/search`, searchHandler),
  http.post(`${reportingApiUrl}/filters/search`, searchHandler),
  http.post(`${inventoryApiUrlV2}/filters`, async ({ request }) => {
    const { name, terms } = await request.json();
    if (
      [name, terms].some(item => !item) ||
      defaultState.devices.groups[name] ||
      !terms.every(term => DEVICE_FILTERING_OPTIONS[term.type] && ['identity', 'inventory', 'system'].includes(term.scope) && !!term.value && !!term.attribute)
    ) {
      return new HttpResponse(null, { status: 510 });
    }
    return new HttpResponse(JSON.stringify({}), { headers: { location: 'find/me/here/createdFilterId' } });
  }),
  http.put(`${deviceAuthV2}/devices/:deviceId/auth/:authId/status`, async ({ params: { authId, deviceId }, request }) => {
    const { status } = await request.json();
    if (defaultState.devices.byId[deviceId].auth_sets.find(authSet => authSet.id === authId) && DEVICE_STATES[status]) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 511 });
  }),
  http.get(`${deviceConfig}/:deviceId`, ({ params: { deviceId } }) => {
    if (deviceId === 'testId') {
      return new HttpResponse(JSON.stringify({ error: { status_code: 404 } }), { status: 404 });
    }
    if (defaultState.devices.byId[deviceId]) {
      return HttpResponse.json({
        configured: { test: true, something: 'else', aNumber: 42 },
        deployment_id: 'config1',
        reported: { test: true, something: 'else', aNumber: 42 },
        updated_ts: defaultState.devices.byId.a1.updated_ts,
        reported_ts: '2019-01-01T09:25:01.000Z'
      });
    }
    return new HttpResponse(null, { status: 512 });
  }),
  http.put(`${deviceConfig}/:deviceId`, async ({ params: { deviceId }, request }) => {
    const body = await request.text();
    if (body.includes('evilValue')) {
      return new HttpResponse(null, { status: 418 });
    }
    return new HttpResponse(null, { status: defaultState.devices.byId[deviceId] ? 201 : 513 });
  }),
  http.post(`${deviceConfig}/:deviceId/deploy`, ({ params: { deviceId } }) => {
    if (defaultState.devices.byId[deviceId]) {
      return HttpResponse.json({ deployment_id: 'config1' });
    }
    return new HttpResponse(null, { status: 514 });
  }),
  http.get(`${deviceConnect}/devices/:deviceId`, ({ params: { deviceId } }) => {
    if (deviceId === 'testId') {
      return new HttpResponse(JSON.stringify({ error: { status_code: 404 } }), { status: 404 });
    }
    if (defaultState.devices.byId[deviceId]) {
      return HttpResponse.json({ status: 'connected', updated_ts: defaultState.devices.byId[deviceId].updated_ts });
    }
    return new HttpResponse(null, { status: 512 });
  }),
  http.post(`${deviceConnect}/devices/:deviceId/:action`, ({ params: { deviceId, action } }) => {
    if (['check-update', 'send-inventory'].includes(action) && defaultState.devices.byId[deviceId]) {
      return new HttpResponse(null, { status: 202 });
    }
    return new HttpResponse(null, { status: 518 });
  }),
  http.get(`${iotManagerBaseURL}/devices/:deviceId/state`, ({ params: { deviceId } }) => {
    if (defaultState.devices.byId[deviceId]) {
      return HttpResponse.json({ deployment_id: defaultState.deployments.byId.d1.id });
    }
    return new HttpResponse(null, { status: 515 });
  }),
  http.put(`${iotManagerBaseURL}/devices/:deviceId/state/:integrationId`, async ({ params: { deviceId }, request }) => {
    const body = await request.json();
    if (defaultState.devices.byId[deviceId] && body) {
      return HttpResponse.json({ deployment_id: defaultState.deployments.byId.d1.id });
    }
    return new HttpResponse(null, { status: 516 });
  }),
  http.put(
    `${deviceConnect}/devices/:deviceId/upload`,
    ({ params: { deviceId } }) => new HttpResponse(null, { status: defaultState.devices.byId[deviceId] ? 200 : 517 })
  ),
  http.get(`${deviceAuthV2}/reports/devices`, () => HttpResponse.text('test,report'))
];
