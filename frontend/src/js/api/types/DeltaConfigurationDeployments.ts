/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { BinaryDeltaConfiguration } from "./BinaryDeltaConfiguration";
import type { BinaryDeltaLimits } from "./BinaryDeltaLimits";
/**
 * Delta configuration options.
 */
export type DeltaConfigurationDeployments = {
  enabled?: boolean;
  binary_delta?: BinaryDeltaConfiguration;
  binary_delta_limits?: BinaryDeltaLimits;
};
