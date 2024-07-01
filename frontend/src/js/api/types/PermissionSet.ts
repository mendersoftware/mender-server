/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Permission } from "./Permission";
/**
 * Permission set descriptor.
 */
export type PermissionSet = {
  /**
   * A unique name.
   */
  name: string;
  action?: string;
  object?: string;
  /**
   * Description of the permission set.
   */
  description?: string;
  permissions: Array<Permission>;
  supported_scope_types?: Array<string>;
};
