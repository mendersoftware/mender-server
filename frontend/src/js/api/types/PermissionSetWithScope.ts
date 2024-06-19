/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Scope } from "./Scope";
/**
 * Permission set with optional scope.
 */
export type PermissionSetWithScope = {
  /**
   * Unique permission set name.
   */
  name: string;
  scope?: Scope;
};
