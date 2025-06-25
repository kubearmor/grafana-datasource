package visualizations

import (
	// "context"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (v *Visualization) getAlertList() backend.DataResponse {

	var response backend.DataResponse

	logs, err := v.backendService.GetLogs(v.ctx, v.qm)
	if err != nil {
		response.Error = err
		return response
	}

	frame := makeAlertListFrame(logs)

	response.Frames = append(response.Frames, frame)

	return response
}

func makeAlertListFrame(alerts []types.Log) *data.Frame {
	n := len(alerts)
	times := make([]time.Time, n)
	namespaces := make([]string, n)
	pods := make([]string, n)
	types := make([]string, n)
	messages := make([]string, n)

	for i, a := range alerts {

		// Parse the string using RFC3339Nano layout
		t, err := time.Parse(time.RFC3339Nano, a.UpdatedTime)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}
		times[i] = t
		namespaces[i] = a.NamespaceName
		pods[i] = a.PodName
		types[i] = a.Type
		messages[i] = a.Message
	}
	frame := data.NewFrame("kubearmor_alerts",
		data.NewField(data.TimeSeriesTimeFieldName, nil, times),
		data.NewField("namespace", nil, namespaces),
		data.NewField("pod", nil, pods),
		data.NewField("alert_type", nil, types),
		data.NewField("message", nil, messages),
	)
	frame.Meta = &data.FrameMeta{Type: data.FrameTypeLogLines}
	return frame
}
