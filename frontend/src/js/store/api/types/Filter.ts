/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { FilterPredicate } from "./FilterPredicate";
/**
 * Inventory filter assigned to the deployment
 */
export type Filter = {
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
