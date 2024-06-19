/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Attribute } from "./Attribute";
export type DeviceInventory = {
  /**
   * Mender-assigned unique device ID.
   */
  id?: string;
  /**
   * Timestamp of the most recent attribute update.
   */
  updated_ts?: string;
  /**
   * A list of attribute descriptors.
   */
  attributes?: Array<Attribute>;
};
