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
import type { Release } from '@northern.tech/store/releasesSlice';
import type { Device, Filter, NewDeploymentPhaseTypeManagement, NewDeploymentTypeManagement } from '@northern.tech/types/MenderTypes';

export type DeploymentSettings = Partial<{
  delta: boolean;
  deploymentDeviceCount: number;
  deploymentDeviceIds: string[];
  devices: Array<Device>;
  filter: Filter;
  forceDeploy: boolean;
  group: string;
  maxDevices: number;
  phases: Array<NewDeploymentPhaseTypeManagement>;
  release: Release;
  retries: number;
  update_control_map: NewDeploymentTypeManagement['update_control_map'];
}>;

export type DeploymentFormValues = Pick<DeploymentSettings, 'delta' | 'forceDeploy' | 'maxDevices' | 'retries' | 'phases' | 'update_control_map'> & {
  group: string | null;
  release: Release | null;
};
