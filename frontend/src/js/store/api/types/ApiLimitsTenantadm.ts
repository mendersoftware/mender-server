/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ApiBurst } from "./ApiBurst";
import type { ApiQuota } from "./ApiQuota";
/**
 * Usage quota and burst limit definitions for an API type.
 */
export type ApiLimitsTenantadm = {
  /**
   * Collection of api burst limit definitions over specified API resources.
   */
  bursts?: Array<ApiBurst>;
  quota?: ApiQuota;
};
