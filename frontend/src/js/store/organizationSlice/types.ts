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
import { AvailableAddon, AvailablePlans } from '@northern.tech/store/appSlice/constants';

//TODO: improve types
interface Card {
  last4: string;
  expiration: {
    month: number;
    year: number;
  };
  brand: string;
}

interface SortOptions {
  direction: 'asc' | 'desc';
  key?: string;
}

interface Tenant {
  id: string;
  name: string;
}

interface AuditLogSelectionState {
  total: number;
  startDate?: string;
  endDate?: string;
  detail?: string;
  selectedIssue?: string;
  type?: string;
  user?: string;
  sort: SortOptions;
}

interface AuditLog {
  events: Array<any>;
  selectionState: AuditLogSelectionState;
}

interface ExternalDeviceIntegration {
  id: string;
  provider: string;
  connection_string: string;
}

interface Webhook {
  events: Array<any>;
  eventsTotal: number;
}

interface TenantList {
  total: number;
  tenants: Tenant[];
  selectedTenant: Tenant | null;
  sort: SortOptions;
}

export interface OrganizationState {
  card: Card;
  intentId: string | null;
  tenantList: TenantList;
  organization: Organization;
  auditlog: AuditLog;
  externalDeviceIntegrations: ExternalDeviceIntegration[];
  ssoConfigs: any[];
  webhooks: Webhook;
}
interface ApiQuota {
  max_calls: number;
  interval_sec: number;
}

interface ApiLimits {
  management: {
    bursts: any[];
    quota: ApiQuota;
  };
  devices: {
    bursts: any[];
    quota: ApiQuota;
  };
}

export interface Addon {
  name: AvailableAddon;
  enabled: boolean;
}
export interface BillingProfile {
  email: string;
  name: string;
  address: { country: string; state: string; city: string; line1: string; postal_code: string };
}

export interface Organization {
  id: string;
  parent_tenant_id: string;
  name: string;
  tenant_token: string;
  status: 'active' | 'inactive';
  additional_info: {
    marketing: boolean;
    campaign: string;
  };
  plan: AvailablePlans;
  api_limits: ApiLimits;
  trial: boolean;
  trial_expiration: string | null;
  service_provider: boolean;
  created_at: string;
  cancelled_at: string | null;
  addons: Addon[];
  max_child_tenants: number;
  children_tenants: any | null;
  device_count: number;
  device_limit: number;
  binary_delta: boolean;
}
