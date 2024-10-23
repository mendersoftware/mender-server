/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeviceFilterTerm = {
  /**
   * Attribute key to compare.
   */
  attribute: string;
  /**
   * Filter matching expression.
   */
  value: any;
  /**
   * Type of filtering operation.
   */
  type: DeviceFilterTerm.type;
  /**
   * The scope the attribute exists in.
   */
  scope?: string;
};
export namespace DeviceFilterTerm {
  /**
   * Type of filtering operation.
   */
  export enum type {
    _EQ = "$eq",
    _GT = "$gt",
    _GTE = "$gte",
    _IN = "$in",
    _LT = "$lt",
    _LTE = "$lte",
    _NE = "$ne",
    _NIN = "$nin",
    _EXISTS = "$exists",
    _REGEX = "$regex",
  }
}
