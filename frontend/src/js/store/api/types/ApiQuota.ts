/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * API usage quota definition.
 */
export type ApiQuota = {
  /**
   * Maximum allowed number of calls within 'interval'.
   */
  max_calls?: number;
  /**
   * Interval definition, in seconds (60 = 1 minute, 3600 = 1 hour, etc.).
   */
  interval_sec?: number;
};
