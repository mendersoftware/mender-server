// Copyright 2016 Northern.tech AS
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
import axios from 'axios';

import { getToken } from '../auth';
import { commonRequestConfig } from './general-api';

const Api = {
  postLogin: (url, { email: username, password, ...body }) =>
    axios
      .post(url, body, { ...commonRequestConfig, auth: { username, password } })
      .then(res => ({ text: res.data, code: res.status, contentType: res.headers?.['content-type'] })),
  putVerifyTFA: (url, userData) => {
    let body = {};
    if (userData.hasOwnProperty('token2fa')) {
      body = { token2fa: userData.token2fa };
    }
    return axios.put(url, body, { ...commonRequestConfig, headers: { ...commonRequestConfig.headers, Authorization: `Bearer ${getToken()}` } });
  }
};

export default Api;
