/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AWSCredentials } from "./AWSCredentials";
import type { AzureSharedAccessSecret } from "./AzureSharedAccessSecret";
import type { HTTP } from "./HTTP";
export type Credentials = {
  /**
   * The credential type
   */
  type: Credentials.type;
} & (AWSCredentials | AzureSharedAccessSecret | HTTP);
export namespace Credentials {
  /**
   * The credential type
   */
  export enum type {
    AWS = "aws",
    SAS = "sas",
    HTTP = "http",
  }
}
