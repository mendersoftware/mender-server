// Copyright 2024 Northern.tech AS
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
import { actions as appActions } from './appSlice';
import { actions as deploymentActions } from './deploymentsSlice';
import { actions as devicesActions } from './devicesSlice';
import { actions as monitorActions } from './monitorSlice';
import { actions as onboardingActions } from './onboardingSlice';
import { actions as organizationActions } from './organizationSlice';
import { actions as releaseActions } from './releasesSlice';
import { actions as userActions } from './usersSlice';

export default {
  ...appActions,
  ...deploymentActions,
  ...devicesActions,
  ...monitorActions,
  ...onboardingActions,
  ...organizationActions,
  ...releaseActions,
  ...userActions
};
