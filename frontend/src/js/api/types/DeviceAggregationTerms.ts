/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { DeviceAggregationTerm } from "./DeviceAggregationTerm";
import type { DeviceFilterTerm } from "./DeviceFilterTerm";
import type { GeoBoundingBoxFilter } from "./GeoBoundingBoxFilter";
import type { GeoDistanceFilter } from "./GeoDistanceFilter";

export type DeviceAggregationTerms = {
  /**
   * Aggregation terms.
   */
  aggregations?: Array<DeviceAggregationTerm>;
  /**
   * Filtering terms.
   */
  filters?: Array<DeviceFilterTerm>;
  geo_distance_filter?: GeoDistanceFilter;
  geo_bounding_box_filter?: GeoBoundingBoxFilter;
};
