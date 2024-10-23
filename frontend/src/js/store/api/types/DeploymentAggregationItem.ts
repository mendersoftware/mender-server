/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeploymentAggregation } from "./DeploymentAggregation";
export type DeploymentAggregationItem = {
  /**
   * Aggregation key
   */
  key?: string;
  /**
   * Aggregation count
   */
  count?: number;
  aggregations?: Array<DeploymentAggregation>;
};
