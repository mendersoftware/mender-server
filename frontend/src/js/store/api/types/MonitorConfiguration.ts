/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { DBusSubsystem } from "./DBusSubsystem";
import type { LogSubsystem } from "./LogSubsystem";
import type { ServiceSubsystem } from "./ServiceSubsystem";
/**
 * Monitor configuration data.
 */
export type MonitorConfiguration = {
  name: string;
  status: string;
  type: MonitorConfiguration.type;
} & (LogSubsystem | ServiceSubsystem | DBusSubsystem);
export namespace MonitorConfiguration {
  export enum type {
    LOG = "log",
    SERVICE = "service",
    DBUS = "dbus",
  }
}
