/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Addon } from "./Addon";
import type { TenantApiLimits } from "./TenantApiLimits";
/**
 * Tenant descriptor.
 */
export type TenantTenantadm = {
  /**
   * Tenant ID.
   */
  id: string;
  /**
   * Name of the tenant's organization.
   */
  name: string;
  /**
   * Currently used tenant token.
   */
  tenant_token: string;
  /**
   * Status of the tenant account.
   */
  status?: TenantTenantadm.status;
  api_limits?: TenantApiLimits;
  addons?: Array<Addon>;
  /**
   * Count of accepted devices for the tenant.
   */
  device_count?: number;
  /**
   * Device limit for the tenant.
   */
  device_limit?: number;
  /**
   * Server side binary delta generation for the tenant is enabled.
   */
  binary_delta?: boolean;
  /**
   * Creation date and time, in ISO8601 format.
   */
  created_at?: string;
};
export namespace TenantTenantadm {
  /**
   * Status of the tenant account.
   */
  export enum status {
    ACTIVE = "active",
    SUSPENDED = "suspended",
  }
}
