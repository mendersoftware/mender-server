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
import React from 'react';

import { mdiAccountKey, mdiGithub, mdiGoogle, mdiMicrosoft } from '@mdi/js';
import MaterialDesignIcon from '@northern.tech/common-ui/materialdesignicon';

export const genericProvider = {
  id: 'generic',
  name: 'SSO provider',
  icon: <MaterialDesignIcon path={mdiAccountKey} />
};

export const OAuth2Providers = [
  {
    id: 'github',
    name: 'Github',
    icon: <MaterialDesignIcon path={mdiGithub} />
  },
  {
    id: 'google',
    name: 'Google',
    icon: <MaterialDesignIcon path={mdiGoogle} />
  },
  {
    id: 'microsoft',
    name: 'Microsoft',
    icon: <MaterialDesignIcon path={mdiMicrosoft} />
  }
];