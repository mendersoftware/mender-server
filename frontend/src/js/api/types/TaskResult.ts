/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { TaskResultCLI } from "./TaskResultCLI";
import type { TaskResultHTTPRequest } from "./TaskResultHTTPRequest";
import type { TaskResultHTTPResponse } from "./TaskResultHTTPResponse";
export type TaskResult = {
  success?: boolean;
  cli?: TaskResultCLI;
  request?: TaskResultHTTPRequest;
  response?: TaskResultHTTPResponse;
};
