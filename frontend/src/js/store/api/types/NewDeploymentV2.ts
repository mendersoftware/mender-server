/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { NewDeploymentPhase } from "./NewDeploymentPhase";
export type NewDeploymentV2 = {
  name: string;
  artifact_name: string;
  /**
   * ID of a filter from inventory service.
   */
  filter_id: string;
  /**
   * Phased rollout feature is available only to Enterprise users.
   */
  phases?: Array<NewDeploymentPhase>;
  /**
   * The number of times a device can retry the deployment in case of failure, defaults to 0
   */
  retries?: number;
  /**
   * max_devices denotes a limit on a number of completed deployments (failed or successful) above which the dynamic deployment will be finished
   */
  max_devices?: number;
  /**
   * A valid JSON object defining the update control map.
   * *NOTE*: Available only in the Enterprise plan.
   */
  update_control_map?: Record<string, any>;
  /**
   * The flag idicating if the autogeneration of delta artifacts is enabled for a given deployment.
   */
  autogenerate_delta?: boolean;
};
