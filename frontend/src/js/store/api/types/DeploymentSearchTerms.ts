/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeploymentAttributeProjection } from "./DeploymentAttributeProjection";
import type { DeploymentFilterTerm } from "./DeploymentFilterTerm";
import type { DeploymentSortTerm } from "./DeploymentSortTerm";
export type DeploymentSearchTerms = {
  /**
   * Pagination parameter for iterating search results.
   */
  page?: number;
  /**
   * Number of devices returned per page.
   */
  per_page?: number;
  /**
   * Filtering terms.
   */
  filters?: Array<DeploymentFilterTerm>;
  /**
   * Attribute keys to sort by.
   */
  sort?: Array<DeploymentSortTerm>;
  /**
   * Restrict the attribute result to the selected attributes.
   */
  attributes?: Array<DeploymentAttributeProjection>;
  /**
   * Restrict the result to the given device IDs.
   */
  device_ids?: Array<string>;
  /**
   * Restrict the result to the given deployment IDs.
   */
  deployment_ids?: Array<string>;
};
