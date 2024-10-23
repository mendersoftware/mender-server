/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeviceSortTerm = {
  /**
   * Attribute key to sort by.
   */
  attribute: string;
  /**
   * Scope the attribute key belongs to.
   */
  scope?: string;
  /**
   * Sort order: ascending/descending.
   */
  order: DeviceSortTerm.order;
};
export namespace DeviceSortTerm {
  /**
   * Sort order: ascending/descending.
   */
  export enum order {
    ASC = "asc",
    DESC = "desc",
  }
}
