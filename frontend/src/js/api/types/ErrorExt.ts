/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Error descriptor with additional metadata.
 */
export type ErrorExt = {
  /**
   * Description of the error.
   */
  error: string;
  /**
   * Request ID (same as in X-MEN-RequestID header).
   */
  request_id?: string;
  metadata?: Record<string, any>;
};
