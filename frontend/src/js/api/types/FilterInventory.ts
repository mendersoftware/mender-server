/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { FilterPredicate } from "./FilterPredicate";

/**
 * Filter definition
 */
export type FilterInventory = {
  /**
   * Unique identifier of the saved filter.
   */
  id: string;
  /**
   * Name of the saved filter.
   */
  name: string;
  terms?: Array<FilterPredicate>;
};
