/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * log subsystem specific configuration
 */
export type LogSubsystem = {
  log: {
    /**
     * Log pattern
     */
    pattern: string;
    /**
     * Path to the log file or command to execute to get the logs (prefixed with '@').
     */
    path: string;
    /**
     * Number of seconds after which the pattern is considered expired and an OK is sent
     */
    expire_seconds?: number;
  };
};
