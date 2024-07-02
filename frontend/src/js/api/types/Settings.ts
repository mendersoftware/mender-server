/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * User settings.
 */
export type Settings = {
    /**
     * Enable/disable two-factor authentication for this user.
     */
    '2fa'?: Settings.'2fa';
};
export namespace Settings {
    /**
     * Enable/disable two-factor authentication for this user.
     */
    export enum '2fa' {
        ENABLED = 'enabled',
        DISABLED = 'disabled',
    }
}

