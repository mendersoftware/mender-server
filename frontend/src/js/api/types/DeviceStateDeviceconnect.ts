/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type DeviceStateDeviceconnect = {
  /**
   * Device ID.
   */
  device_id?: string;
  /**
   * Device status.
   */
  status?: DeviceStateDeviceconnect.status;
  /**
   * Server-side timestamp of the last device information update.
   */
  updated_ts?: string;
  /**
   * Server-side timestamp of the device creation.
   */
  created_ts?: string;
};

export namespace DeviceStateDeviceconnect {
  /**
   * Device status.
   */
  export enum status {
    CONNECTED = "connected",
    DISCONNECTED = "disconnected",
    UNKNOWN = "unknown",
  }
}
