package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/accuknox/kubearmor/pkg/adapters"
	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
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

type BackendService interface {
	GetLogs(ctx context.Context, qm models.QueryModel) []types.Log
	HealthCheck(ctx context.Context) (*backend.CheckHealthResult, error)
}

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
	backendSVC := getBackendService(ctx, Backend, datastoreConfig)

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
	BackendSvc BackendService

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
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm models.QueryModel

	ctxLogger := log.DefaultLogger.FromContext(ctx)

	err := json.Unmarshal(q.JSON, &qm)
	if err != nil {
		ctxLogger.Error("Error while marshalling the query json")
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	} else {
		ctxLogger.Info("Query json is sucessfully marshalled operation: ")
	}

	Nodegraph := getNodeGraph(ctx, d, qm)

	Nodefields := models.GetNodeFields()
	EdgeFields := models.GetEdgeFields()
	NetworkFields := models.GetNetworkNodeFields()

	Nodeframe := data.NewFrame("Nodes")
	if qm.Operation == "Process" {

		Nodeframe.Fields = Nodefields
	} else {
		Nodeframe.Fields = NetworkFields
	}

	EdgeFrame := data.NewFrame("Edges")
	EdgeFrame.Fields = EdgeFields

	var frameMeta = data.FrameMeta{
		PreferredVisualization: data.VisTypeNodeGraph,
	}
	Nodeframe.SetMeta(&frameMeta)
	EdgeFrame.SetMeta(&frameMeta)

	for _, node := range Nodegraph.Nodes {
		if qm.Operation == "Process" {

			Nodeframe.AppendRow(
				node.ID,
				node.Title,
				node.MainStat,
				node.Color,
				node.DetailClusterName,
				node.DetailHostName,
				node.DetailNamespaceName,
				node.DetailPodName,
				node.DetailLabels,
				node.DetailContainerID,
				node.DetailContainerName,
				node.DetailContainerImage,
				node.DetailParentProcessName,
				node.DetailProcessName,
				int64(node.DetailHostPPID),
				int64(node.DetailHostPID),
				int64(node.DetailPPID),
				int64(node.DetailPID),
				int64(node.DetailUID),
				node.DetailType,
				node.DetailSource,
				node.DetailOperation,
				node.DetailResource,
				node.DetailData,
				node.DetailResult,
				node.DetailCwd,
				node.DetailTTY,
			)
		} else if qm.Operation == "Network" {

			Nodeframe.AppendRow(
				node.ID,
				node.Title,
				node.MainStat,
				node.Color,
				node.DetailPodName,
				node.DetailNamespaceName,
			)
		}
	}

	for _, edge := range Nodegraph.Edges {
		EdgeFrame.AppendRow(edge.ID, edge.Source, edge.Target, edge.Mainstat, edge.Count)
	}

	response.Frames = append(response.Frames, Nodeframe)
	response.Frames = append(response.Frames, EdgeFrame)

	return response
}

func getBackendService(ctx context.Context, backendName string, dsc models.DataStoreConfig) BackendService {

	ctxLogger := log.DefaultLogger.FromContext(ctx)
	switch backendName {

	case "ELASTICSEARCH":
		return nil
	case "OPENSEARCH":
		client, err := adapters.NewOpenSearchClient(dsc, true)
		if err != nil {
			ctxLogger.Error("Cannot create opensearch client %v :", err)
		}
		return client

	}
	return nil
}

func getNodeGraph(ctx context.Context, datasource *Datasource, qm models.QueryModel) models.NodeGraph {
	service := datasource.BackendSvc // GetLogs gets the KubeArmor logs from the respective datastore
	logs := service.GetLogs(ctx, qm)

	ctxLogger := log.DefaultLogger.FromContext(ctx)
	ctxLogger.Info(fmt.Sprintf("received logs with len: %d", len(logs)))

	if qm.Operation == "Process" {
		return getProcessGraph(logs, qm)
	}

	return models.NodeGraph{}
}

