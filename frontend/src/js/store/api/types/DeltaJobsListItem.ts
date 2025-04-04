/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeltaJobsListItem = {
  /**
   * Identifier of the job
   */
  id?: string;
  /**
   * Workflows id that corresponds to the job executed by the generator, it is an internal id important
   * to include in case of support requests
   */
  delta_job_id?: string;
  /**
   * Release or artifact name of the second argument that we used to generate the delta
   */
  to_version?: string;
  /**
   * Release or artifact name of the first argument that we used to generate the delta
   */
  from_version?: string;
  /**
   * Array of the devices types names compatible with this artifact
   */
  devices_types_compatible?: Array<string>;
  /**
   * Date we started the generation
   */
  started?: string;
  /**
   * Gneration status
   */
  status?: DeltaJobsListItem.status;
};
export namespace DeltaJobsListItem {
  /**
   * Gneration status
   */
  export enum status {
    PENDING = "pending",
    QUEUED = "queued",
    SUCCESS = "success",
    FAILED = "failed",
    ARTIFACT_UPLOADED = "artifact_uploaded",
  }
}
