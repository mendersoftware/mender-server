/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Update user information.
 */
export type UserUpdate = {
  /**
   * A unique email address.
   */
  email?: string;
  /**
   * New password.
   */
  password?: string;
  /**
   * Current password.
   */
  current_password?: string;
  /**
   * List of roles for the user. If not provided existing roles are kept.
   */
  roles?: Array<string>;
};
