/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Artifact } from "./Artifact";
/**
 * Groups artifacts with the same release name into a single resource.
 */
export type ReleaseDeployments = {
  /**
   * release name.
   */
  name?: string;
  /**
   * Last modification time for the release.
   */
  modified?: string;
  /**
   * List of artifacts for this release.
   */
  artifacts?: Array<Artifact>;
  /**
   * Tags assigned to the release used for filtering releases. Each tag
   * must be valid a ASCII string and contain only lowercase and uppercase
   * letters, digits, underscores, periods and hyphens.
   */
  tags?: Array<string>;
  /**
   * Additional information describing a Release limited to 1024 characters. Please use the v2 API to set this field.
   */
  notes?: string;
};
