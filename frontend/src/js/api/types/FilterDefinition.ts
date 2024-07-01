/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { FilterPredicate } from "./FilterPredicate";
/**
 * Filter definition
 */
export type FilterDefinition = {
  /**
   * Name of the filter, must be unique.
   */
  name?: string;
  /**
   * List of filter predicates, chained with boolean AND operators to build the search condition definition.
   */
  terms?: Array<FilterPredicate>;
};
