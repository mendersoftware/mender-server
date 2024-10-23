/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Deployment } from "./Deployment";
import type { Device } from "./Device";
import type { Tenant } from "./Tenant";
import type { User } from "./User";
/**
 * Various types of objects are supported.
 * Depending on the type of object different information will be available.
 */
export type Object = {
  /**
   * An unique identifier of the object.
   */
  id: string;
  /**
   * The type of the object.
   */
  type: Object.type;
  tenant?: Tenant;
  user?: User;
  deployment?: Deployment;
  device?: Device;
};
export namespace Object {
  /**
   * The type of the object.
   */
  export enum type {
    TENANT = "tenant",
    USER = "user",
    DEPLOYMENT = "deployment",
    ARTIFACT = "artifact",
    DEVICE = "device",
  }
}
