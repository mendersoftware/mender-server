/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { DeploymentAggregationItem } from "./DeploymentAggregationItem";

export type DeploymentAggregation = {
  /**
   * Aggregation name
   */
  name?: string;
  items?: Array<DeploymentAggregationItem>;
  /**
   * Count of the documents not included in the items
   */
  other_count?: number;
};
