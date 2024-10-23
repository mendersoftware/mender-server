/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
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
  users?: Array<{
    /**
     * Email address of an existing user to be added to the newly created tenant
     */
    email?: string;
    /**
     * Role of the user to be added
     */
    role?: string;
  }>;
  /**
   * Device limit for the tenant.
   */
  device_limit?: number;
  /**
   * Enable server side binary delta generation for the tenant.
   */
  binary_delta?: boolean;
  /**
   * Enable SSO for the tenant.
   */
  sso?: boolean;
};
