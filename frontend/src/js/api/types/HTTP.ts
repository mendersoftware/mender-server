/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * HTTP Webhook configuration.
 */
export type HTTP = {
  http: {
    /**
     * The destination URL for the webhook. The webhook will send POST requests with event details to this target URL.
     */
    url: string;
    /**
     * An optional secret used to verify the integrity of the payload. The string must be in hexadecimal format.
     */
    secret?: string;
  };
};
