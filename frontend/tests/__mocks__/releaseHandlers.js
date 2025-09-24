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
import { SORTING_OPTIONS, deploymentsApiUrl, deploymentsApiUrlV2, headerNames } from '@northern.tech/store/constants';
import { customSort } from '@northern.tech/utils/helpers';
import { HttpResponse, http } from 'msw';

import { defaultState, releasesList } from '../mockData';

const deltaJobs = {
  'delta-job-1': {
    id: 'delta-job-1',
    to_release: 'mender-demo-artifact-3.3.1',
    from_release: 'mender-demo-artifact-3.2.1',
    to_version: 'mender-demo-artifact-3.3.1',
    from_version: 'mender-demo-artifact-3.2.1',
    device_types_compatible: ['qemux86-64'],
    started: '2022-07-11T20:49:00.000Z',
    status: 'success'
  },
  'delta-job-2': {
    id: 'delta-job-2',
    to_release: 'mender-demo-artifact-3.3.1',
    from_release: 'mender-demo-artifact-3.3.0',
    to_version: 'mender-demo-artifact-3.3.1',
    from_version: 'mender-demo-artifact-3.3.0',
    device_types_compatible: ['raspberrypi0w', 'raspberrypi0-wifi', 'raspberrypi3', 'raspberrypi4'],
    started: '2022-07-11T20:49:00.000Z',
    status: 'failed'
  },
  'delta-job-3': {
    id: 'delta-job-3',
    to_release: 'mender-demo-artifact-3.3.1',
    from_release: 'mender-demo-artifact-3.3.0',
    to_version: 'mender-demo-artifact-3.3.1',
    from_version: 'mender-demo-artifact-3.3.0',
    device_types_compatible: ['beaglebone', 'beaglebone-yocto', 'beaglebone-yocto'],
    started: '2022-07-11T20:49:00.000Z',
    status: 'pending'
  },
  'delta-job-4': {
    id: 'delta-job-4',
    to_release: 'mender-demo-artifact-3.2.0',
    from_release: 'mender-demo-artifact-3.1.0',
    to_version: 'mender-demo-artifact-3.2.0',
    from_version: 'mender-demo-artifact-3.1.0',
    device_types_compatible: ['qemux86-64'],
    started: '2022-07-10T15:30:00.000Z',
    status: 'artifact_uploaded'
  }
};

const mockDeltaJobDetails = {
  'delta-job-1': {
    ...deltaJobs['delta-job-1'],
    delta_artifact_size: 1024000,
    deployment_id: 'd1',
    devices_types_compatible: ['qemux86-64'],
    exit_code: 0,
    log: 'Delta generation started at 2022-07-11T20:49:00.000Z\nProcessing artifacts...\nDelta generation completed successfully at 2022-07-11T20:52:15.000Z',
    to_artifact_size: 2048000
  },
  'delta-job-2': {
    ...deltaJobs['delta-job-2'],
    delta_artifact_size: null,
    deployment_id: 'd2',
    devices_types_compatible: ['raspberrypi0w', 'raspberrypi0-wifi', 'raspberrypi3', 'raspberrypi4'],
    exit_code: 1,
    log: 'Delta generation started at 2022-07-11T20:49:00.000Z\nError: incompatible artifact types\nDelta generation failed at 2022-07-11T20:51:30.000Z',
    to_artifact_size: 3072000
  },
  'delta-job-3': {
    ...deltaJobs['delta-job-3'],
    delta_artifact_size: null,
    deployment_id: 'd3',
    devices_types_compatible: ['beaglebone', 'beaglebone-yocto'],
    exit_code: null,
    log: 'Delta generation started at 2022-07-11T20:49:00.000Z\nQueued for processing...',
    to_artifact_size: 2560000
  },
  'delta-job-4': {
    ...deltaJobs['delta-job-4'],
    delta_artifact_size: 512000,
    deployment_id: 'd1',
    devices_types_compatible: ['qemux86-64'],
    exit_code: 0,
    log: 'Delta generation started at 2022-07-10T15:30:00.000Z\nProcessing artifacts...\nArtifact uploaded successfully at 2022-07-10T15:45:20.000Z',
    to_artifact_size: 1536000
  }
};

