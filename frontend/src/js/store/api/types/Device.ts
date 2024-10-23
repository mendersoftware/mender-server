/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeviceAttribute } from "./DeviceAttribute";
export type Device = {
  /**
   * Device ID.
   */
  id?: string;
  /**
   * Last device check-in itme.
   */
  check_in_time?: string;
  attributes?: Array<DeviceAttribute>;
  /**
   * Timestamp of the last update to the device attributes.
   */
  updated_ts?: string;
};
