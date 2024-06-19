/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type NewConfigurationDeployment = {
  /**
   * The number of times a device can retry the deployment in case of failure, defaults to 0
   */
  retries?: number;
  /**
   * A valid JSON object defining the update control map.
   * *NOTE*: Available only in the Enterprise plan.
   */
  update_control_map?: Record<string, any>;
};
