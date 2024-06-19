/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { LineDescriptor } from "./LineDescriptor";
/**
 * Additional details on the alert.
 */
export type AlertDetails = {
  /**
   * Alert description.
   */
  description?: string;
  /**
   * Log lines before matched pattern.
   */
  lines_before?: Array<LineDescriptor>;
  /**
   * Log lines after matched pattern.
   */
  lines_after?: Array<LineDescriptor>;
  line_matching?: LineDescriptor;
};
