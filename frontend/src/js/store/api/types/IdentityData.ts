/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Device identity attributes, in the form of a JSON structure.
 *
 * The attributes are completely vendor-specific, the provided ones are just an example.
 * In reference implementation structure contains vendor-selected fields,
 * such as MACs, serial numbers, etc.
 */
export type IdentityData = {
  /**
   * MAC address.
   */
  mac?: string;
  /**
   * Stock keeping unit.
   */
  sku?: string;
  /**
   * Serial number.
   */
  sn?: string;
};
