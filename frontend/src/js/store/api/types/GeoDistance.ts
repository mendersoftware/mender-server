/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { GeoPoint } from "./GeoPoint";
export type GeoDistance = {
  /**
   * Distance from given geo point. Supported units are:
   * mi or miles, yd or yards, ft or feet, in or inch, km or kilometers,
   * m or meters, cm or centimeters, mm or millimeters, NM, nmi, or nauticalmiles.
   * The default unit is meters if none is specified.
   */
  distance: string;
  location: GeoPoint;
};
