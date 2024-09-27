/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Plan/add-on change request of a tenant account.
 */
export type PlanChangeRequest = {
  current_plan?: string;
  requested_plan?: string;
  current_addons?: string;
  requested_addons?: string;
  user_message?: string;
};
