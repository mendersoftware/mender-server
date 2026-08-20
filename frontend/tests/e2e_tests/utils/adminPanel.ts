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
import type { APIRequestContext } from '@playwright/test';

const apiPath = 'api/internal/v1/admin-panel';

/** the operator facing panel is its own single page app, served next to the tenant facing UI */
export const adminPanelUrl = (baseUrl: string, path = '') => `${baseUrl}admin/${path}`;

export const adminPanelApiUrl = (baseUrl: string, path: string) => `${baseUrl}${apiPath}/${path}`;

/**
 * The panel comes up as part of the enterprise compose setup, but is not routed on hosted
 * installations. Probing beats inferring from the environment alone, so the specs can gate on what
 * is actually deployed rather than on what the environment name suggests.
 */
export const isAdminPanelAvailable = async (request: APIRequestContext, baseUrl: string): Promise<boolean> => {
  try {
    const response = await request.get(adminPanelApiUrl(baseUrl, 'health'), { failOnStatusCode: false });
    return response.status() === 204;
  } catch {
    // an unrouted panel surfaces as a connection error rather than as a status code
    return false;
  }
};

export interface AdminPanelTenant {
  id: string;
  name: string;
  plan: string;
  service_provider?: boolean;
  status: string;
  trial: boolean;
}

export interface AdminPanelDeviceCounts {
  accepted: number;
  noauth: number;
  pending: number;
  preauthorized: number;
  rejected: number;
  total: number;
}

export interface AdminPanelTenantDetail extends AdminPanelTenant {
  artifact_size_limit_micro: number;
  device_counts?: AdminPanelDeviceCounts;
  warnings?: string[];
}

export interface AdminPanelUser {
  email: string;
  id: string;
  tenant_id?: string;
}

type QueryValue = boolean | number | string | undefined;
type TenantQuery = { page?: number; per_page?: number; plan?: string; q?: string; trial?: boolean };
type UserQuery = { email?: string; page?: number; per_page?: number; tenant_id?: string };

const toParams = (query: Record<string, QueryValue>) =>
  Object.fromEntries(
    Object.entries(query)
      .filter(([, value]) => value !== undefined)
      .map(([key, value]) => [key, String(value)])
  );

/** the tenant total is reported through `X-Total-Count` only - the payload itself is a plain array */
export const listTenants = async (request: APIRequestContext, baseUrl: string, query: TenantQuery = {}) => {
  const response = await request.get(adminPanelApiUrl(baseUrl, 'tenants'), { params: toParams(query) });
  if (!response.ok()) {
    throw new Error(`admin panel tenant listing failed with ${response.status()}`);
  }
  const tenants: AdminPanelTenant[] = await response.json();
  return { tenants, total: Number(response.headers()['x-total-count'] ?? 0) };
};

export const getTenantDetail = async (request: APIRequestContext, baseUrl: string, id: string): Promise<AdminPanelTenantDetail> => {
  const response = await request.get(adminPanelApiUrl(baseUrl, `tenants/${id}`));
  if (!response.ok()) {
    throw new Error(`admin panel tenant detail for ${id} failed with ${response.status()}`);
  }
  return response.json();
};

/** unlike the tenant listing the user store has no total to report, so there is no count header here */
export const listUsers = async (request: APIRequestContext, baseUrl: string, query: UserQuery = {}): Promise<AdminPanelUser[]> => {
  const response = await request.get(adminPanelApiUrl(baseUrl, 'users'), { params: toParams(query) });
  if (!response.ok()) {
    throw new Error(`admin panel user listing failed with ${response.status()}`);
  }
  return response.json();
};

export const tenantNames = { main: 'test', secondary: 'secondary' };

// the tenants the runner creates fit on a single page many times over
const tenantLookupPageSize = 100;

const findTenant = async (
  request: APIRequestContext,
  baseUrl: string,
  predicate: (tenant: AdminPanelTenant) => boolean,
  description: string
): Promise<AdminPanelTenant> => {
  const { tenants } = await listTenants(request, baseUrl, { per_page: tenantLookupPageSize });
  const tenant = tenants.find(predicate);
  if (!tenant) {
    throw new Error(`could not find the ${description} tenant among: ${tenants.map(({ name }) => name).join(', ') || '<none>'}`);
  }
  return tenant;
};

/** the tenant the suite runs its device & deployment tests against */
export const findMainTenant = (request: APIRequestContext, baseUrl: string) =>
  findTenant(request, baseUrl, ({ name }) => name === tenantNames.main, tenantNames.main);

/**
 * The runner creates two tenants named `secondary` - the service provider the tenant tests depend on
 * and a plain one. Anything that mutates tenant state has to stay on the plain one to avoid
 * knocking out the service provider coverage.
 */
export const findSecondaryTenant = (request: APIRequestContext, baseUrl: string) =>
  findTenant(request, baseUrl, tenant => tenant.name === tenantNames.secondary && !tenant.service_provider, 'plain secondary');
