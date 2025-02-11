/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * DeviceInventoryEvent describes an event that relates to changes to a device's inventory data.
 */
export type DeviceInventoryEvent = {
  /**
   * Device unique ID.
   */
  device_id: string;
  /**
   * Tenant ID.
   */
  tenant_id: string;
  /**
   * Arbitrary key-value pairs of inventory attributes as device has sent them
   */
  inventory: Record<string, any>;
};
