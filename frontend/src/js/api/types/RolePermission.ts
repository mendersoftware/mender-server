/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { RolePermissionObject } from "./RolePermissionObject";
/**
 * Role permission
 */
export type RolePermission = {
  /**
   * Action
   */
  action: RolePermission.action;
  object: RolePermissionObject;
};
export namespace RolePermission {
  /**
   * Action
   */
  export enum action {
    ANY = "any",
    HTTP = "http",
    CREATE_DEPLOYMENT = "CREATE_DEPLOYMENT",
  }
}
