/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Upgrade a trial tenant to a given plan.
 */
export type UpgradeCompleteRequest = {
  /**
   * customer plan
   */
  plan: UpgradeCompleteRequest.plan;
};
export namespace UpgradeCompleteRequest {
  /**
   * customer plan
   */
  export enum plan {
    OS = "os",
    PROFESSIONAL = "professional",
    ENTERPRISE = "enterprise",
  }
}
