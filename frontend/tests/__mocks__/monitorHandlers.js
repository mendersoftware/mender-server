// Copyright 2021 Northern.tech AS
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
import { alertChannels, headerNames, monitorApiUrlv1 } from '@northern.tech/store/constants';
import { HttpResponse, http } from 'msw';

import { defaultState } from '../mockData';

export const monitorHandlers = [
  http.get(`${monitorApiUrlv1}/devices/:id/alerts`, () => new HttpResponse(JSON.stringify([]), { headers: { [headerNames.total]: 1 } })),
  http.get(`${monitorApiUrlv1}/devices/:id/alerts/latest`, () => HttpResponse.json([])),
  http.get(`${monitorApiUrlv1}/devices/:id/config`, ({ params: { id } }) => {
    if (id === defaultState.devices.byId.a1.id) {
      return HttpResponse.json([{ something: 'here' }]);
    }
    return HttpResponse.json([]);
  }),
  http.put(`${monitorApiUrlv1}/settings/global/channel/alerts/:channel/status`, async ({ params: { channel }, request }) => {
    const body = await request.json();
    if (Object.keys(alertChannels).includes(channel) && body.hasOwnProperty('enabled')) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 590 });
  })
];
