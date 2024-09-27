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
import Api from './general-api';

const testLocation = '/test';

describe('General API module', () => {
  it('should allow GET requests', done => {
    Api.get(testLocation)
      .then(res => {
        expect(res.config.headers.Authorization).toMatch(/Bearer/);
        return res.config.method === 'get' ? done() : done('failed');
      })
      .catch(done);
  });
  it('should allow POST requests', done => {
    Api.post(testLocation)
      .then(res => {
        expect(res.config.headers.Authorization).toMatch(/Bearer/);
        return res.config.method === 'post' ? done() : done('failed');
      })
      .catch(done);
  });
  it('should allow PUT requests', done => {
    Api.put(testLocation)
      .then(res => {
        expect(res.config.headers.Authorization).toMatch(/Bearer/);
        return res.config.method === 'put' ? done() : done('failed');
      })
      .catch(done);
  });
  it('should allow DELETE requests', done => {
    Api.delete(testLocation)
      .then(res => {
        expect(res.config.headers.Authorization).toMatch(/Bearer/);
        return res.config.method === 'del' || res.config.method === 'delete' ? done() : done('failed');
      })
      .catch(done);
  });
});
