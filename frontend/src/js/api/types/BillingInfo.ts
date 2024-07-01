/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CardData } from "./CardData";
import type { SubscriptionData } from "./SubscriptionData";
/**
 * Billing related data, such as credit card information and next billing date
 */
export type BillingInfo = {
  card?: CardData;
  subscription?: SubscriptionData;
};
