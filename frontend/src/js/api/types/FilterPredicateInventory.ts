/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

/**
 * Attribute filter predicate
 */
export type FilterPredicateInventory = {
  /**
   * Attribute name.
   */
  attribute: string;
  scope: string;
  /**
   * Type or operator of the filter predicate.
   *
   * | *Operator* | *Name*                       | *Argument type* |
   * |:----------:|:-----------------------------|:----------------|
   * | $eq        | Equal (`==`)                 | any             |
   * | $ne        | Not equal (`!=`)             | any             |
   * | $gt        | Greater than (`>`)           | any             |
   * | $gte       | Greater than or equal (`>=`) | any             |
   * | $lt        | Less than (`<`)              | any             |
   * | $lte       | Less than or equal (`<=`)    | any             |
   * | $exists    | Attribute exists             | bool            |
   * | $in        | Is an element of             | array           |
   * | $nin       | Is not an element of         | array           |
   * | $regex     | Regex filter                 | string          |
   */
  type: FilterPredicateInventory.type;
  /**
   * The value of the attribute to be used in filtering.
   * Attribute type is implicit, inferred from the JSON type.
   *
   * The $exists operator expects a boolean value: true means the specified
   * attribute exists, false means the specified attribute doesn't exist.
   *
   * The $regex operator expects a string as a Perl compatible regular expression
   * (PCRE), automatically anchored by ^. If the regular expression is not valid,
   * the filter will produce no results. If you need to specify options and flags,
   * you can provide the full regex in the format of /regex/flags, for example
   * `/[a-z]+/i`.
   */
  value: string;
};

export namespace FilterPredicateInventory {
  /**
   * Type or operator of the filter predicate.
   *
   * | *Operator* | *Name*                       | *Argument type* |
   * |:----------:|:-----------------------------|:----------------|
   * | $eq        | Equal (`==`)                 | any             |
   * | $ne        | Not equal (`!=`)             | any             |
   * | $gt        | Greater than (`>`)           | any             |
   * | $gte       | Greater than or equal (`>=`) | any             |
   * | $lt        | Less than (`<`)              | any             |
   * | $lte       | Less than or equal (`<=`)    | any             |
   * | $exists    | Attribute exists             | bool            |
   * | $in        | Is an element of             | array           |
   * | $nin       | Is not an element of         | array           |
   * | $regex     | Regex filter                 | string          |
   */
  export enum type {
    _EQ = "$eq",
    _GT = "$gt",
    _GTE = "$gte",
    _IN = "$in",
    _LT = "$lt",
    _LTE = "$lte",
    _NE = "$ne",
    _NIN = "$nin",
    _EXISTS = "$exists",
    _REGEX = "$regex",
  }
}
