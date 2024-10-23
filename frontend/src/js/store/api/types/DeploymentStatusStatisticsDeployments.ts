/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeploymentStatusStatisticsDeployments = {
  /**
   * Number of successful deployments.
   */
  success: number;
  /**
   * Number of pending deployments.
   */
  pending: number;
  /**
   * Number of deployments being downloaded.
   */
  downloading: number;
  /**
   * Number of deployments devices are rebooting into.
   */
  rebooting: number;
  /**
   * Number of deployments devices being installed.
   */
  installing: number;
  /**
   * Number of failed deployments.
   */
  failure: number;
  /**
   * Do not have appropriate artifact for device type.
   */
  noartifact: number;
  /**
   * Number of devices unaffected by upgrade, since they are already running the specified software version.
   */
  "already-installed": number;
  /**
   * Number of deployments aborted by user.
   */
  aborted: number;
  /**
   * Number of deployments paused before install state.
   */
  pause_before_installing: number;
  /**
   * Number of deployments paused before reboot phase.
   */
  pause_before_rebooting: number;
  /**
   * Number of deployments paused before commit phase.
   */
  pause_before_committing: number;
};
