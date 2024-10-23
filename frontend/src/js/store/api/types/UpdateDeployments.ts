/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ArtifactTypeInfo } from "./ArtifactTypeInfo";
import type { UpdateFile } from "./UpdateFile";
/**
 * Single updated to be applied.
 */
export type UpdateDeployments = {
  type_info?: ArtifactTypeInfo;
  files?: Array<UpdateFile>;
  /**
   * meta_data is an array of objects of unknown structure as this
   * is dependent of update type (also custom defined by user)
   */
  meta_data?: Array<Record<string, any>>;
};
