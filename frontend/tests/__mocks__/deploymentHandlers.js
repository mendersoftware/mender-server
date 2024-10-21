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
import { deploymentsApiUrl, deploymentsApiUrlV2, headerNames, limitDefault } from '@northern.tech/store/constants';
import { HttpResponse, http } from 'msw';

import { defaultState } from '../mockData';

const createdDeployment = {
  ...defaultState.deployments.byId.d1,
  id: 'created-123'
};

const defaultDeploymentConfig = {
  delta: {
    enabled: true,
    binary_delta: {
      xdelta_args: {
        disable_checksum: false,
        disable_external_decompression: false,
        compression_level: 6,
        source_window_size: 0,
        input_window_size: 0,
        compression_duplicates_window: 0,
        instruction_buffer_size: 0
      },
      timeout: 0
    },
    binary_delta_limits: {
      xdelta_args_limits: {
        source_window_size: limitDefault,
        input_window_size: limitDefault,
        compression_duplicates_window: limitDefault,
        instruction_buffer_size: limitDefault
      },
      timeout: { min: 60, max: 3600, default: 60 },
      jobs_in_parallel: limitDefault,
      queue_length: limitDefault
    }
  }
};

export const deploymentHandlers = [
  http.get(
    `${deploymentsApiUrl}/deployments`,
    () =>
      new HttpResponse(JSON.stringify(Object.values(defaultState.deployments.byId).reverse()), {
        headers: { [headerNames.total]: Object.keys(defaultState.deployments.byId).length }
      })
  ),
  http.get(`${deploymentsApiUrl}/deployments/releases`, ({ request }) => {
    const { searchParams } = new URL(request.url);
    const releaseName = searchParams.get('name');
    const release = defaultState.releases.byId[releaseName] || {};
    if (releaseName) {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { device_types_compatible, ...remainder } = release;
      return Object.keys(remainder).length || releaseName === 'createdRelease' ? HttpResponse.json([remainder]) : new HttpResponse(null, { status: 520 });
    }
    const releases = Object.values(defaultState.releases.byId).map(stateRelease => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { device_types_compatible, ...remainder } = stateRelease;
      return remainder;
    });
    return HttpResponse.json(releases);
  }),
  http.get(`${deploymentsApiUrl}/deployments/:deploymentId`, ({ params: { deploymentId } }) => {
    if (deploymentId === createdDeployment.id) {
      return HttpResponse.json(createdDeployment);
    } else if (deploymentId === 'config1') {
      return HttpResponse.json({
        ...createdDeployment,
        id: 'config1',
        created: '2019-01-01T09:25:01.000Z',
        finished: '2019-01-01T09:25:03.000Z',
        status: 'finished'
      });
    }
    return HttpResponse.json(defaultState.deployments.byId[deploymentId]);
  }),
  http.get(`${deploymentsApiUrl}/deployments/:deploymentId/devices`, ({ params: { deploymentId } }) => {
    if (deploymentId === createdDeployment.id) {
      return HttpResponse.json(Object.values(createdDeployment.devices));
    } else if (defaultState.deployments.byId[deploymentId]) {
      return HttpResponse.json(Object.values(defaultState.deployments.byId[deploymentId].devices));
    }
    return new HttpResponse(null, { status: 521 });
  }),
  http.post(`${deploymentsApiUrl}/deployments/statistics/list`, async ({ request }) => {
    const { deployment_ids = [] } = await request.json();
    if (deployment_ids.includes(createdDeployment.id)) {
      return HttpResponse.json([]);
    } else if (deployment_ids.every(id => defaultState.deployments.byId[id])) {
      const stats = deployment_ids.map(id => ({ id, stats: defaultState.deployments.byId[id].statistics.status }));
      return HttpResponse.json(stats);
    }
    return new HttpResponse(null, { status: 522 });
  }),
  http.get(`${deploymentsApiUrl}/deployments/:deploymentId/devices/:deviceId/log`, ({ params: { deploymentId, deviceId } }) => {
    if (defaultState.deployments.byId[deploymentId] && defaultState.deployments.byId[deploymentId].devices[deviceId]) {
      return HttpResponse.text('test');
    }
    return new HttpResponse(null, { status: 523 });
  }),
  http.post(`${deploymentsApiUrl}/deployments`, async ({ request }) => {
    const body = await request.json();
    if (!Object.keys(body).length) {
      return new HttpResponse(JSON.stringify({}), { status: 524 });
    }
    return new HttpResponse(JSON.stringify({}), { headers: { location: `find/me/here/${createdDeployment.id}` } });
  }),
  http.post(`${deploymentsApiUrlV2}/deployments`, async ({ request }) => {
    const { filter_id, devices = [] } = await request.json();
    if (!filter_id || !!devices.length) {
      return new HttpResponse(JSON.stringify({}), { status: 525 });
    }
    return new HttpResponse(JSON.stringify({}), { headers: { location: `find/me/here/${createdDeployment.id}` } });
  }),
  http.post(`${deploymentsApiUrl}/deployments/group/:deploymentGroup`, async ({ params: { deploymentGroup }, request }) => {
    const { filter_id, devices = [] } = await request.json();
    if (filter_id || !!devices.length || deploymentGroup !== Object.keys(defaultState.devices.groups.byId)[0]) {
      return new HttpResponse(JSON.stringify({}), { status: 526 });
    }
    return new HttpResponse(JSON.stringify({}), { headers: { location: `find/me/here/${createdDeployment.id}` } });
  }),
  http.patch(`${deploymentsApiUrl}/deployments/:deploymentId`, async ({ params: { deploymentId }, request }) => {
    const { update_control_map } = await request.json();
    if (deploymentId === createdDeployment.id && Object.keys(update_control_map).length) {
      return new HttpResponse(null, { status: 204 });
    }
    return new HttpResponse(null, { status: 581 });
  }),
  http.put(`${deploymentsApiUrl}/deployments/:deploymentId/status`, async ({ params: { deploymentId }, request }) => {
    const { status } = await request.json();
    return new HttpResponse(JSON.stringify({}), {
      status:
        status === 'aborted' &&
        [...defaultState.deployments.byStatus.pending.deploymentIds, ...defaultState.deployments.byStatus.inprogress.deploymentIds].includes(deploymentId)
          ? 200
          : 528
    });
  }),
  http.get(`${deploymentsApiUrl}/deployments/:deploymentId/devices/list`, ({ params: { deploymentId } }) => {
    if (deploymentId === createdDeployment.id) {
      return new HttpResponse(JSON.stringify(Object.values(createdDeployment.devices)), {
        headers: { [headerNames.total]: Object.keys(createdDeployment.devices).length }
      });
    } else if (defaultState.deployments.byId[deploymentId]) {
      return new HttpResponse(JSON.stringify(Object.values(defaultState.deployments.byId[deploymentId].devices)), {
        headers: { [headerNames.total]: Object.keys(defaultState.deployments.byId[deploymentId].devices).length }
      });
    }
    return new HttpResponse(null, { status: 529 });
  }),
  http.get(`${deploymentsApiUrl}/config`, () => HttpResponse.json(defaultDeploymentConfig)),
  http.put(`${deploymentsApiUrl}/config/binary_delta`, () => new HttpResponse(null, { status: 200 })),
  http.get(`${deploymentsApiUrl}/deployments/devices/:deviceId`, ({ params: { deviceId } }) => {
    if (deviceId === defaultState.devices.byId.a1.id) {
      return new HttpResponse(
        JSON.stringify([
          {
            id: createdDeployment.id + 'something',
            deployment: { ...createdDeployment, id: defaultState.deployments.byId.d1.id, status: 'inprogress' },
            device: {
              created: '2019-01-01T12:35:00.000Z',
              finished: '2019-01-01T12:40:00.000Z',
              status: 'noartifact',
              id: 'something',
              image: { ...defaultState.releases.byId.r1 }
            },
            log: true,
            attempts: 1,
            delta_job_id: ''
          }
        ]),
        {
          headers: { [headerNames.total]: 34 }
        }
      );
    }
    return new HttpResponse(null, { status: 529 });
  }),
  http.delete(`${deploymentsApiUrl}/deployments/devices/:deviceId/history`, () => new HttpResponse(null, { status: 204 }))
];
