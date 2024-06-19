/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AlertSubject } from "./AlertSubject";
export type Alert = {
  /**
   * A unique ID for the alert.
   */
  id?: string;
  /**
   * The name of the alert.
   */
  name?: string;
  device_id: string;
  /**
   * Alert severity level
   */
  level?: Alert.level;
  subject: AlertSubject;
  /**
   * Time and date when the alert occurred
   */
  timestamp: string;
};
export namespace Alert {
  /**
   * Alert severity level
   */
  export enum level {
    OK = "OK",
    CRITICAL = "CRITICAL",
  }
}
