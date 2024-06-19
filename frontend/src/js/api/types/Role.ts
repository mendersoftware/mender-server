/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { RolePermission } from "./RolePermission";
/**
 * Role descriptor.
 */
export type Role = {
  /**
   * A unique name.
   */
  name: string;
  /**
   * Description of the role, as shown in the UI.
   */
  description?: string;
  permissions?: Array<RolePermission>;
};
