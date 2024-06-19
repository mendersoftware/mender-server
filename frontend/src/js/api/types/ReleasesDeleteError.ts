/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Releases used by active deployment.
 */
export type ReleasesDeleteError = {
  /**
   * Description of the error.
   */
  error?: string;
  /**
   * List of IDs of active deployments which are using releases from the request.
   */
  active_deployments?: Array<string>;
  /**
   * Request ID (same as in X-MEN-RequestID header).
   */
  request_id?: string;
};
