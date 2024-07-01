/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AuthSet } from "./AuthSet";
/**
 * DeviceAuthEvent describes an event that relates to changes to a device's authentication data. The properties included depends on the event type: device provisioning includes the entire device with the accepted authentication set, status change events only includes the device id and the new status, and device decommissioning will only include the device id.
 */
export type DeviceAuthEvent = {
  /**
   * Device unique ID.
   */
  id: string;
  /**
   * The authentication status of the device.
   */
  status?: string;
  auth_sets?: Array<AuthSet>;
  /**
   * The time the device was initialized in Mender.
   */
  created_ts?: string;
};
