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
