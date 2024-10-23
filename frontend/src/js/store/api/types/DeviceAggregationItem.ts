/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeviceAggregation } from "./DeviceAggregation";
export type DeviceAggregationItem = {
  /**
   * Aggregation key
   */
  key?: string;
  /**
   * Aggregation count
   */
  count?: number;
  aggregations?: Array<DeviceAggregation>;
};
