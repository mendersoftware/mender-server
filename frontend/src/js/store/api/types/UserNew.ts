/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { SsoObject } from "./SsoObject";
/**
 * New user descriptor.
 */
export type UserNew = {
  /**
   * A unique email address. Non-ascii characters are invalid.
   */
  email: string;
  /**
   * Password.
   */
  password?: string;
  /**
   * Alternative login schemes
   */
  login?: Record<string, any>;
  /**
   * SSO login schemes
   */
  sso?: Array<SsoObject>;
};
