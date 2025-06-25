import React from 'react';
import { InlineField, Stack, Select, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, data }: Props) {

  // const Allselectable: SelectableValue<string> = {
  //   label: "All",
  //   value: "All"
  // }
  // const [namespace, setNamespace] = useState<SelectableValue<string>[]>([Allselectable])
  // const [labels, setLabels] = useState<SelectableValue<string>[]>([Allselectable])
  const onBatchChange = (BatchSize: number) => {

    onChange({ ...query, BatchSize: BatchSize });
  }


  const onQueryChange = (qname: string, type: string) => {
    switch (type) {
      case "NAMESPACE":
        onChange({ ...query, NamespaceQuery: qname });
        break;

      case "LABEL":
        onChange({ ...query, LabelQuery: qname });
        break;

      case "OPERATION":
        onChange({ ...query, Operation: qname });
        break;   // ← add this

      case "VISUALIZATION":
        onChange({ ...query, Visualization: qname });
        break;   // ← and this, too

      default:
        return;
    }

    onRunQuery();
  };

  const { NamespaceQuery, LabelQuery, Operation, BatchSize, Visualization } = query;
  const frame = data?.series[0];
  const Namespaces = frame?.fields.find(i => i.name === 'detail__NamespaceName') || frame?.fields.find(i => i.name == "Namespace")

  const Labels = frame?.fields.find(i => i.name === 'detail__Labels') || frame?.fields.find(i => i.name == "labels")
  const uniqueNamespaces = new Set<string>(["All"]);
  const uniqueLabels = new Set<string>(["All"])

  if (Namespaces && Namespaces.values) {
    // Iterate over each value and add it to the set
    Namespaces.values.forEach(value => {
      uniqueNamespaces.add(value);
    });
  }

  if (Labels && Labels.values) {
    // Iterate over each value and add it to the set
    Labels.values.forEach(value => {
      uniqueLabels.add(value)
    });
  }

  const uniqueNamespaceArray = Array.from(uniqueNamespaces);

  const uniqueLabelsArray = Array.from(uniqueLabels);


  const namespaceOptions = uniqueNamespaceArray.map(i => {
    const selectable: SelectableValue<string> = {
      label: i,
      value: i
    }
    return selectable
  })

  const labelOptions = uniqueLabelsArray.map(i => {
    const selectable: SelectableValue<string> = {
      label: i,
      value: i
    }
    return selectable
  })

  const operationOptions: SelectableValue[] = [
    {
      label: "Process",
      value: "Process",

    },
    {
      label: "Network",
      value: "Network"
    },

    {
      label: "File",
      value: "File"
    },

    {
      label: "Syscall",
      value: "Syscall"
    },
  ]

  const VisualizationOptions: SelectableValue[] = [
    {
      label: "ProcessGraph",
      value: "PROCESSGRAPH",

    },
    {
      label: "NetworkGraph",
      value: "NETWORKGRAPH"
    },

    {
      label: "AlertCount",
      value: "ALERTCOUNTGRAPH"
    },

    {
      label: "Profile View",
      value: "PROFILE"
    },
  ]

  return (
    <Stack gap={0}>
      <InlineField label="namespace" labelWidth={16} tooltip="filter using Namespaces">

        <Select
          options={namespaceOptions}
          value={NamespaceQuery || ''}
          onChange={v => {
            onQueryChange(v.value!, "NAMESPACE")
          }} />


      </InlineField>

      <InlineField label="label" labelWidth={16} tooltip="filter using labels">

        <Select
          options={labelOptions}
          value={LabelQuery || ''}
          onChange={v => {
            onQueryChange(v.value!, "LABEL")
          }} />


      </InlineField>

      <InlineField label="Operation" labelWidth={16} >

        <Select
          options={operationOptions}
          value={Operation || 'Process'}
          onChange={v => {
            onQueryChange(v.value!, "OPERATION")
          }} />


      </InlineField>

      <InlineField label="Visualization" labelWidth={16} >

        <Select
          options={VisualizationOptions}
          value={Visualization || 'PROCESSGRAPH'}
          onChange={v => {
            onQueryChange(v.value!, "VISUALIZATION")
          }} />


      </InlineField>


      <InlineField label="Batch Size" labelWidth={16} tooltip="Number of items to process in each batch">
        <Input
          type="number"
          value={BatchSize || 1000}
          onChange={e => onBatchChange(Number(e.currentTarget.value))}
          min={1}
        />
      </InlineField>
    </Stack>
  );
}
