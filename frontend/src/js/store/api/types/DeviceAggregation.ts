/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { DeviceAggregationItem } from "./DeviceAggregationItem";

export type DeviceAggregation = {
  /**
   * Aggregation name
   */
  name?: string;
  items?: Array<DeviceAggregationItem>;
  /**
   * Count of the documents not included in the items
   */
  other_count?: number;
};
