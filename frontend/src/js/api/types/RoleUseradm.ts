/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { PermissionSetWithScope } from "./PermissionSetWithScope";
/**
 * Role descriptor.
 */
export type RoleUseradm = {
  /**
   * A unique name.
   */
  name: string;
  /**
   * Description of the role, as shown in the UI.
   */
  description?: string;
  permission_sets_with_scope?: Array<PermissionSetWithScope>;
};
