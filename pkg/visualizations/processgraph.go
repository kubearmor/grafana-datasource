package visualizations

import (
	"fmt"

	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/accuknox/kubearmor/pkg/utils"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
)

const (
	pts0   = "pts0"
	denied = "Permission denied"
)

func (v *Visualization) getProcessGraph() backend.DataResponse {
	var response backend.DataResponse

	Nodefields := GetProcessNodeFields()
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

	nodegraph := getProcessTree(logs, v.qm)

	for _, node := range nodegraph.Nodes {

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
	}

	for _, edge := range nodegraph.Edges {
		EdgeFrame.AppendRow(edge.ID, edge.Source, edge.Target, edge.Mainstat, edge.Count)
	}

	response.Frames = append(response.Frames, Nodeframe)
	response.Frames = append(response.Frames, EdgeFrame)
	return response
}

func getProcessTree(logs []types.Log, MyQuery models.QueryModel) models.NodeGraph {

	colors := []string{"orange", "green", "cyan", "rose"}

	var processLogs []types.Log

	for _, log := range logs {

		if log.TTY == pts0 && log.Operation == "Process" &&
			(MyQuery.NamespaceQuery == "All" || log.NamespaceName == MyQuery.NamespaceQuery) &&
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
			colorIndex := utils.Random(0, len(colors)-1)
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
