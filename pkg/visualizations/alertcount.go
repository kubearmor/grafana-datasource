package visualizations

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type AlertCounts struct {
	file       uint16
	process    uint16
	network    uint16
	capability uint16
	blocked    uint16
	total      uint16
}

func getAlertCount(alerts []types.Log) AlertCounts {
	alertCounts := AlertCounts{
		file:       0,
		process:    0,
		network:    0,
		capability: 0,
		blocked:    0,
		total:      uint16(len(alerts)),
	}
	for _, log := range alerts {
		switch log.Operation {
		case "File":
			alertCounts.file += 1

		case "Process":
			alertCounts.process += 1

		case "Network":
			alertCounts.network += 1
		}

		if log.Action == "Block" {
			alertCounts.blocked += 1
		}

	}
	return alertCounts
}

func (v *Visualization) getAlertCountGraph() backend.DataResponse {

	var response backend.DataResponse

	logs, err := v.backendService.GetLogs(v.ctx, v.qm)
	if err != nil {
		response.Error = err
		return response
	}

	alertcount := getAlertCount(logs)
	frame := makeGaugeFrame(alertcount)

	response.Frames = append(response.Frames, frame)

	return response
}
func makeGaugeFrame(ac AlertCounts) *data.Frame {

	frame := data.NewFrame("alert_counts",
		data.NewField("File Alerts", nil, []uint16{ac.file}),
		data.NewField("Process Alerts", nil, []uint16{ac.process}),
		data.NewField("Network Alerts", nil, []uint16{ac.network}),
		data.NewField("Capability Alerts", nil, []uint16{ac.capability}),
		data.NewField("Blocked operations", nil, []uint16{ac.blocked}),
		data.NewField("Total Alerts", nil, []uint16{ac.total}),
	)

	return frame
}
