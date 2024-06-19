/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { NewDeploymentPhase } from "./NewDeploymentPhase";
export type NewDeployment = {
  /**
   * Name of the deployment
   */
  name: string;
  /**
   * Name of the artifact to deploy
   */
  artifact_name: string;
  /**
   * An array of devices' identifiers.
   */
  devices?: Array<string>;
  /**
   * When set, the deployment will be created for all
   * currently accepted devices.
   */
  all_devices?: boolean;
  /**
   * Force the installation of the Artifact disabling the `already-installed` check.
   */
  force_installation?: boolean;
  phases?: Array<NewDeploymentPhase>;
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
   * The flag idicating if the autogeneration of delta artifacts is enabled for a given deployment.
   */
  autogenerate_delta?: boolean;
};
