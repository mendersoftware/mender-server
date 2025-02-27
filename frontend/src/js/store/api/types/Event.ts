/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DeviceAuthEvent } from "./DeviceAuthEvent";
import type { DeviceInventoryEvent } from "./DeviceInventoryEvent";
export type Event = {
  /**
   * A unique event identifier generated by the Mender server
   */
  id?: string;
  /**
   * Type of the event
   */
  type?: Event.type;
  delivery_statuses?: Array<{
    /**
     * The ID of the integration the status belongs.
     */
    integration_id: string;
    /**
     * Whether the event hook was executed successfully.
     */
    success: boolean;
    /**
     * The (HTTP) status code of the hook.
     */
    status_code?: number;
    /**
     * An error message if the hook failed.
     */
    error?: string;
  }>;
  /**
   * Creation timestamp
   */
  time?: string;
  data?: DeviceAuthEvent | DeviceInventoryEvent;
};
export namespace Event {
  /**
   * Type of the event
   */
  export enum type {
    DEVICE_PROVISIONED = "device-provisioned",
    DEVICE_DECOMMISSIONED = "device-decommissioned",
    DEVICE_STATUS_CHANGED = "device-status-changed",
    DEVICE_INVENTORY_CHANGED = "device-inventory-changed",
  }
}
