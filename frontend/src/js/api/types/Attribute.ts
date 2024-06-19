/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Attribute descriptor.
 */
export type Attribute = {
  /**
   * A human readable, unique attribute ID, e.g. 'device_type', 'ip_addr', 'cpu_load', etc.
   */
  name: string;
  /**
   * The scope of the attribute.
   *
   * Scope is a string and acts as namespace for the attribute name.
   */
  scope: string;
  /**
   * Attribute description.
   */
  description?: string;
  /**
   * The current value of the attribute.
   *
   * Attribute type is implicit, inferred from the JSON type.
   *
   * Supported types: number, string, array of numbers, array of strings.
   * Mixed type arrays are not allowed.
   */
  value: string;
  /**
   * The date and time of last tag update in RFC3339 format.
   */
  timestamp?: string;
};
