/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Log in options
 */
export type LoginOptions = {
  /**
   * Generate a JWT token with no expiration date.
   */
  no_expiry?: boolean;
  /**
   * Two factor authentication token, required if two factor authentication is
   * enabled and tenant's plan is Professional or Enterprise.
   */
  token2fa?: string;
};
