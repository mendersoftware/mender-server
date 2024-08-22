/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Personal Access Token Object.
 */
export type PersonalAccessToken = {
  /**
   * Token identifier.
   */
  id: string;
  /**
   * Name of a token.
   */
  name: string;
  /**
   * Date of last usage of a token. The accuracy is 5 minutes.
   */
  last_used?: string;
  /**
   * Expiration date.
   */
  expiration_date: string;
  /**
   * Server-side timestamp of the token creation.
   */
  created_ts: string;
};
