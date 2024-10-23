/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Admission status of the device.
 */
export type Status = {
  status: Status.status;
};
export namespace Status {
  export enum status {
    PENDING = "pending",
    ACCEPTED = "accepted",
    REJECTED = "rejected",
    PREAUTHORIZED = "preauthorized",
  }
}
