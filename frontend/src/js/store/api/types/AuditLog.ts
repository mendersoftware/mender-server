/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Actor } from "./Actor";
import type { Object } from "./Object";
/**
 * The audit log.
 */
export type AuditLog = {
  actor: Actor;
  time: string;
  action: AuditLog.action;
  object: Object;
  change?: string;
};
export namespace AuditLog {
  export enum action {
    CREATE = "create",
    UPDATE = "update",
    DELETE = "delete",
    ABORT = "abort",
    REJECT = "reject",
    DECOMMISSION = "decommission",
  }
}
