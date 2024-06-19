/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { XDeltaArgs } from "./XDeltaArgs";
/**
 * The mender-binary-delta-generator configuration options.
 */
export type BinaryDeltaConfiguration = {
  xdelta_args?: XDeltaArgs;
  /**
   * Delta generation job timeout in seconds.
   */
  timeout?: number;
};
