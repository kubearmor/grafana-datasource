import { DataSourceJsonData } from "@grafana/data";
import { DataQuery } from "@grafana/schema";

export interface MyQuery extends DataQuery, QueryType {
  // backendURL?: string;
}

export interface QueryType {
  NamespaceQuery?: string;
  LabelQuery?: string;
  Operation: string;
  BatchSize?: number;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  NamespaceQuery: "All",
  LabelQuery: "All",
  Operation: "Process",
};

// export interface DataPoint {
//   Time: number;
//   Value: number;
// }
//
// export interface DataSourceResponse {
//   datapoints: DataPoint[];
// }

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
  backendName: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
  username?: string;
  password?: string;
}