func getProcessGraph(logs []types.Log, MyQuery models.QueryModel) models.NodeGraph {

	colors := []string{"orange", "green", "cyan", "rose"}

	var processLogs []types.Log

	for _, log := range logs {

		if log.TTY == pts0 &&
			log.Operation == MyQuery.Operation && (MyQuery.NamespaceQuery == "All" || log.NamespaceName == MyQuery.NamespaceQuery) &&
			(MyQuery.LabelQuery == "All" || log.Labels == MyQuery.LabelQuery) {
			processLogs = append(processLogs, log)
		}

	}

	/* Nodes */

	var ProcessNodes []models.NodeFields

	var processEdges []models.EdgeFields

	nodeMap := make(map[string]interface{})

	for _, log := range processLogs {
		isBlocked := log.Result == denied

		if log.PPID == 0 {
			colorIndex := random(0, len(colors)-1)
			cnode := models.NodeFields{
				ID:                  log.ContainerName + log.NamespaceName,
				Title:               log.ContainerName,
				Color:               colors[colorIndex],
				ChildNode:           fmt.Sprintf("%d%s", log.HostPID, log.PodName),
				DetailContainerName: log.ContainerName,
				DetailNamespaceName: log.NamespaceName,
			}

			ProcessNodes = append(ProcessNodes, cnode)

			edge := models.EdgeFields{
				ID:     fmt.Sprintf("%s%s%s%s", cnode.ID, cnode.ChildNode, cnode.DetailNamespaceName, cnode.DetailContainerName),
				Source: fmt.Sprintf("%s", cnode.ID),
				Target: fmt.Sprintf("%s", cnode.ChildNode),

				Mainstat: fmt.Sprintf("%s", "ContainerEdge"),
				Count:    "None",
			}
			nodeMap[cnode.ID] = ""

			processEdges = append(processEdges, edge)

		} else {

			edge := models.EdgeFields{
				ID:       fmt.Sprintf("%s%d%d", fmt.Sprintf("%d%s%s", log.HostPID, log.ContainerName, log.NamespaceName), log.PPID, log.HostPID),
				Source:   fmt.Sprintf("%d%s", log.HostPPID, log.PodName),
				Target:   fmt.Sprintf("%d%s", log.HostPID, log.PodName),
				Mainstat: fmt.Sprintf("%s", log.Data),

				Count: "None",
			}
			processEdges = append(processEdges, edge)
		}
		nodeId := fmt.Sprintf("%d%s", log.HostPID, log.PodName)
		nodeMap[nodeId] = ""
		node := models.NodeFields{
			ID:       nodeId,
			Title:    log.ProcessName,
			MainStat: log.Source,
			Color:    "white",
			// DetailTimestamp:         log.Timestamp,
			// NodeRadius:              "5",
			DetailClusterName:       log.ClusterName,
			DetailHostName:          log.HostName,
			DetailNamespaceName:     log.NamespaceName,
			DetailPodName:           log.ContainerName, // Using ContainerName as PodName for demonstration
			DetailLabels:            log.Labels,
			DetailContainerID:       log.ContainerID,
			DetailContainerName:     log.ContainerName,
			DetailContainerImage:    log.ContainerImage,
			DetailParentProcessName: log.ParentProcessName,
			DetailProcessName:       log.ProcessName,
			DetailHostPPID:          int64(log.HostPPID),
			DetailHostPID:           int64(log.HostPID),
			DetailPPID:              int64(log.PPID),
			DetailPID:               int64(log.PID),
			DetailUID:               int64(log.UID),
			DetailType:              log.Type,
			DetailSource:            log.Source,
			DetailOperation:         log.Operation,
			DetailResource:          log.Resource,
			DetailData:              log.Data,
			DetailResult:            log.Result,
			DetailCwd:               log.Cwd,
			DetailTTY:               log.TTY,
		}

		if isBlocked {
			node.Color = "red"
		}

		ProcessNodes = append(ProcessNodes, node)

	}

	NewprocessEdges := []models.EdgeFields{}

	for _, edge := range processEdges {
		if _, ok := nodeMap[edge.Source]; ok {
			if _, ok := nodeMap[edge.Target]; ok {
				NewprocessEdges = append(NewprocessEdges, edge)
			}
		}
	}

	var nodeGraph = models.NodeGraph{
		Nodes: ProcessNodes,
		Edges: NewprocessEdges,
	}

	return nodeGraph
}

func getNGraph(logs []types.Log) models.NodeGraph {
	return models.NodeGraph{}
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	backendService := d.BackendSvc
	return backendService.HealthCheck(ctx)

}
