/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Sort criteria definition
 */
export type SortCriteria = {
  /**
   * Attribute name.
   */
  attribute: string;
  /**
   * Attribute scope.
   */
  scope: string;
  /**
   * Order direction, ascending ("asc") or descending ("desc").
   */
  order: SortCriteria.order;
};
export namespace SortCriteria {
  /**
   * Order direction, ascending ("asc") or descending ("desc").
   */
  export enum order {
    ASC = "asc",
    DESC = "desc",
  }
}
