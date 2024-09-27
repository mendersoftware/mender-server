/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Card information stored in Stripe. The full type description is available at "https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.yaml#/components/schemas/payment_method_card"
 */
export type CardData = {
  /**
   * Card brand. Can be `amex`, `diners`, `discover`, `jcb`, `mastercard`, `unionpay`, `visa`, or `unknown`.
   */
  brand?: string;
  /**
   * Checks on Card address and CVC if provided.
   */
  checks?: any;
  /**
   * Two-letter ISO code representing the country of the card. You could use this attribute to get a sense of the international breakdown of cards you've collected.
   */
  country?: string;
  /**
   * Two-digit number representing the card's expiration month.
   */
  exp_month?: number;
  /**
   * Four-digit number representing the card's expiration year.
   */
  exp_year?: number;
  /**
   * Uniquely identifies this particular card number. You can use this attribute to check whether two customers who’ve signed up with you are using the same card number,for example. For payment methods that tokenize card information (Apple Pay, Google Pay), the tokenized number might be provided instead of the underlying card number.
   */
  fingerprint?: string;
  /**
   * Card funding type. Can be `credit`, `debit`, `prepaid`, or `unknown`.
   */
  funding?: string;
  /**
   * The last four digits of the card.
   */
  last4?: string;
  /**
   * Contains details on how this Card maybe be used for 3D Secure authentication.
   */
  three_d_secure_usage?: any;
  /**
   * If this Card is part of a card wallet, this contains the details of the card wallet.
   */
  wallet?: any;
};
