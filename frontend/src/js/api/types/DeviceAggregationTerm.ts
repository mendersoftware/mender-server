/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type DeviceAggregationTerm = {
  /**
   * Name of the aggregation.
   */
  name: string;
  /**
   * Attribute key(s) to aggregate.
   */
  attribute?: string;
  /**
   * The scope the attribute(s) exists in.
   */
  scope?: string;
  /**
   * Number of top results to return.
   */
  limit?: number;
  /**
   * Sub-aggregation terms; it supports up to 5 nested subaggregations.
   */
  aggregations?: Array<DeviceAggregationTerm>;
};
