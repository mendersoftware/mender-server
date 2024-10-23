/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type DeploymentSortTerm = {
  /**
   * Attribute key to sort by.
   */
  attribute: string;
  /**
   * Sort order: ascending/descending.
   */
  order: DeploymentSortTerm.order;
};
export namespace DeploymentSortTerm {
  /**
   * Sort order: ascending/descending.
   */
  export enum order {
    ASC = "asc",
    DESC = "desc",
  }
}
