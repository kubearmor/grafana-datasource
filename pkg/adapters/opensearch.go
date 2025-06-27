package adapters

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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

func newOpenSearchClient(dsc models.DataStoreConfig) (*OpenSearchClient, error) {

	osAddress := firstNonEmpty(dsc.URL, os.Getenv("OS_URL"))
	if osAddress == "" {
		return nil, errors.New("OpenSearch URL must be configured")
	}

	index := firstNonEmpty(dsc.Index, os.Getenv("OS_INDEX"))
	if index == "" {
		return nil, errors.New("OpenSearch index must be configured")
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12, // Enforce minimum TLS version
		InsecureSkipVerify: dsc.TLSSkipVerify,
	}

	// Handle CA certificate if provided
	if dsc.TLSAuthWithCACert && len(dsc.CACert) > 0 && !dsc.TLSSkipVerify {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(dsc.CACert) {
			return nil, errors.New("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		// Add connection pooling parameters
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	cfg := opensearch.Config{
		Addresses: []string{osAddress},
		Username:  dsc.Username,
		Password:  dsc.Password,
		Transport: transport,
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	return &OpenSearchClient{
		client: client,
		index:  index,
	}, nil
}

// Helper function to get first non-empty string
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func (osc *OpenSearchClient) GetLogs(ctx context.Context, qm models.QueryModel) ([]types.Log, error) {

	ctxLogger := log.DefaultLogger.FromContext(ctx)

	if qm.BatchSize < 0 {
		ctxLogger.Warn("Invalid batch size, using default", "requested", qm.BatchSize)
	}
	qm.BatchSize = 1000

	// Prepare search query
	query := getQuery(qm)
	ctxLogger.Info("myq: %s", query)
	indices := []string{osc.index}
	searchReq := opensearchapi.SearchRequest{
		Index: indices,
		Body:  strings.NewReader(query),
	}

	res, err := searchReq.Do(ctx, osc.client)
	if err != nil {
		ctxLogger.Error("Search error: %v", err)
		return []types.Log{}, err
	}
	defer res.Body.Close()

	// Check for OpenSearch errors
	if res.IsError() {
		var osErr struct {
			Error struct {
				Reason string `json:"reason"`
			} `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&osErr); err != nil {
			ctxLogger.Error("Failed to parse OpenSearch error", "status", res.StatusCode)
			return nil, fmt.Errorf("search failed with status %d", res.StatusCode)
		}
		ctxLogger.Error("OpenSearch error", "status", res.StatusCode, "reason", osErr.Error.Reason)
		return nil, fmt.Errorf("search failed: %s (status %d)", osErr.Error.Reason, res.StatusCode)
	}

	// Parse successful response
	var osResponse struct {
		Hits struct {
			Hits []struct {
				Source types.Log `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&osResponse); err != nil {
		ctxLogger.Error("Error decoding JSON: %v", err)
		return []types.Log{}, err
	}

	logs := make([]types.Log, len(osResponse.Hits.Hits))
	for i, hit := range osResponse.Hits.Hits {
		logs[i] = hit.Source
	}
	ctxLogger.Debug("Successfully fetched logs", "count", len(logs))
	return logs, nil
}

func getQuery(qm models.QueryModel) string {
	processGraphquery := fmt.Sprintf(`{
  "query": {
    "wildcard": {
      "TTY.keyword": "pts*"
    }
  },
  "size": %d
}`, qm.BatchSize)

	networkGraphQuery := fmt.Sprintf(`{
  "query": {
    "wildcard": {
      "Data.keyword": "*tcp_*"
    }
  },
  "size": %d
}`, qm.BatchSize)

	alertQuery := fmt.Sprintf(`{
  "query": {
    "wildcard": {
      "Type.keyword": "*Matche*"
    }
  },
  "size": %d
}`, qm.BatchSize)

	op := qm.Operation

	var profileQuery string
	if op != "" {
		// both clauses: wildcard + term filter
		profileQuery = fmt.Sprintf(`{
  "query": {
    "bool": {
      "must": [
        {
          "wildcard": {
            "Type.keyword": "*Log*"
          }
        },
        {
          "term": {
            "Operation.keyword": "%s"
          }
        }
      ]
    }
  },
  "size": %d
}`, op, qm.BatchSize)
	} else {
		// only wildcard clause
		profileQuery = fmt.Sprintf(`{
  "query": {
    "wildcard": {
      "Type.keyword": "*Log*"
    }
  },
  "size": %d
}`, qm.BatchSize)
	}

	switch qm.Visualization {
	case models.PROCESSGRAPH:
		return processGraphquery
	case models.NETWORKGRAPH:
		return networkGraphQuery
	case models.ALERTCOUNTGRAPH:
		return alertQuery
	case models.PROFILE:
		return profileQuery

	}
	return ""
}

func (osc *OpenSearchClient) HealthCheck(ctx context.Context) (*backend.CheckHealthResult, error) {
	ctxLogger := log.DefaultLogger.FromContext(ctx)
	res := &backend.CheckHealthResult{}

	// 1. Check cluster health
	healthReq := opensearchapi.ClusterHealthRequest{}
	healthRes, err := healthReq.Do(ctx, osc.client)
	if err != nil {
		ctxLogger.Error("Cluster health request failed", err)
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("Cluster health request failed: %s", err.Error())
		return res, nil
	}
	defer healthRes.Body.Close()

	if healthRes.StatusCode != http.StatusOK {
		ctxLogger.Error("Cluster unhealthy status", healthRes.StatusCode)
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("Cluster unhealthy (status: %d)", healthRes.StatusCode)
		return res, nil
	}

	// 2. Check index existence
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{osc.index},
	}
	existsRes, err := existsReq.Do(ctx, osc.client)
	if err != nil {
		ctxLogger.Error("Index existence check failed", err)
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("Index check failed: %s", err.Error())
		return res, nil
	}
	defer existsRes.Body.Close()

	// 3. Evaluate index existence status
	switch existsRes.StatusCode {
	case http.StatusOK:
		// Both cluster and index are healthy
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusOk,
			Message: "OpenSearch cluster and index are healthy",
		}, nil
	case http.StatusNotFound:
		ctxLogger.Warn("Index missing", osc.index)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Index '%s' does not exist", osc.index),
		}, nil
	default:
		ctxLogger.Error("Unexpected index check status", existsRes.StatusCode)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Unexpected index check status: %d", existsRes.StatusCode),
		}, nil
	}
}
