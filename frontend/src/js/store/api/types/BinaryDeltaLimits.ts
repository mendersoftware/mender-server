/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Limit } from "./Limit";
import type { XDeltaArgsLimits } from "./XDeltaArgsLimits";
/**
 * The mender-binary-delta-generator configuration limits.
 */
export type BinaryDeltaLimits = {
  xdelta_args_limits?: XDeltaArgsLimits;
  timeout?: Limit;
  /**
   * Maximum number of delta generation jobs which can be run in parallel.
   */
  jobs_in_parallel?: Limit;
  /**
   * Maximum number of queued delta generation jobs.
   */
  queue_length?: Limit;
};
