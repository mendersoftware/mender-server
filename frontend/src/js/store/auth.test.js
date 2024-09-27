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
import { cleanUp, getToken, updateMaxAge } from './auth';

describe('auth functions', () => {
  it('getToken returns the jwt token if set in a cookie', async () => {
    expect(getToken()).toBeTruthy();
  });
  it('cleanup removes the JWT token', async () => {
    jest.clearAllMocks();
    cleanUp();
    expect(window.localStorage.removeItem).toHaveBeenCalledTimes(2);
  });
  it('updateMaxAge extends the expiration date of the jwt token', async () => {
    jest.clearAllMocks();
    updateMaxAge({ expiresAt: 'some-day', token: 'foo' });
    expect(window.localStorage.getItem).toHaveBeenCalledTimes(1);
    expect(window.localStorage.setItem).toHaveBeenCalledTimes(1);
  });
  it('updateMaxAge should keep any long expiration from the jwt cookie when the staying logged in setting is set', async () => {
    jest.clearAllMocks();
    updateMaxAge({ expiresAt: undefined, token: 'foo' });
    expect(window.localStorage.setItem).toHaveBeenCalledTimes(0);
  });
});
