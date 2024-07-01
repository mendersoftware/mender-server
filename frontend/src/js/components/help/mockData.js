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
export const helpProps = {
  docsVersion: undefined,
  isHosted: false,
  menderVersion: 'master',
  isEnterprise: true,
  tenantCapabilities: {
    hasDeviceConfig: true,
    hasMonitor: true
  },
  tokens: [],
  versions: {
    repos: {
      mender: '3.3.0',
      'mender-artifact': '3.8.0',
      'mender-binary-delta': '1.3.1',
      'mender-cli': '1.8.0',
      'mender-configure-module': '1.0.4',
      'mender-connect': '2.0.2',
      'mender-convert': '3.0.0',
      'mender-gateway': '1.0.0',
      'monitor-client': '1.2.0'
    },
    releaseDate: '2003-01-01'
  }
};
