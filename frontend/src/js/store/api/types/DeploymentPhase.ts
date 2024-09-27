/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeploymentPhase = {
  /**
   * Phase identifier.
   */
  id?: string;
  /**
   * Percentage of devices to update in the phase.
   */
  batch_size?: number;
  /**
   * Start date of a phase.
   * May be undefined for the first phase of a deployment.
   */
  start_ts?: string;
  /**
   * Number of devices which already requested an update within this phase.
   */
  device_count?: number;
};
