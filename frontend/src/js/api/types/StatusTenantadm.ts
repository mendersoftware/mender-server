/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Status of a tenant account.
 */
export type StatusTenantadm = {
  status: StatusTenantadm.status;
};

export namespace StatusTenantadm {
  export enum status {
    ACTIVE = "active",
    INACTIVE = "inactive",
    SUSPENDED = "suspended",
  }
}
