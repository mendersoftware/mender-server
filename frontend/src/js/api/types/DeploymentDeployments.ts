/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { DeploymentPhase } from "./DeploymentPhase";
import type { DeploymentStatistics } from "./DeploymentStatistics";
import type { Filter } from "./Filter";

export type DeploymentDeployments = {
  /**
   * Deployment identifier
   */
  id: string;
  /**
   * Name of the deployment
   */
  name: string;
  /**
   * Name of the artifact to deploy
   */
  artifact_name: string;
  /**
   * Deployment's creation date and time
   */
  created: string;
  /**
   * Deployment's completion date and time
   */
  finished?: string;
  /**
   * Status of the deployment
   */
  status: DeploymentDeployments.status;
  /**
   * Number of devices the deployment acted upon
   */
  device_count: number;
  /**
   * An array of artifact's identifiers.
   */
  artifacts?: Array<string>;
  /**
   * An array of groups the devices targeted by the deployment belong to.
   * Available only if the user created the deployment for a group or a single device (if the device was in a static group).
   */
  groups?: Array<string>;
  phases?: Array<DeploymentPhase>;
  /**
   * The number of times a device can retry the deployment in case of failure, defaults to 0
   */
  retries?: number;
  /**
   * A valid JSON object defining the update control map.
   * *NOTE*: Available only in the Enterprise plan.
   */
  update_control_map?: Record<string, any>;
  /**
   * max_devices denotes a limit on a number of completed deployments (failed or successful) above which the dynamic deployment will be finished.
   */
  max_devices?: number;
  /**
   * In case of dynamic deployments this is a number of devices targeted initially (maching the filter at the moment of deployment creation).
   */
  initial_device_count?: number;
  /**
   * Flag indicating if the deployment is dynamic or not.
   */
  dynamic?: boolean;
  filter?: Filter;
  type?: DeploymentDeployments.type;
  /**
   * A string containing a configuration object provided
   * with the deployment constructor.
   */
  configuration?: string;
  /**
   * The flag idicating if the autogeneration of delta artifacts is enabled for a given deployment.
   */
  autogenerate_delta?: boolean;
  statistics?: DeploymentStatistics;
};

export namespace DeploymentDeployments {
  /**
   * Status of the deployment
   */
  export enum status {
    SCHEDULED = "scheduled",
    PENDING = "pending",
    INPROGRESS = "inprogress",
    FINISHED = "finished",
  }

  export enum type {
    CONFIGURATION = "configuration",
    SOFTWARE = "software",
  }
}
