/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AddressUpdate } from "./AddressUpdate";
/**
 * Billing profile contains partial billing information.
 */
export type BillingProfileUpdate = {
  /**
   * The customer's email address.
   */
  email?: string;
  /**
   * The customer's full name or business name.
   */
  name?: string;
  /**
   * The customer's phone number.
   */
  phone?: string;
  address?: AddressUpdate;
  /**
   * Mailing and shipping address for the customer. Appears on invoices.
   */
  shipping?: {
    /**
     * Name of the recipient.
     */
    name?: string;
    /**
     * Phone number of the recipient (including extension).
     */
    phone?: string;
    address?: AddressUpdate;
  };
};
