/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Features } from "./Features";
/**
 * Plan descriptor.
 */
export type Plan = {
  /**
   * Unique name of the plan.
   */
  name: string;
  /**
   * Short information about the plan.
   */
  display_name: string;
  features: Features;
};
