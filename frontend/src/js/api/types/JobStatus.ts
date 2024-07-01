/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { InputParameter } from "./InputParameter";
import type { TaskResult } from "./TaskResult";
export type JobStatus = {
  id?: string;
  workflowName?: string;
  inputParameters?: Array<InputParameter>;
  status?: string;
  results?: Array<TaskResult>;
};
