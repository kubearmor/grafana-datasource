package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	// "strings"

	"net/http"

	"github.com/accuknox/kubearmor/pkg/adapters"
	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/accuknox/kubearmor/pkg/visualizations"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	// "github.com/grafana/grafana-plugin-sdk-go/data"
	// "github.com/kubearmor/KubeArmor/KubeArmor/types"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

const (
	pts0   = "pts0"
	denied = "Permission denied"
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

	opts, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("http client options: %w", err)
	}
	PluginSettings, err := models.LoadPluginSettings(settings)

	if err != nil {
		return nil, fmt.Errorf("Error in plugin settings: %w", err)
	}

	Backend := PluginSettings.Backend

	cl, err := httpclient.New(opts)
	if err != nil {
		return nil, fmt.Errorf("httpclient new: %w", err)
	}

	var datastoreConfig models.DataStoreConfig

	if err := json.Unmarshal(settings.JSONData, &datastoreConfig); err != nil {
		return nil, fmt.Errorf("error unmarshaling settings: %w", err)
	}

	// Secure fields
	if password, ok := settings.DecryptedSecureJSONData["basicAuthPassword"]; ok {
		datastoreConfig.Password = password
	}
	datastoreConfig.Username = settings.BasicAuthUser
	if caCertPem, ok := settings.DecryptedSecureJSONData["tlsCACert"]; ok {
		datastoreConfig.CACert = []byte(caCertPem)
	}

	datastoreConfig.URL = settings.URL
	datastoreConfig.Index = PluginSettings.Index
	backendSVC, err := adapters.GetBackendService(ctx, Backend, datastoreConfig)
	if err != nil {
		return nil, err
	}

	return &Datasource{
		settings:   settings,
		httpClient: cl,
		BackendSvc: backendSVC,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	settings   backend.DataSourceInstanceSettings
	BackendSvc adapters.BackendService

	httpClient *http.Client
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
	d.httpClient.CloseIdleConnections()
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res

	}

	return response, nil
}

func (d *Datasource) query(ctx context.Context, _ backend.PluginContext, q backend.DataQuery) backend.DataResponse {
	// var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm models.QueryModel

	ctxLogger := log.DefaultLogger.FromContext(ctx)

	err := json.Unmarshal(q.JSON, &qm)
	if err != nil {
		ctxLogger.Error("Error while marshalling the query json")
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}
	visualization := visualizations.NewVisualization(ctx, d.BackendSvc, qm)

	return visualization.GetVisualization()
}

//	func getNGraph(logs []types.Log) models.NodeGraph {
//		return models.NodeGraph{}
//	}
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	backendService := d.BackendSvc
	return backendService.HealthCheck(ctx)

}
