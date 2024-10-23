/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type Deployment = {
  /**
   * Device Deployment ID.
   */
  id?: string;
  tenant_id?: string;
  device_id?: string;
  deployment_id?: string;
  deployment_name?: string;
  deployment_artifact_name?: string;
  deployment_type?: string;
  deployment_created?: string;
  deployment_filter_id?: string;
  deployment_all_devices?: boolean;
  deployment_force_installation?: boolean;
  deployment_group?: string;
  deployment_phased?: boolean;
  deployment_phase_id?: string;
  deployment_retries?: number;
  deployment_max_devices?: number;
  deployment_autogenerate_deta?: boolean;
  device_created?: string;
  device_finished?: string;
  device_elapsed_seconds?: number;
  device_deleted?: string;
  device_status?: string;
  device_is_log_available?: boolean;
  device_retries?: number;
  device_attempts?: number;
  image_id?: string;
  image_description?: string;
  image_artifact_name?: string;
  image_device_types?: Array<string>;
  image_signed?: boolean;
  image_artifact_info_format?: string;
  image_artifact_info_version?: number;
  image_provides?: Record<string, any>;
  image_depends?: Record<string, any>;
  image_clear_provides?: string;
  image_size?: number;
};
