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
import { APPLICATION_JWT_CONTENT_TYPE, useradmApiUrl, useradmApiUrlv2 } from '@northern.tech/store/constants';
import { HttpResponse, http } from 'msw';

import { accessTokens, defaultPassword, defaultState, userId as defaultUserId, permissionSets, rbacRoles, testSsoId, token } from '../mockData';

export const userHandlers = [
  http.post(`${useradmApiUrl}/auth/login`, ({ request }) => {
    const authHeader = request.headers.get('authorization');
    const authInfo = atob(authHeader?.split(' ')[1]);
    const [user, password] = authInfo.split(':');

    if (!password) {
      return HttpResponse.json({ id: testSsoId, kind: 'sso/saml' });
    } else if (password !== defaultPassword) {
      return new HttpResponse(null, { status: 401 });
    } else if (user.includes('limited')) {
      return new HttpResponse('limitedToken', { headers: { 'Content-Type': APPLICATION_JWT_CONTENT_TYPE } });
    } else if (user.includes('2fa')) {
      return new HttpResponse(JSON.stringify({ error: '2fa needed' }), { status: 401 });
    }

    return new HttpResponse(token, { headers: { 'Content-Type': APPLICATION_JWT_CONTENT_TYPE } });
  }),
  http.get(`https://hosted.mender.io${useradmApiUrl}/auth/magic/:id`, ({ params: { id } }) => {
    if (id) {
      return new HttpResponse('test', { headers: { 'Set-Cookie': 'JWT=test' } });
    }
    return new HttpResponse(null, { status: 400 });
  }),
  http.post(`${useradmApiUrl}/auth/logout`, () => new HttpResponse(null, { status: 200 })),
  http.post(`${useradmApiUrl}/auth/password-reset/:status`, async ({ params: { status }, request }) => {
    const { email, secret_hash, password } = await request.json();
    if (!['start', 'complete'].includes(status) && ![email, secret_hash, password].some(item => item)) {
      return new HttpResponse(null, { status: 560 });
    }
    if (status === 'start' && email) {
      return new HttpResponse(null, { status: 200 });
    }
    if (status === 'complete' && secret_hash && password) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 561 });
  }),
  http.put(`${useradmApiUrl}/2faverify`, async ({ request }) => {
    const { token2fa } = await request.json();
    if (!token2fa) {
      return new HttpResponse(null, { status: 562 });
    }
    return new HttpResponse(null, { status: 200 });
  }),
  http.get(`${useradmApiUrl}/users`, () => HttpResponse.json(Object.values(defaultState.users.byId))),
  http.get(`${useradmApiUrl}/users/me`, ({ request }) => {
    const authHeader = request.headers.get('authorization');
    if (authHeader?.includes('limited')) {
      return new HttpResponse(JSON.stringify({ error: 'forbidden by role-based access control' }), { status: 403 });
    }
    return HttpResponse.json(defaultState.users.byId[defaultUserId]);
  }),
  http.get(`${useradmApiUrl}/users/:userId`, ({ params: { userId } }) => {
    if (userId === 'me' || defaultState.users.byId[userId]) {
      const user = userId === 'me' ? defaultUserId : userId;
      return HttpResponse.json(defaultState.users.byId[user]);
    }
    return new HttpResponse(null, { status: 563 });
  }),
  http.post(`${useradmApiUrl}/users`, async ({ request }) => {
    const { email, password } = await request.json();
    if (email === 'test@test.com' || [email, password].every(value => value)) {
      return HttpResponse.json(defaultState.users.byId.a1);
    }
    return new HttpResponse(null, { status: 564 });
  }),
  http.put(`${useradmApiUrl}/users/:userId`, async ({ params: { userId }, request }) => {
    const { email, password } = await request.json();
    if (
      (defaultState.users.byId[userId] && [email, password].some(value => Object.values(defaultState.users.byId[userId]).includes(value))) ||
      email === 'test@test.com'
    ) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 565 });
  }),
  http.delete(`${useradmApiUrl}/users/:userId`, ({ params: { userId } }) => new HttpResponse(null, { status: defaultState.users.byId[userId] ? 200 : 566 })),
  http.get(`${useradmApiUrl}/roles`, () => HttpResponse.json(rbacRoles)),
  http.post(`${useradmApiUrl}/roles`, async ({ request }) => {
    const { name, permissions } = await request.json();
    if (
      [name, permissions].every(value => value) &&
      permissions.every(permission => permission.action && permission.object && permission.object.type && permission.object.value)
    ) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 567 });
  }),
  http.put(`${useradmApiUrl}/roles/:roleId`, async ({ params: { roleId }, request }) => {
    const { description, name, permissions } = await request.json();
    if (defaultState.users.rolesById[roleId] && [description, name, permissions].some(value => value)) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 568 });
  }),
  http.delete(
    `${useradmApiUrl}/roles/:roleId`,
    ({ params: { roleId } }) => new HttpResponse(null, { status: defaultState.users.rolesById[roleId] ? 200 : 569 })
  ),
  http.get(`${useradmApiUrlv2}/roles`, () => HttpResponse.json(rbacRoles)),
  http.get(`${useradmApiUrlv2}/roles/:roleId`, async ({ params: { roleId } }) => {
    if (defaultState.users.rolesById[roleId]) {
      return HttpResponse.json(defaultState.users.rolesById[roleId]);
    }
    return new HttpResponse(null, { status: 571 });
  }),
  http.post(`${useradmApiUrlv2}/roles`, async ({ request }) => {
    const { name, permission_sets_with_scope } = await request.json();
    if (
      !!name &&
      permission_sets_with_scope.every(permission => permission.name && permissionSets.find(permissionSet => permissionSet.name === permission.name))
    ) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 572 });
  }),
  http.put(`${useradmApiUrlv2}/roles/:roleId`, async ({ params: { roleId }, request }) => {
    const { description, name, permission_sets_with_scope } = await request.json();
    if (defaultState.users.rolesById[roleId] && [description, name, permission_sets_with_scope].some(value => value)) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 573 });
  }),
  http.delete(`${useradmApiUrlv2}/roles/:roleId`, ({ params: { roleId } }) => {
    if (defaultState.users.rolesById[roleId]) {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 574 });
  }),
  http.get(`${useradmApiUrl}/settings`, () => HttpResponse.json(defaultState.users.globalSettings)),
  http.post(`${useradmApiUrl}/settings`, () => new HttpResponse(null, { status: 200 })),
  http.get(`${useradmApiUrl}/settings/me`, () => HttpResponse.json(defaultState.users.userSettings)),
  http.post(`${useradmApiUrl}/settings/me`, () => new HttpResponse(null, { status: 200 })),
  http.get(`${useradmApiUrl}/settings/tokens`, () => HttpResponse.json(accessTokens)),
  http.post(`${useradmApiUrl}/settings/tokens`, () => HttpResponse.json('aNewToken')),
  http.delete(
    `${useradmApiUrl}/settings/tokens/:tokenId`,
    ({ params: { tokenId } }) => new HttpResponse(null, { status: tokenId === 'some-id-1' ? 200 : 577 })
  ),
  http.get(`${useradmApiUrl}/2faqr`, () => HttpResponse.json({ qr: btoa('test') })),
  http.post(`${useradmApiUrl}/users/:userId/2fa/enable`, ({ params: { userId } }) => {
    if (defaultState.users.byId[userId] || userId === 'me') {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 570 });
  }),
  http.post(`${useradmApiUrl}/users/:userId/2fa/disable`, ({ params: { userId } }) => {
    if (defaultState.users.byId[userId] || userId === 'me') {
      return new HttpResponse(null, { status: 200 });
    }
    return new HttpResponse(null, { status: 571 });
  }),
  http.post(`${useradmApiUrl}/auth/verify-email/start`, () => new HttpResponse(null, { status: 200 })),
  http.post(`${useradmApiUrl}/auth/verify-email/complete`, async ({ request }) => {
    const { secret_hash } = await request.json();
    return new HttpResponse(null, { status: secret_hash === 'superSecret' ? 200 : 576 });
  }),
  http.get(`${useradmApiUrlv2}/permission_sets`, () => HttpResponse.json(permissionSets)),
  http.post(`${useradmApiUrl}/users/:userId/assign`, async ({ request }) => {
    const { tenant_ids } = await request.json();
    return new HttpResponse(null, { status: tenant_ids.length ? 200 : 577 });
  }),
  http.get(`${useradmApiUrl}/users/tenants/:tenantId/token`, () => HttpResponse.text('differentToken'))
];
