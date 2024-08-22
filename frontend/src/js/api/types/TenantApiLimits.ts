/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ApiLimits } from "./ApiLimits";
/**
 * API usage quota and burst limit definitions for a tenant - per API type.
 */
export type TenantApiLimits = {
  management?: ApiLimits;
  devices?: ApiLimits;
};
