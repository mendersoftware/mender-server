/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { ArtifactInfo } from "./ArtifactInfo";
import type { Update } from "./Update";
/**
 * Detailed artifact.
 */
export type Artifact = {
  id: string;
  name: string;
  description: string;
  /**
   * An array of compatible device types.
   */
  device_types_compatible: Array<string>;
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
  /**
   * Artifact total size in bytes - the size of the actual file that will be transferred to the device (compressed).
   */
  size?: number;
  /**
   * Represents creation / last edition of any of the artifact properties.
   */
  modified: string;
};
