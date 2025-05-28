package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	// "strings"

	"net/http"
	// "strconv"

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
var Backend string = ""

const (
	pts0   = "pts0"
	denied = "Permission denied"
)

type BackendService interface {
	GetLogs(ctx context.Context, qm models.QueryModel, index string) []types.Log
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
	// clientset := getK8sClient(ctxLogger)

	Backend = PluginSettings.Backend
	// mux := &sync.RWMutex{}
	// clustercache := &ClusterCache{
	// 	ipPodCache: make(map[string]PodServiceInfo),
	// 	mu:         mux,
	// }

	cl, err := httpclient.New(opts)
	if err != nil {
		return nil, fmt.Errorf("httpclient new: %w", err)
	}

	// client := &Client{
	// 	k8sClient:      clientset,
	// 	ClusterIPCache: clustercache,
	// }

	// go startInformers(client, ctxLogger)
	return &Datasource{
		settings:   settings,
		httpClient: cl,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	settings backend.DataSourceInstanceSettings

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

type queryModel struct {
	NamespaceQuery string `json:"NamespaceQuery,omitempty"`
	LabelQuery     string `json:"LabelQuery,omitempty"`
	Operation      string `json:"Operation"`
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

	// Nodegraph, _, _ := getGraphData(ctx, d, qm)
	Nodegraph := getNodeGraph(ctx, d, qm)

	Nodefields := getNodeFields()
	EdgeFields := getEdgeFields()
	NetworkFields := getNetworkNodeFields()

	Nodeframe := data.NewFrame("Nodes")
	if qm.Operation == "Process" {

		Nodeframe.Fields = Nodefields
	} else {
		Nodeframe.Fields = NetworkFields
	}

	EdgeFrame := data.NewFrame("Edges")
	EdgeFrame.Fields = EdgeFields

	// edgetest := models.EdgeFields{
	// 	ID:     "id",
	// 	Source: qm.NamespaceQuery,
	// 	Target: qm.Operation,
	// }
	// EdgeFrame.AppendRow(tty, string(qm.NamespaceQuery), fmt.Sprintf("%d", tot))
	// EdgeFrame.AppendRow(edgetest.ID, edgetest.Source, edgetest.Target)

	var frameMeta = data.FrameMeta{
		PreferredVisualization: data.VisTypeNodeGraph,
	}
	Nodeframe.SetMeta(&frameMeta)
	EdgeFrame.SetMeta(&frameMeta)
	// EdgeFrame := data.NewFrame("Edges")
	// add the frames to the response.

	for _, node := range Nodegraph.Nodes {
		if qm.Operation == "Process" {

			Nodeframe.AppendRow(
				node.ID,
				node.Title,
				node.MainStat,
				node.Color,
				// node.ChildNode,
				// node.NodeRadius,
				// node.Highlighted,
				// int64(node.DetailTimestamp),
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
				// node.ChildNode,
				// node.NodeRadius,
				// node.Highlighted,
				// node.DetailTimestamp,
				// node.DetailClusterName,
				// node.DetailHostName,
				// node.DetailNamespaceName,
				// node.DetailPodName,
				// node.DetailLabels,
				// node.DetailContainerID,
				// node.DetailContainerName,
				// node.DetailContainerImage,
				// node.DetailParentProcessName,
				// node.DetailProcessName,
				// int64(node.DetailHostPPID),
				// int64(node.DetailHostPID),
				// int64(node.DetailPPID),
				// int64(node.DetailPID),
				// int64(node.DetailUID),
				// node.DetailType,
				// node.DetailSource,
				// node.DetailOperation,
				// node.DetailResource,
				// node.DetailData,
				// node.DetailResult,
				// node.DetailCwd,
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

func getNodeFields() []*data.Field {

	fields := make([]*data.Field, len(models.NodeframeFields))
	for i, field := range models.NodeframeFields {
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

func getNetworkNodeFields() []*data.Field {

	fields := make([]*data.Field, len(models.NetworkNodeframeFields))
	for i, field := range models.NetworkNodeframeFields {
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

func getEdgeFields() []*data.Field {

	fields := make([]*data.Field, len(models.EdgeframeFields))
	for i, field := range models.EdgeframeFields {
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
	datastoreConfig := models.DataStoreConfig{
		URL:      datasource.settings.URL,
		Username: datasource.settings.User,
		Password: "",
	}
	service := getBackendService(ctx, Backend, datastoreConfig)
	// GetLogs gets the KubeArmor logs from the respective datastore
	logs := service.GetLogs(ctx, qm, "test_alert")

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

// func getNetworkGraph(ctxlogger log.Logger, logs []models.Log, MyQuery queryModel, datasource *Datasource) models.NodeGraph {
//
// 	var networkGraphs = []models.NetworkGraph{}
// 	var networkData = models.NetworkData{}
// 	var networkLogs []models.Log
// 	var NodeData = []models.NodeFields{}
// 	var EdgeData = []models.EdgeFields{}
// 	var EdgeAcceptMap = make(map[string]int)
// 	var EdgeConnectMap = make(map[string]int)
// 	var EdgeMap = make(map[string]int)
//
// 	for _, log := range logs {
// 		if log.Operation == MyQuery.Operation {
// 			networkLogs = append(networkLogs, log)
// 		}
// 	}
// 	for _, log := range networkLogs {
//
// 		datamap := extractdata(log.Data)
// 		containsHostname := strings.Contains(log.Resource, "hostname")
// 		if containsKprobe := strings.Contains(log.Data, "kprobe"); containsKprobe && containsHostname {
// 			/* Extracting from data field */
// 			kprobeData := datamap["kprobe"]
// 			domainData := datamap["domain"]
// 			peertype := datamap["ownertype"]
//
// 			// if peertype == "pod" || peertype == "service" {
//
// 			/* Extracting from Resource field */
//
// 			resourceMap := extractdata(log.Resource)
// 			remoteIP := resourceMap["remoteip"]
// 			peerhostName := resourceMap["hostname"]
// 			peerNamespace := resourceMap["namespace"]
// 			podServiceName := resourceMap["podname"]
// 			if peertype == "service" {
// 				peerhostName += " SVC"
// 				podServiceName = resourceMap["servicename"]
// 				podServiceName += " SVC"
//
// 			}
// 			// if podServiceName == "" {
// 			// 	podServiceName = log.PodName
// 			// }
// 			port := resourceMap["port"]
// 			protocol := resourceMap["protocol"]
//
// 			networkData = models.NetworkData{
// 				NetworkType: "kprobe:" + kprobeData,
// 				SockType:    "",
// 				Kprobe:      kprobeData,
// 				Domain:      domainData,
// 				RemoteIP:    remoteIP,
// 				HostName:    peerhostName,
// 				Port:        port,
// 				Protocol:    protocol,
// 			}
//
// 			ownerName := log.PodName
// 			ownerNamespace := log.NamespaceName
//
// 			if log.Owner.Name != "" {
// 				ownerName = log.Owner.Name
// 			}
//
// 			if log.Owner.Namespace != "" {
// 				ownerNamespace = log.Owner.Namespace
// 			}
//
// 			// if ownerName == "" {
// 			//
// 			// 	ownerName = log.Source
// 			// }
// 			// if ownerNamespace == "" {
// 			// 	ownerNamespace = log.NamespaceName
// 			// }
//
// 			node := models.NodeFields{
//
// 				ID:                  fmt.Sprintf("%s%s", log.PodName, log.NamespaceName),
// 				Title:               ownerName,
// 				MainStat:            log.PodName,
// 				Color:               "white",
// 				DetailPodName:       log.PodName,
// 				DetailNamespaceName: ownerNamespace,
// 			}
//
// 			if log.Result == denied {
// 				node.Color = "red"
// 			}
//
// 			switch kprobeData {
// 			case "tcp_accept":
// 				ID := fmt.Sprintf("%s%s%s%s%s", log.PodName, podServiceName, port, peerNamespace, protocol)
// 				var NetworkGraph = models.NetworkGraph{
// 					NData: networkData,
// 					ID:    ID,
// 					Source: models.NodeFields{
// 						ID:       fmt.Sprintf("%s%s", podServiceName, peerNamespace),
// 						Title:    fmt.Sprintf("%s", peerhostName),
// 						MainStat: fmt.Sprintf("%s", podServiceName),
//
// 						DetailPodName:       podServiceName,
// 						DetailNamespaceName: peerNamespace,
//
// 						Color: "white",
// 					},
// 					Target: node,
// 				}
// 				networkGraphs = append(networkGraphs, NetworkGraph)
// 				EdgeAcceptMap[ID] += 1
//
// 				break
// 			case "tcp_connect":
//
// 				ID1 := fmt.Sprintf("%s%s%s%s%s", podServiceName, log.PodName, port, peerNamespace, protocol)
// 				var NetworkGraph = models.NetworkGraph{
//
// 					NData:  networkData,
// 					ID:     ID1,
// 					Source: node,
// 					Target: models.NodeFields{
// 						ID:                  fmt.Sprintf("%s%s", podServiceName, peerNamespace),
// 						Title:               fmt.Sprintf("%s", peerhostName),
// 						MainStat:            fmt.Sprintf("%s", podServiceName),
// 						DetailPodName:       podServiceName,
// 						DetailNamespaceName: peerNamespace,
// 						Color:               "white",
// 					},
// 				}
// 				networkGraphs = append(networkGraphs, NetworkGraph)
//
// 				EdgeConnectMap[ID1] += 1
// 				break
// 			}
//
// 			// }
//
// 		} else if containsKprobe := strings.Contains(log.Data, "kprobe"); containsKprobe && !containsHostname {
//
// 			kprobeData := datamap["kprobe"]
// 			domainData := datamap["domain"]
//
// 			resourceMap := extractdata(log.Resource)
// 			remoteIP := resourceMap["remoteip"]
//
// 			port := resourceMap["port"]
// 			protocol := resourceMap["protocol"]
//
// 			networkData = models.NetworkData{
// 				NetworkType: "kprobe:" + kprobeData,
// 				SockType:    "",
// 				Kprobe:      kprobeData,
// 				Domain:      domainData,
// 				RemoteIP:    remoteIP,
// 				HostName:    "External",
// 				Port:        port,
// 				Protocol:    protocol,
// 			}
//
// 			node := models.NodeFields{
//
// 				ID:                  fmt.Sprintf("%s%s", log.PodName, log.NamespaceName),
// 				Title:               log.PodName,
// 				MainStat:            log.PodName,
// 				Color:               "white",
// 				DetailPodName:       log.PodName,
// 				DetailNamespaceName: log.NamespaceName,
// 			}
//
// 			if log.Result == denied {
// 				node.Color = "red"
// 			}
//
// 			switch kprobeData {
// 			case "tcp_accept":
// 				ID := fmt.Sprintf("%s%s%s%s%s", log.PodName, remoteIP, port, log.NamespaceName, protocol)
// 				var NetworkGraph = models.NetworkGraph{
// 					NData: networkData,
// 					ID:    ID,
// 					Source: models.NodeFields{
// 						ID:       fmt.Sprintf("%s%s", remoteIP, "external"),
// 						Title:    fmt.Sprintf("%s", remoteIP),
// 						MainStat: fmt.Sprintf("%s", "External "),
//
// 						DetailPodName:       "External",
// 						DetailNamespaceName: "external",
//
// 						Color: "white",
// 					},
// 					Target: node,
// 				}
// 				networkGraphs = append(networkGraphs, NetworkGraph)
// 				EdgeAcceptMap[ID] += 1
//
// 				break
// 			case "tcp_connect":
//
// 				ID1 := fmt.Sprintf("%s%s%s%s%s", log.PodName, remoteIP, port, log.NamespaceName, protocol)
// 				var NetworkGraph = models.NetworkGraph{
//
// 					NData:  networkData,
// 					ID:     ID1,
// 					Source: node,
// 					Target: models.NodeFields{
// 						ID:       fmt.Sprintf("%s%s", remoteIP, "external"),
// 						Title:    fmt.Sprintf("%s", remoteIP),
// 						MainStat: fmt.Sprintf("%s", "External "),
//
// 						DetailPodName:       "External",
// 						DetailNamespaceName: "external",
//
// 						Color: "white",
// 					},
// 				}
// 				networkGraphs = append(networkGraphs, NetworkGraph)
//
// 				EdgeConnectMap[ID1] += 1
// 				break
// 			}
//
// 		}
//
// 	}
//
// 	for key, value := range EdgeAcceptMap {
// 		EdgeMap[key] = value
// 	}
//
// 	// Merge EdgeConnectMap into EdgeMap
// 	for key, value := range EdgeConnectMap {
// 		if _, exists := EdgeMap[key]; exists {
// 			if EdgeMap[key] < EdgeConnectMap[key] {
// 				EdgeMap[key] = value
// 			}
// 		} else {
// 			EdgeMap[key] = value
// 		}
// 	}
//
// 	for _, netGraph := range networkGraphs {
// 		NodeData = append(NodeData, netGraph.Source)
// 		NodeData = append(NodeData, netGraph.Target)
//
// 		var edge = models.EdgeFields{
// 			ID:       netGraph.ID,
// 			Source:   netGraph.Source.ID,
// 			Target:   netGraph.Target.ID,
// 			Mainstat: netGraph.NData.Protocol + " " + netGraph.NData.Port,
// 			Count:    strconv.Itoa(EdgeMap[netGraph.ID]),
// 		}
// 		EdgeData = append(EdgeData, edge)
//
// 	}
//
// 	var nodeGraph = models.NodeGraph{
// 		Nodes: NodeData,
// 		Edges: EdgeData,
// 	}
//
// 	return nodeGraph
// }

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, _ := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	ctxLogger := log.DefaultLogger.FromContext(ctx)

	healthendpoint := d.settings.URL
	switch config.Backend {
	case "ELASTICSEARCH":
		healthendpoint += "/_cluster/health"
	case "LOKI":
		healthendpoint += "/ready"
	case "OPENSEARCH":
		// healthendpoint += "/_cluster/health"
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusOk,
			Message: fmt.Sprintf("Data source is working "),
		}, nil
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, healthendpoint, nil)

	resp, err := d.httpClient.Do(r)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			ctxLogger.Error("check health: failed to close response body", "err", err.Error())
		}
	}()

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		ctxLogger.Error("load settings: failed to load settings")
		return res, nil
	}

	if resp.StatusCode != http.StatusOK {
		res.Status = backend.HealthStatusError
		res.Message = fmt.Sprintf("error on checking health check status from backend  %s", resp.Status)

		return res, nil
	}

	if config.Secrets.ApiKey == "" {
		res.Status = backend.HealthStatusError
		res.Message = "API key is missing"
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Data source is working "),
	}, nil
}
