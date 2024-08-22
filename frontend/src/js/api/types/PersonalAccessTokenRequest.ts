/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Personal Access Token Request.
 */
export type PersonalAccessTokenRequest = {
  /**
   * Name of a token.
   */
  name: string;
  /**
   * Expiration time in seconds (maximum one year - 31536000s).
   * If you omit it or set it to zero, the Personal Access Token will never expire.
   */
  expires_in?: number;
};
