import { DataSourcePluginOptionsEditorProps, } from "@grafana/data";
import { DataSourceHttpSettings, InlineField, Select, Input } from "@grafana/ui";
import React from "react";
import { MyDataSourceOptions, } from "../types";

interface Props
  extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> { }

export const ConfigEditor: React.FC<Props> = ({ onOptionsChange, options }) => {
  const BackendOptions = [
    { label: "Elasticsearch", value: "ELASTICSEARCH" },
    { label: "Loki ", value: "LOKI" },
    { label: "Opensearch", value: "OPENSEARCH" },
  ];

  const onBackendOptionsChange = (backendName: string) => {
    const jsonData = {
      ...options.jsonData,
      backendName: backendName,
    };
    onOptionsChange({ ...options, jsonData });
  };

  return (
    <div>
      <DataSourceHttpSettings
        defaultUrl="https://elasticsearch:9200"
        dataSourceConfig={options}
        onChange={onOptionsChange}
      />

      <InlineField label="Backend" labelWidth={16}>
        <Select
          options={BackendOptions}
          value={options.jsonData.backendName}
          onChange={(v) => {
            onBackendOptionsChange(v.value!);
          }}
        />
      </InlineField>

      <InlineField label="Index Name" labelWidth={16}>
        <Input
          value={options.jsonData.index || ""}
          placeholder="e.g. logs-*"
          onChange={(e) => {
            const jsonData = {
              ...options.jsonData,
              index: e.currentTarget.value,
            };
            onOptionsChange({ ...options, jsonData });
          }}
        />
      </InlineField>
    </div>
  );
};
