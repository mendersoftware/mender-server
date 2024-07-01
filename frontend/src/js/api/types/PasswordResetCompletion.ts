/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Password reset completion
 */
export type PasswordResetCompletion = {
  /**
   * Secret hash received by email by the user
   */
  secret: string;
  /**
   * New password of the user
   */
  password: string;
};
