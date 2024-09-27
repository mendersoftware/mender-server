// Copyright 2017 Northern.tech AS
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
import axios, { isCancel } from 'axios';

import { cleanUp, getToken } from '../auth';
import { TIMEOUTS } from '../constants';

const unauthorizedRedirect = error => {
  if (!isCancel(error) && error.response?.status === 401 && getToken()) {
    cleanUp();
    window.location.replace('/ui/');
  }
  return Promise.reject(error);
};

export const commonRequestConfig = { timeout: TIMEOUTS.refreshDefault, headers: { 'Content-Type': 'application/json' } };

export const authenticatedRequest = axios.create(commonRequestConfig);
authenticatedRequest.interceptors.response.use(res => res, unauthorizedRedirect);
authenticatedRequest.interceptors.request.use(
  config => ({ ...config, headers: { ...config.headers, Authorization: `Bearer ${getToken()}` } }),
  error => Promise.reject(error)
);

const Api = {
  get: authenticatedRequest.get,
  delete: (url, data) => authenticatedRequest.request({ method: 'delete', url, data }),
  patch: authenticatedRequest.patch,
  post: authenticatedRequest.post,
  postUnauthorized: (url, data, config = {}) => axios.post(url, data, { ...commonRequestConfig, ...config }),
  put: authenticatedRequest.put,
  upload: (url, formData, progress, cancelSignal) =>
    authenticatedRequest.post(url, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: progress,
      timeout: 0,
      signal: cancelSignal
    }),
  uploadPut: (url, formData, progress, cancelSignal) =>
    authenticatedRequest.put(url, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: progress,
      timeout: 0,
      signal: cancelSignal
    })
};

export default Api;
