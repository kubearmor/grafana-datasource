import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { MyQuery, MyDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    return {
      ...query,

      NamespaceQuery: getTemplateSrv().replace(query.NamespaceQuery, scopedVars),
      Operation: getTemplateSrv().replace(query.Operation, scopedVars),
    };
  }

  filterQuery(query: MyQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.Operation;
  }
}
