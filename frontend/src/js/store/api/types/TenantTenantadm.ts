/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { TenantV1 } from "./TenantV1";
/**
 * Tenant descriptor.
 */
export type TenantTenantadm = TenantV1 & {
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
};
