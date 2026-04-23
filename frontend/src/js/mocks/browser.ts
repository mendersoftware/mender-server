// Copyright 2026 Northern.tech AS
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
import { token } from '@northern.tech/testing/mockData';
import handlers from '@northern.tech/testing/requestHandlers/requestHandlers';
import { APPLICATION_JWT_CONTENT_TYPE, useradmApiUrl } from '@northern.tech/utils/constants';
import { HttpResponse, http } from 'msw';
import { setupWorker } from 'msw/browser';

export const worker = setupWorker(
  http.post(`${useradmApiUrl}/auth/login`, () => new HttpResponse(token, { headers: { 'Content-Type': APPLICATION_JWT_CONTENT_TYPE } })),
  ...handlers
);
