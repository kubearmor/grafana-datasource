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

	osCaCertPath := os.Getenv("OS_CA_CERT_PATH")

	osAddress := os.Getenv("OS_URL")
	if osAddress == "" {
		osAddress = dsc.URL
	}
	username := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	index := os.Getenv("OS_INDEX")
	if index == "" {
		index = "*"
	}

	cfg := opensearch.Config{
		Addresses: []string{osAddress},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecureTLS,
			},
		},
	}

	if osCaCertPath != "" {
		caCertBytes, err := os.ReadFile(osCaCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		cfg.CACert = caCertBytes
	}

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
