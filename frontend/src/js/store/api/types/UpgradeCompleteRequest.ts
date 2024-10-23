/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { BillingProfile } from "./BillingProfile";
/**
 * Upgrade a trial tenant to a given plan.
 */
export type UpgradeCompleteRequest = {
  /**
   * customer plan
   */
  plan: UpgradeCompleteRequest.plan;
  billing_profile?: BillingProfile;
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
