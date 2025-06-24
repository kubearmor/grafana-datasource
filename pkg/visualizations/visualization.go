package visualizations

import (
	"context"

	"github.com/accuknox/kubearmor/pkg/adapters"

	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func NewVisualization(ctx context.Context, bsvc adapters.BackendService, qm models.QueryModel) Visualization {
	return Visualization{
		ctx,
		bsvc,
		qm,
	}
}

type Visualization struct {
	ctx            context.Context
	backendService adapters.BackendService
	qm             models.QueryModel
}

func (v *Visualization) GetVisualization() backend.DataResponse {
	switch v.qm.Visualization {
	case models.PROCESSGRAPH:
		return v.getProcessGraph()
	case models.NETWORKGRAPH:
		return v.getNetworkGraph()

	}
	return backend.DataResponse{}
}

type FrameFieldType struct {
	Name        string         `json:"name"`
	Type        data.FieldType `json:"type"`
	DisplayName string
}
