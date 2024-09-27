/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeltaConfiguration } from "./DeltaConfiguration";
/**
 * New Tenant
 */
export type NewTenant = {
  /**
   * Name of the tenant.
   */
  name?: string;
  admin?: {
    /**
     * Email address of the admin user
     */
    email?: string;
    /**
     * Password of the admin user, must be provided if not using SSO
     */
    password?: string;
    /**
     * Alternative SSO login schemes, must be provided if password is empty
     */
    login?: Record<string, any>;
  };
  /**
   * Plan for the tenant, e.g.: os, professional, enterprise.
   */
  plan?: string;
  /**
   * Device limit for the tenant.
   */
  device_limit?: number;
  /**
   * List of add-ons to enable for the tenant, e.g.: troubleshoot, configure, monitor.
   */
  addons?: Array<string>;
  binary_delta?: DeltaConfiguration;
};
