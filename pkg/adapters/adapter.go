package adapters

import (
	"context"
	"github.com/accuknox/kubearmor/pkg/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
)

type BackendService interface {
	GetLogs(ctx context.Context, qm models.QueryModel) ([]types.Log, error)
	HealthCheck(ctx context.Context) (*backend.CheckHealthResult, error)
}

func GetBackendService(ctx context.Context, backendName string, dsc models.DataStoreConfig) (BackendService, error) {

	ctxLogger := log.DefaultLogger.FromContext(ctx)
	switch backendName {

	case "ELASTICSEARCH":
		return nil, nil

	case "OPENSEARCH":
		client, err := newOpenSearchClient(dsc)
		if err != nil {
			ctxLogger.Error("Cannot create opensearch client %v :", err)
			return nil, err
		}

		return client, nil
	}
	return nil, nil
}
