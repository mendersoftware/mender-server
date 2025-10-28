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
import { Tenant as APITenant, UpgradeCompleteRequest } from '@northern.tech/store/api/types';

//TODO: rely on API tenant directly once type generation fixed
export interface Tenant extends APITenant {
  additional_info: {
    campaign: string;
    marketing: boolean;
  };
  binary_delta: boolean;
  cancelled_at: string | null;
  children_tenants: any[] | null;
  device_count: number;
  device_limit: number;
  max_child_tenants: number;
  parent_tenant_id: string;
  plan: UpgradeCompleteRequest.plan;
  service_provider: boolean;
  trial: boolean;
  trial_expiration: string | null;
}
