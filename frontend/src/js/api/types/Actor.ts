/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * The actor may be a user or device.
 * Depending on the type of the actor different information will be available.
 */
export type Actor = {
  /**
   * An unique identifier of the actor.
   */
  id: string;
  /**
   * The type of the actor.
   */
  type: Actor.type;
  /**
   * The email address of the user.
   */
  email?: string;
  /**
   * The Indentity data of the device.
   */
  identity_data?: string;
};
export namespace Actor {
  /**
   * The type of the actor.
   */
  export enum type {
    USER = "user",
    DEVICE = "device",
  }
}
