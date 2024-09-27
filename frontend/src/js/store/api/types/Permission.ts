/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { PermissionObject } from "./PermissionObject";
export type Permission = {
  /**
   * Action
   */
  action: Permission.action;
  object: PermissionObject;
};
export namespace Permission {
  /**
   * Action
   */
  export enum action {
    ANY = "any",
    HTTP = "http",
  }
}
