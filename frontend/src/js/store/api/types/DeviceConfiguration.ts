/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ManagementAPIConfiguration } from "./ManagementAPIConfiguration";
export type DeviceConfiguration = {
  id?: string;
  configured?: ManagementAPIConfiguration;
  reported?: ManagementAPIConfiguration;
  /**
   * ID of the latest configuration deployment
   */
  deployment_id?: string;
  reported_ts?: string;
  updated_ts?: string;
};
