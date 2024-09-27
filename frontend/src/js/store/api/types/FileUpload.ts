/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type FileUpload = {
  /**
   * The destination path on the device
   */
  path: string;
  /**
   * The numerical UID of the file on the device
   */
  uid?: number;
  /**
   * The numerical GID of the file on the device
   */
  gid?: number;
  /**
   * The octal representation of the mode of the file on the device
   */
  mode?: string;
  file?: Blob;
};
