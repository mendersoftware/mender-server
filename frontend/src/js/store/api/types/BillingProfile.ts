/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Billing profile contains information about the customer.
 */
export type BillingProfile = {
  email: string;
  name: string;
  phone?: string;
  address: {
    country: string;
    state: string;
    city: string;
    postal_code: string;
    line1: string;
    line2?: string;
  };
};
