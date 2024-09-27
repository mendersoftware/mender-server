/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeploymentStatusStatistics } from "./DeploymentStatusStatistics";
export type DeploymentStatistics = {
  status?: DeploymentStatusStatistics;
  /**
   * Sum of sizes (in bytes) of all artifacts assigned to all device deployments,
   * which are part of this deployment.
   * If the same artifact is assigned to multiple device deployments,
   * its size will be counted multiple times.
   */
  total_size?: number;
};
