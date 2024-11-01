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
  EXTERNAL_PROVIDER,
  PLANS,
  auditLogsApiUrl,
  headerNames,
  iotManagerBaseURL,
  ssoIdpApiUrlv1,
  tenantadmApiUrlv1,
  tenantadmApiUrlv2
} from '@northern.tech/store/constants';
import { HttpResponse, http } from 'msw';

import { defaultState, webhookEvents } from '../mockData';

const releasesSample = {
  lts: ['3.3'],
  releases: {
    '3.2': {
      '3.2.1': {
        release_date: '2022-02-02',
        release: '3.2.1',
        repos: [
          { name: 'integration', version: '1.2.3' },
          { name: 'mender', version: '3.2.1' },
          { name: 'mender-artifact', version: '1.3.7' },
          { name: 'service', version: '3.0.0' },
          { name: 'other-service', version: '1.1.0' }
        ]
      },
      '3.2.0': {
        release_date: '2022-01-24',
        release: '3.2.0',
        repos: [
          { name: 'service', version: '3.0.0' },
          { name: 'another-service', version: '4.1.0' }
        ]
      }
    },
    '2.7': {
      '2.7.2': {
        release_date: '2021-11-04',
        release: '2.7.2',
        repos: [
          { name: 'service', version: '1.1.0' },
          { name: 'more-service', version: '1.0.2' }
        ]
      }
    },
    '1.0': {
      supported_until: '2018-02',
      '1.0.1': {
        release_date: '2017-04-03',
        release: '1.0.1',
        repos: [{ name: 'service', version: '1.0.1' }]
      },
      '1.0.0': {
        release_date: '2017-02-16',
        release: '1.0.0',
        repos: [{ name: 'mono-service', version: '1.0.0' }]
      }
    }
  },
  saas: [
    { tag: 'saas-v2022.03.10', date: '2022-03-09' },
    { tag: 'saas-v2020.09.25', date: '2020-09-24' },
    { tag: 'saas-v2020.07.09', date: '2020-07-09' }
  ]
};

const tagsSample = [{ name: 'saas-v2023.05.02', more: 'here' }];

const signupHandler = async ({ request }) => {
  const signup = await request.json();
  if (['email', 'organization', 'plan', 'tos'].every(item => !!signup[item])) {
    return HttpResponse.text('test');
  }
  return new HttpResponse(null, { status: 400 });
};

