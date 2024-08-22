/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Admission status of the device.
 */
export type StatusDeviceauth = {
  status: StatusDeviceauth.status;
};
export namespace StatusDeviceauth {
  export enum status {
    PENDING = "pending",
    ACCEPTED = "accepted",
    REJECTED = "rejected",
    PREAUTHORIZED = "preauthorized",
  }
}
