/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Tag descriptor.
 */
export type Tag = {
  /**
   * Tag is an attribute with 'tags' scope.
   *
   * A human readable, unique tag ID, e.g. 'location', 'environment', etc.
   */
  name: string;
  /**
   * Tag description.
   */
  description?: string;
  /**
   * The current value of the tag.
   */
  value: string;
  /**
   * The date and time of last tag update in RFC3339 format.
   */
  timestamp?: string;
};
