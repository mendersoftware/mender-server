/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * dbus subsystem specific configuration
 */
export type DBusSubsystem = {
  dbus: {
    /**
     * Check name.
     */
    name: string;
    /**
     * DBus pattern to look for
     */
    pattern: string;
    /**
     * DBus watch expression
     */
    watch: string;
    /**
     * Number of seconds after which the alert is considered expired and an OK is sent
     */
    alert_expiration?: number;
  };
};
