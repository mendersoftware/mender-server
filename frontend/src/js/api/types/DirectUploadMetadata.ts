/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Update } from "./Update";
/**
 * Artifact metadata
 */
export type DirectUploadMetadata = {
  /**
   * wsize of the artifact file.
   */
  size?: number;
  /**
   * List of updates for this artifact.
   */
  updates?: Array<Update>;
};
