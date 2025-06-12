package adapters

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
	opensearch "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type OpenSearchClient struct {
	client *opensearch.Client
	index  string
}

func NewOpenSearchClient(dsc models.DataStoreConfig, allowInsecureTLS bool) (*OpenSearchClient, error) {

	osAddress := dsc.URL

	if osAddress == "" {
		osAddress = os.Getenv("OS_URL")
	}
	// username := os.Getenv("OS_USERNAME")
	// password := os.Getenv("OS_PASSWORD")
	index := os.Getenv("OS_INDEX")
	if index == "" {
		index = "test_alert"
	}

	cfg := opensearch.Config{
		Addresses: []string{dsc.URL},
		Username:  dsc.Username,
		Password:  dsc.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// if dsc.CACert != nil {
	// 	cfg.CACert = dsc.CACert
	// }

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	osc := &OpenSearchClient{
		client: client,
		index:  index,
	}
	return osc, nil
}

func (osc *OpenSearchClient) GetLogs(ctx context.Context, qm models.QueryModel, index string) []types.Log {

	ctxLogger := log.DefaultLogger.FromContext(ctx)

	batchSize := 1000
	if qm.BatchSize > 0 {
		batchSize = qm.BatchSize
	}

	// Prepare search query
	query := fmt.Sprintf(`{
  "query": {
    "wildcard": {
      "TTY.keyword": "pts*"
    }
  },
  "size": %d
}`, batchSize)

	indices := []string{osc.index}
	searchReq := opensearchapi.SearchRequest{
		Index: indices,
		Body:  strings.NewReader(query),
	}

	res, err := searchReq.Do(context.Background(), osc.client)
	if err != nil {
		ctxLogger.Error("Search error: %v", err)
	}
	defer res.Body.Close()

	var osresponse struct {
		Hits struct {
			Hits []struct {
				Source types.Log `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&osresponse); err != nil {
		ctxLogger.Error("Error decoding JSON: %v", err)
		return []types.Log{}
	}

	logs := make([]types.Log, len(osresponse.Hits.Hits))
	for _, source := range osresponse.Hits.Hits {
		logs = append(logs, source.Source)
	}

	return logs
}

func (osc *OpenSearchClient) HealthCheck(ctx context.Context) (*backend.CheckHealthResult, error) {

	res := &backend.CheckHealthResult{}
	healthReq := opensearchapi.ClusterHealthRequest{}
	healthRes, err := healthReq.Do(context.Background(), osc.client)

	status := healthRes.StatusCode

	ctxLogger := log.DefaultLogger.FromContext(ctx)
	ctxLogger.Info("Healthcheck status", status)

	if err != nil {
		ctxLogger.Error("Healthcheck error in opensearch")
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("HealthCheck failed in opensearch %v", err)
		return res, err
	}
	defer healthRes.Body.Close()

	if status != http.StatusOK {

		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("error on checking health check status from backend  %d", status)

		return res, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Opensearch Health check sucess",
	}, nil
}