export const releaseHandlers = [
  http.get(`${deploymentsApiUrl}/artifacts/:id`, () => new HttpResponse(null, { status: 200 })),
  http.get(`${deploymentsApiUrl}/artifacts/:id/download`, () => HttpResponse.json({ uri: 'https://testlocation.com/artifact.mender' })),
  http.delete(
    `${deploymentsApiUrl}/artifacts/:id`,
    ({ params: { id } }) => new HttpResponse(null, { status: id === defaultState.releases.byId.r1.artifacts[0].id ? 200 : 591 })
  ),
  http.put(`${deploymentsApiUrl}/artifacts/:id`, async ({ params: { id }, request }) => {
    const { description } = await request.json();
    return new HttpResponse(null, { status: id === defaultState.releases.byId.r1.artifacts[0].id && description ? 200 : 592 });
  }),
  http.post(
    `${deploymentsApiUrl}/artifacts/generate`,
    () =>
      new HttpResponse(null, {
        headers: { [headerNames.location]: `${deploymentsApiUrl}/artifacts/generate/${defaultState.releases.byId.r1.artifacts[0].id}`, status: 201 }
      })
  ),
  http.post(`${deploymentsApiUrl}/artifacts`, () => new HttpResponse(null, { status: 200 })),
  http.get(`${deploymentsApiUrlV2}/deployments/releases`, async ({ request }) => {
    const { searchParams } = new URL(request.url);
    const page = Number(searchParams.get('page'));
    const perPage = Number(searchParams.get('per_page'));
    if (!page || ![1, 10, 20, 50, 100, 250, 500].includes(perPage)) {
      return new HttpResponse(null, { status: 593 });
    }
    if (searchParams.get('device_type')) {
      return HttpResponse.json([]);
    }
    if (page == 42) {
      return new HttpResponse(JSON.stringify([defaultState.releases.byId.r1]), { headers: { [headerNames.total]: 1 } });
    }
    const sort = searchParams.get('sort');
    const releaseListSection = releasesList.sort(customSort(sort.includes(SORTING_OPTIONS.desc), 'name')).slice((page - 1) * perPage, page * perPage);
    if (page === 1 && perPage === 1 && searchParams.get('name')) {
      return HttpResponse.json([defaultState.releases.byId.r1]);
    }
    if (searchParams.get('name')) {
      return new HttpResponse(JSON.stringify(releaseListSection), { headers: { [headerNames.total]: 1234 } });
    }
    return new HttpResponse(JSON.stringify(releaseListSection), { headers: { [headerNames.total]: releasesList.length } });
  }),
  http.get(`${deploymentsApiUrlV2}/deployments/releases/:name`, () => HttpResponse.json(defaultState.releases.byId.r1)),
  http.get(`${deploymentsApiUrlV2}/releases/all/tags`, () => HttpResponse.json(['foo', 'bar'])),
  http.get(`${deploymentsApiUrlV2}/releases/all/types`, () => HttpResponse.json(['single-file', 'not-this'])),
  http.put(`${deploymentsApiUrlV2}/deployments/releases/:name/tags`, async ({ params: { name }, request }) => {
    const tags = await request.json();
    if (name && tags.every(i => i && i.toString() === i)) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 593 });
  }),
  http.patch(`${deploymentsApiUrlV2}/deployments/releases/:name`, async ({ params: { name }, request }) => {
    const { notes } = await request.json();
    return new HttpResponse(null, { status: name && notes.length ? 200 : 594 });
  }),
  http.get(`${deploymentsApiUrlV2}/deployments/releases/delta/jobs`, () => {
    const jobs = Object.values(deltaJobs);
    return new HttpResponse(JSON.stringify(jobs), { headers: { [headerNames.total]: jobs.length } });
  }),
  http.get(`${deploymentsApiUrlV2}/deployments/releases/delta/jobs/:jobId`, ({ params: { jobId } }) => {
    const jobDetails = mockDeltaJobDetails[jobId];
    if (jobDetails) {
      return HttpResponse.json(jobDetails);
    }
    return new HttpResponse(null, { status: 404 });
  })
];
