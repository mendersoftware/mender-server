/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AlertDetails } from "./AlertDetails";
/**
 * Alert subject: the description of the alert origin
 */
export type AlertSubject = {
  /**
   * Name of an entity that caused the alert
   */
  name: string;
  /**
   * The type of executable that triggered the alert
   */
  type: string;
  /**
   * Status of an entity that caused the alert
   */
  status: string;
  details?: AlertDetails;
};
