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
import { AvailableAddon } from '@northern.tech/store/appSlice/constants';

import { Address } from '../api/types/Address';
import { ApiQuota } from '../api/types/ApiQuota';
import { Tenant } from '../api/types/Tenant';
import { SORTING_OPTIONS } from '../commonConstants';

//TODO: improve types
interface Card {
  brand: string;
  expiration: {
    month: number;
    year: number;
  };
  last4: string;
}

export interface SortOptions {
  direction: keyof typeof SORTING_OPTIONS;
  key?: string;
}

interface AuditLogSelectionState {
  detail?: string;
  endDate?: string;
  selectedIssue?: string;
  sort: SortOptions;
  startDate?: string;
  total: number;
  type?: string;
  user?: string;
}

interface AuditLog {
  events: Array<any>;
  selectionState: AuditLogSelectionState;
}

interface ExternalDeviceIntegration {
  connection_string: string;
  id: string;
  provider: string;
}

interface Webhook {
  events: Array<any>;
  eventsTotal: number;
}

interface TenantList {
  selectedTenant: Tenant | null;
  sort: SortOptions;
  tenants: Tenant[];
  total: number;
}

export interface OrganizationState {
  auditlog: AuditLog;
  card: Card;
  externalDeviceIntegrations: ExternalDeviceIntegration[];
  intentId: string | null;
  organization: Organization;
  ssoConfigs: any[];
  tenantList: TenantList;
  webhooks: Webhook;
}

interface ApiLimits {
  devices: {
    bursts: any[];
    quota: ApiQuota;
  };
  management: {
    bursts: any[];
    quota: ApiQuota;
  };
}

export interface Addon {
  enabled: boolean;
  name: AvailableAddon;
}

export interface BillingProfile {
  address: Address;
  email: string;
  name: string;
}

export interface Organization extends Tenant {
  addons: Addon[];
  api_limits: ApiLimits;
  created_at: string;
  id: string;
  name: string;
  status: 'active' | 'inactive';
  tenant_token: string;
}
