/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Tenant account storage limit and storage usage.
 */
export type StorageLimit = {
  /**
   * Storage limit in bytes. If set to 0 - there is no limit for storage.
   */
  limit: number;
  /**
   * Current storage usage in bytes.
   */
  usage: number;
};
