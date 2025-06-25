package visualizations

import (
	"fmt"
	"net"
	"strings"

	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
)

const (
	HOST_UNKNOWN = "HOST_UNKNOWN"
	ACCEPT       = "ACCEPT"
	CONNECT      = "CONNECT"
	POD          = "POD"
	SERVICE      = "SERVICE"
)

func (v *Visualization) getNetworkGraph() backend.DataResponse {
	var response backend.DataResponse

	Nodefields := GetNetworkNodeFields()
	EdgeFields := GetEdgeFields()

	Nodeframe := data.NewFrame("Nodes")
	Nodeframe.Fields = Nodefields

	EdgeFrame := data.NewFrame("Edges")
	EdgeFrame.Fields = EdgeFields

	var frameMeta = data.FrameMeta{
		PreferredVisualization: data.VisTypeNodeGraph,
	}
	Nodeframe.SetMeta(&frameMeta)
	EdgeFrame.SetMeta(&frameMeta)

	logs, err := v.backendService.GetLogs(v.ctx, v.qm)
	if err != nil {
		response.Error = err
		return response
	}

	nodegraph := getNetworkTree(logs, v.qm)

	for _, node := range nodegraph.Nodes {

		Nodeframe.AppendRow(
			node.ID,
			node.Title,
			node.MainStat,
			node.Color,
			node.DetailNamespaceName,
			node.DetailPodName,
		)
	}

	for _, edge := range nodegraph.Edges {
		EdgeFrame.AppendRow(edge.ID, edge.Source, edge.Target, edge.Mainstat, edge.Count)
	}

	response.Frames = append(response.Frames, Nodeframe)
	response.Frames = append(response.Frames, EdgeFrame)
	return response
}

func getNetworkTree(logs []types.Log, MyQuery models.QueryModel) models.NodeGraph {

	var NetworkLogs []types.Log

	for _, log := range logs {

		if log.Operation == "Network" && (MyQuery.NamespaceQuery == "All" || log.NamespaceName == MyQuery.NamespaceQuery) &&
			(MyQuery.LabelQuery == "All" || log.Labels == MyQuery.LabelQuery) {
			NetworkLogs = append(NetworkLogs, log)
		}

	}

	var networkNodes []models.NodeFields

	var networkEdges []models.EdgeFields

	// nodeMap := make(map[string]interface{})

	for _, log := range NetworkLogs {
		netinfo, err := extractNetworkInfo(log)
		if err != nil {
			continue
		}

		node1 := models.NodeFields{
			ID:                  fmt.Sprintf("pod/%s/%s", log.NamespaceName, log.PodName),
			Title:               fmt.Sprintf("pod/%s/%s", log.NamespaceName, log.PodName),
			MainStat:            fmt.Sprintf("type:POD"),
			DetailPodName:       log.PodName,
			DetailNamespaceName: log.NamespaceName,
			Color:               "white",
		}
		networkNodes = append(networkNodes, node1)

		node2 := models.NodeFields{
			ID:                  netinfo.IP,
			Title:               netinfo.remoteHost,
			MainStat:            netinfo.IP,
			DetailPodName:       netinfo.remoteHost,
			DetailNamespaceName: netinfo.remoteHost,
			Color:               "white",
		}

		networkNodes = append(networkNodes, node2)

		if netinfo.remoteHost != "HOST_UNKNOWN" {

			node2type := "POD"
			if strings.Contains(netinfo.remoteHost, "svc") {
				node2type = "SERVICE"
			}
			splithost := strings.Split(netinfo.remoteHost, "/")
			node2Namespace := splithost[2]
			node2name := splithost[1]
			node2.MainStat = fmt.Sprintf("type:%s", node2type)
			node2.DetailNamespaceName = node2Namespace
			node2.DetailPodName = node2name
		}
		edge := models.EdgeFields{
			ID:       fmt.Sprintf("%s%s", netinfo.IP, netinfo.port),
			Source:   node1.ID,
			Target:   node2.ID,
			Mainstat: fmt.Sprintf("ip:%s port:%s", netinfo.IP, netinfo.port),
			Count:    "",
		}
		if netinfo.connectionType == ACCEPT {
			edge.Source = node2.ID
			edge.Target = node1.ID
		}
		networkEdges = append(networkEdges, edge)

	}

	var nodeGraph = models.NodeGraph{
		Nodes: networkNodes,
		Edges: networkEdges,
	}

	return nodeGraph

}

type NetworkInfo struct {
	IP             string
	port           string
	remoteHost     string
	connectionType string
}

func extractNetworkInfo(log types.Log) (NetworkInfo, error) {
	var netInfo NetworkInfo

	if strings.Contains(log.Data, "tcp_connect") {

		netInfo.connectionType = CONNECT
	} else if strings.Contains(log.Data, "tcp_accept") {
		netInfo.connectionType = ACCEPT
	} else {
		return NetworkInfo{}, fmt.Errorf("unknown log")
	}

	for _, field := range strings.Fields(log.Resource) {
		// Extract IP
		if strings.Contains(field, "remoteip=") {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 && net.ParseIP(parts[1]) != nil {
				netInfo.IP = parts[1]
			}
		}

		// Extract Remote host
		if strings.Contains(field, "remotehost=") {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				netInfo.remoteHost = parts[1]
			}
		} else {
			netInfo.remoteHost = HOST_UNKNOWN
		}

		// Extract Port
		if strings.Contains(field, "port=") {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				netInfo.port = parts[1]
			}
		}

	}

	return netInfo, nil
}
