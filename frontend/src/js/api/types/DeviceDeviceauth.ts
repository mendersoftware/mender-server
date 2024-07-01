/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AuthSet } from "./AuthSet";
import type { ExternalIdentity } from "./ExternalIdentity";
import type { IdentityData } from "./IdentityData";
export type DeviceDeviceauth = {
  /**
   * Mender assigned Device ID.
   */
  id?: string;
  identity_data?: IdentityData;
  status?: DeviceDeviceauth.status;
  /**
   * Created timestamp
   */
  created_ts?: string;
  /**
   * Updated timestamp
   */
  updated_ts?: string;
  /**
   * Time when accepted device contacted server for the last time.
   */
  check_in_time?: string;
  auth_sets?: Array<AuthSet>;
  /**
   * Devices that are part of ongoing decommissioning process will return True
   */
  decommissioning?: boolean;
  external_id?: ExternalIdentity;
};
export namespace DeviceDeviceauth {
  export enum status {
    PENDING = "pending",
    ACCEPTED = "accepted",
    REJECTED = "rejected",
    PREAUTHORIZED = "preauthorized",
    NOAUTH = "noauth",
  }
}
