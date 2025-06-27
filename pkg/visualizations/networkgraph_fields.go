package visualizations

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var NetworkNodeframeFields = []FrameFieldType{
	{
		Name: "id",
		Type: data.FieldTypeString,
	},
	{
		Name: "title",
		Type: data.FieldTypeString,
	},
	{
		Name: "mainstat",
		Type: data.FieldTypeString,
	},
	{
		Name: "color",
		Type: data.FieldTypeString,
	},

	{
		Name:        "detail__ResourceName",
		Type:        data.FieldTypeString,
		DisplayName: "ResourceName",
	},

	{
		Name:        "detail__NamespaceName",
		Type:        data.FieldTypeString,
		DisplayName: "Namespace",
	},
}

func GetNetworkNodeFields() []*data.Field {

	fields := make([]*data.Field, len(NetworkNodeframeFields))
	for i, field := range NetworkNodeframeFields {
		f := data.NewFieldFromFieldType(field.Type, 0)
		f.Name = field.Name
		if field.DisplayName != "" {
			f.Config = &data.FieldConfig{
				DisplayName:       field.DisplayName,
				DisplayNameFromDS: field.DisplayName,
			}
		}
		fields[i] = f

	}

	return fields

}
