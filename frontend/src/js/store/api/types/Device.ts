/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeviceStatus } from "./DeviceStatus";
export type Device = {
  /**
   * Device identifier.
   */
  id: string;
  status: DeviceStatus;
  created?: string;
  started?: string;
  finished?: string;
  deleted?: string;
  device_type?: string;
  /**
   * Availability of the device's deployment log.
   */
  log: boolean;
  /**
   * State reported by device
   */
  state?: string;
  /**
   * Additional state information
   */
  substate?: string;
};