export const organizationHandlers = [
  http.get('/tags.json', () => HttpResponse.json(tagsSample)),
  http.get('/versions.json', () => HttpResponse.json(releasesSample)),
  http.get(`${tenantadmApiUrlv1}/user/tenant`, () => HttpResponse.json(defaultState.organization.organization)),
  http.put(`${tenantadmApiUrlv2}/tenants/:tenantId/child`, () => new HttpResponse(null, { status: 200 })),
  http.post(`${tenantadmApiUrlv2}/tenants/:tenantId/cancel`, () => new HttpResponse(null, { status: 200 })),
  http.post(`${tenantadmApiUrlv2}/tenants/trial`, signupHandler),
  http.post(`${tenantadmApiUrlv2}/tenants`, () => new HttpResponse(null, { status: 200 })),
  http.post(`https://hosted.mender.io${tenantadmApiUrlv2}/tenants/trial`, signupHandler),
  http.get(`${tenantadmApiUrlv2}/tenants`, () => new HttpResponse([])),
  http.get(`${tenantadmApiUrlv2}/billing`, () => HttpResponse.json({ card: { last4: '7890', exp_month: 1, exp_year: 2024, brand: 'testCorp' } })),
  http.post(`${tenantadmApiUrlv2}/billing/card`, () => HttpResponse.json({ intent_id: defaultState.organization.intentId, secret: 'testSecret' })),
  http.post(
    `${tenantadmApiUrlv2}/billing/card/:intentId/confirm`,
    ({ params: { intentId } }) => new HttpResponse(null, { status: intentId == defaultState.organization.intentId ? 200 : 540 })
  ),
  http.post(`${tenantadmApiUrlv2}/tenants/:tenantId/upgrade/:status`, async ({ params: { status, tenantId }, request }) => {
    if (tenantId != defaultState.organization.organization.id || !['cancel', 'complete', 'start'].includes(status)) {
      return new HttpResponse(null, { status: 541 });
    }
    if (status === 'start') {
      return HttpResponse.json({ secret: 'testSecret' });
    }
    let plan;
    try {
      const body = await request.json();
      plan = body.plan;
    } catch {
      // no completion;
    }
    if (plan && !Object.keys(PLANS).includes(plan)) {
      return new HttpResponse(null, { status: 542 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.post(`${tenantadmApiUrlv2}/tenants/:tenantId/plan`, async ({ params: { tenantId }, request }) => {
    const body = await request.json();
    const expectedKeys = ['current_plan', 'requested_plan', 'current_addons', 'requested_addons', 'user_message'];
    if (tenantId != defaultState.organization.organization.id || !Object.keys(body).every(key => expectedKeys.includes(key))) {
      return new HttpResponse(null, { status: 544 });
    }
    if (body.requested_plan && !Object.values(PLANS).some(item => item.name === body.requested_plan)) {
      return new HttpResponse(null, { status: 545 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.post(`${tenantadmApiUrlv2}/contact/support`, async ({ request }) => {
    const { subject, body } = await request.json();
    if (!(subject && body)) {
      return new HttpResponse(null, { status: 543 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.get(`${auditLogsApiUrl}/logs`, ({ request }) => {
    const { searchParams } = new URL(request.url);
    const perPage = Number(searchParams.get('per_page'));
    if (perPage === 500) {
      return HttpResponse.json([
        { meta: defaultState.organization.auditlog.events[2].meta, time: defaultState.organization.auditlog.events[1].time, action: 'close_terminal' }
      ]);
    }
    return new HttpResponse(JSON.stringify(defaultState.organization.auditlog.events), {
      headers: { [headerNames.total]: defaultState.organization.auditlog.events.length }
    });
  }),
  http.get(`${auditLogsApiUrl}/logs/export`, () =>
    HttpResponse.text(`action,actor.id,actor.type,actor.email,actor.identity_data,object.id,object.type,object.user.email,object.deployment.name,object.deployment.artifact_name,change
    update,5c56c2ed-2a9a-5de9-bb86-cf38b3d4a5e1,user,test@example.coim,,067f23a9-76a5-5585-b119-32402a120978,user,test@example.com,,,"Update user 067f23a9-76a5-5585-b119-32402a120978 (test@example.com).
    Diff:
    --- Original
    +++ Current
    @@ -9 +9,2 @@
    -    ""RBAC_ROLE_OBSERVER""
    +    ""RBAC_ROLE_OBSERVER"",
    +    ""RBAC_ROLE_PERMIT_ALL""
    "
    `)
  ),
  http.get(`${iotManagerBaseURL}/integrations`, () =>
    HttpResponse.json([
      { connection_string: 'something_else', id: 1, provider: EXTERNAL_PROVIDER['iot-hub'].provider },
      { id: 2, provider: EXTERNAL_PROVIDER['iot-core'].provider, something: 'new' }
    ])
  ),
  http.post(`${iotManagerBaseURL}/integrations`, () =>
    HttpResponse.json([{ connection_string: 'something_else', provider: EXTERNAL_PROVIDER['iot-hub'].provider }])
  ),
  http.put(`${iotManagerBaseURL}/integrations/:integrationId`, ({ params: { integrationId } }) => {
    if (!integrationId) {
      return new HttpResponse(null, { status: 547 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.put(`${iotManagerBaseURL}/integrations/:integrationId/credentials`, ({ params: { integrationId } }) => {
    if (!integrationId) {
      return new HttpResponse(null, { status: 548 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.delete(`${iotManagerBaseURL}/integrations/:integrationId`, ({ params: { integrationId } }) => {
    if (!integrationId) {
      return new HttpResponse(null, { status: 549 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.get(`${iotManagerBaseURL}/events`, ({ request }) => {
    const { searchParams } = new URL(request.url);
    const page = Number(searchParams.get('page'));
    const perPage = Number(searchParams.get('per_page'));
    return HttpResponse.json(webhookEvents.slice(page - 1, page * perPage));
  }),
  http.get(ssoIdpApiUrlv1, () => {
    return HttpResponse.json([
      { id: '1', issuer: 'https://samltest.id/saml/idp', valid_until: '2038-08-24T21:14:09Z' },
      { id: '2', issuer: 'https://samltest2.id/saml/idp', valid_until: '2030-10-24T21:14:09Z' }
    ]);
  }),
  http.post(ssoIdpApiUrlv1, () => new HttpResponse(null, { status: 200 })),
  http.get(`${ssoIdpApiUrlv1}/:configId`, ({ params: { configId } }) => {
    if (!configId) {
      return new HttpResponse(null, { status: 550 });
    }
    return HttpResponse.json({ email: 'user@acme.com', password: 'mypass1234', login: { google: 'bob@gmail.com' }, config: '<div>not quite right</div>' });
  }),
  http.put(`${ssoIdpApiUrlv1}/:configId`, ({ params: { configId } }) => new HttpResponse(null, { status: configId ? 200 : 551 })),
  http.delete(`${ssoIdpApiUrlv1}/:configId`, ({ params: { configId } }) => new HttpResponse(null, { status: configId ? 200 : 552 }))
];
