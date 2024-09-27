/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type DeploymentFilterTerm = {
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
  type: DeploymentFilterTerm.type;
};

export namespace DeploymentFilterTerm {
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
