/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ArtifactInfo } from "./ArtifactInfo";
import type { DeviceStatus } from "./DeviceStatus";
import type { Update } from "./Update";
export type DeviceWithImage = {
  /**
   * Device identifier.
   */
  id: string;
  status: DeviceStatus;
  created?: string;
  started?: string;
  finished?: string;
  deleted?: string;
  device_type?: string;
  /**
   * Availability of the device's deployment log.
   */
  log: boolean;
  /**
   * State reported by device
   */
  state?: string;
  /**
   * Additional state information
   */
  substate?: string;
  image?: {
    /**
     * Image ID
     */
    id?: string;
    meta?: {
      /**
       * Image description
       */
      description?: string;
    };
    meta_artifact?: {
      name?: string;
      /**
       * An array of compatible device types.
       */
      device_types_compatible?: Array<string>;
      info?: ArtifactInfo;
      /**
       * Idicates if artifact is signed or not.
       */
      signed?: boolean;
      updates?: Array<Update>;
      /**
       * List of Artifact provides.
       *
       * Map of key/value pairs, where both keys and values are strings.
       */
      artifact_provides?: Record<string, string>;
      /**
       * List of Artifact depends.
       *
       * Map of key/value pairs, where keys are strings and values are lists of strings.
       */
      artifact_depends?: Record<string, Array<string>>;
      /**
       * List of Clear Artifact provides.
       */
      clears_artifact_provides?: Array<string>;
    };
    /**
     * Image size in bytes
     */
    size?: number;
    /**
     * Creation / last edition of any of the artifact properties
     */
    modified?: string;
  };
};
