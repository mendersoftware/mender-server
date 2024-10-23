/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeploymentAggregationTerm = {
  /**
   * Name of the aggregation.
   */
  name: string;
  /**
   * Attribute key(s) to aggregate.
   */
  attribute?: string;
  /**
   * Number of top results to return.
   */
  limit?: number;
  /**
   * Sub-aggregation terms; it supports up to 5 nested subaggregations.
   */
  aggregations?: Array<DeploymentAggregationTerm>;
};
