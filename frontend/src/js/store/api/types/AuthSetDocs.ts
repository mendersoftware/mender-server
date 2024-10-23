/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * AuthSet describes the identity a device uses to authenticate with the Mender server.
 */
export type AuthSetDocs = {
  /**
   * The unique ID of the authentication set.
   */
  id?: string;
  /**
   * The unique ID of the device the authentication set belongs.
   */
  device_id?: string;
  /**
   * The identity data presented by the device.
   */
  identity_data?: Record<string, any>;
  /**
   * PEM-encoded public key of the device authentication set.
   */
  pubkey?: string;
  /**
   * Authentication status of the set.
   */
  status?: string;
  /**
   * The creation timestamp of the authentication set.
   */
  ts?: string;
};
