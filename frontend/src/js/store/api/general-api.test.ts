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
import { describe, expect, it } from 'vitest';

import Api from './general-api';

const testLocation = '/test';

describe('General API module', () => {
  it('should allow GET requests', async () => {
    const res = await Api.get(testLocation);
    expect(res.config.headers.Authorization).toMatch(/Bearer/);
    expect(res.config.method).toBe('get');
  });
  it('should allow POST requests', async () => {
    const res = await Api.post(testLocation);
    expect(res.config.headers.Authorization).toMatch(/Bearer/);
    expect(res.config.method).toBe('post');
  });
  it('should allow PUT requests', async () => {
    const res = await Api.put(testLocation);
    expect(res.config.headers.Authorization).toMatch(/Bearer/);
    expect(res.config.method).toBe('put');
  });
  it('should allow DELETE requests', async () => {
    const res = await Api.delete(testLocation);
    expect(res.config.headers.Authorization).toMatch(/Bearer/);
    expect(res.config.method === 'del' || res.config.method === 'delete').toBe(true);
  });
});
