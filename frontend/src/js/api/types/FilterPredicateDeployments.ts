/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Attribute filter predicate
 */
export type FilterPredicateDeployments = {
  /**
   * The scope of the attribute.
   *
   * Scope is a string and acts as namespace for the attribute name.
   */
  scope: string;
  /**
   * Name of the attribute to be queried for filtering.
   */
  attribute: string;
  /**
   * Type or operator of the filter predicate.
   */
  type: FilterPredicateDeployments.type;
  /**
   * The value of the attribute to be used in filtering.
   *
   * Attribute type is implicit, inferred from the JSON type.
   *
   * Supported types: number, string, array of numbers, array of strings.
   * Mixed arrays are not allowed.
   */
  value: string;
};
export namespace FilterPredicateDeployments {
  /**
   * Type or operator of the filter predicate.
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
  }
}
