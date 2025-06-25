package visualizations

import (
	// "context"
	"fmt"
	"time"

	"github.com/accuknox/kubearmor/pkg/models"
	// "github.com/accuknox/kubearmor/pkg/utils"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type TableFields struct {
	LogSource     string
	Namespace     string
	ContainerName string
	ProcessName   string
	Syscall       string
	Result        string
	Count         int
	TimeStamp     time.Time
}

func (v *Visualization) getProfileView() backend.DataResponse {

	var response backend.DataResponse

	logs, err := v.backendService.GetLogs(v.ctx, v.qm)
	if err != nil {
		response.Error = err
		return response
	}

	frame := makeAlertListFrame(logs, v.qm)

	response.Frames = append(response.Frames, frame)

	return response
}

func makeAlertListFrame(logs []types.Log, qm models.QueryModel) *data.Frame {

	aggregateCount := make(map[string]uint32)
	addOrIncrement := func(key string) {
		// Check if key exists
		if _, exists := aggregateCount[key]; exists {
			aggregateCount[key]++ // Increment if exists
		} else {
			aggregateCount[key] = 1 // Initialize if new
		}
	}
	n := len(logs)

	logSource := make([]string, n)
	nameSpace := make([]string, n)
	containerName := make([]string, n)
	processName := make([]string, n)
	resource := make([]string, n)
	result := make([]string, n)
	count := make([]uint32, n)
	timestamp := make([]string, n)

	for i, l := range logs {

		if !IsCorrectLog(l, qm) {
			continue
		}

		logType := "Container"
		if l.Type == "HostLog" {
			logType = "Host"
			l.NamespaceName = "--"
			l.ContainerName = "--"
		}
		logSource[i] = logType
		nameSpace[i] = l.NamespaceName
		containerName[i] = l.ContainerName
		processName[i] = l.ProcessName
		resource[i] = l.Resource
		result[i] = l.Result

		key := makeKeyFromEntry(l)
		addOrIncrement(key)
		count[i] = aggregateCount[key]

		timestamp[i] = l.UpdatedTime

		if l.Operation == "Syscall" {
			resource[i] = l.Data
		}

	}

	frame := data.NewFrame("KubeArmor Alerts",
		data.NewField("LogSource", nil, logSource),
		data.NewField("Namespace", nil, nameSpace),
		data.NewField("ContainerName", nil, containerName),
		data.NewField("ProcessName", nil, processName),
		data.NewField(qm.Operation, nil, resource),
		data.NewField("Result", nil, result),
		// data.NewField("Count", nil, count),
		data.NewField("TimeStamp", nil, timestamp),
	)

	frame.Meta = &data.FrameMeta{Type: data.FrameTypeLogLines}
	return frame
}

func IsCorrectLog(log types.Log, qm models.QueryModel) bool {

	if log.Operation != qm.Operation {
		return false
	}

	if qm.NamespaceQuery != "All" && log.NamespaceName != qm.NamespaceQuery {
		return false
	}

	if qm.LabelQuery != "All" && log.Labels != qm.LabelQuery {
		return false
	}

	return true
}

func makeKeyFromEntry(e types.Log) string {
	return fmt.Sprintf("%s|%s|%s", e.NamespaceName, e.ContainerName, e.ProcessName)
}
