package models

import "github.com/grafana/grafana-plugin-sdk-go/data"

type NodeGraph struct {
	Nodes []NodeFields
	Edges []EdgeFields
}

type FrameFieldType struct {
	Name        string         `json:"name"`
	Type        data.FieldType `json:"type"`
	DisplayName string
}

type NodeFields struct {
	ID        string `json:"id"`
	Title     string `json:"title,omitempty"`
	MainStat  string `json:"mainStat,omitempty"`
	Color     string `json:"color,omitempty"`
	ChildNode string `json:"childNode,omitempty"`
	// NodeRadius string `json:"nodeRadius,omitempty"`
	// Highlighted bool   `json:"highlighted,omitempty"`

	DetailClusterName       string `json:"detail__ClusterName,omitempty"`
	DetailHostName          string `json:"detail__HostName,omitempty"`
	DetailNamespaceName     string `json:"detail__NamespaceName,omitempty"`
	DetailPodName           string `json:"detail__PodName,omitempty"`
	DetailLabels            string `json:"detail__Labels,omitempty"`
	DetailContainerID       string `json:"detail__ContainerID,omitempty"`
	DetailContainerName     string `json:"detail__ContainerName,omitempty"`
	DetailContainerImage    string `json:"detail__ContainerImage,omitempty"`
	DetailParentProcessName string `json:"detail__ParentProcessName,omitempty"`
	DetailProcessName       string `json:"detail__ProcessName,omitempty"`
	DetailHostPPID          int64  `json:"detail__HostPPID,omitempty"`
	DetailHostPID           int64  `json:"detail__HostPID,omitempty"`
	DetailPPID              int64  `json:"detail__PPID,omitempty"`
	DetailPID               int64  `json:"detail__PID,omitempty"`
	DetailUID               int64  `json:"detail__UID,omitempty"`
	DetailType              string `json:"detail__Type,omitempty"`
	DetailSource            string `json:"detail__Source,omitempty"`
	DetailOperation         string `json:"detail__Operation,omitempty"`
	DetailResource          string `json:"detail__Resource,omitempty"`
	DetailData              string `json:"detail__Data,omitempty"`
	DetailResult            string `json:"detail__Result,omitempty"`
	DetailCwd               string `json:"detail__Cwd,omitempty"`
	DetailTTY               string `json:"detail__TTY,omitempty"`

	DetailOwnerName string
}

type EdgeFields struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`

	Mainstat string
	Count    string
}

type NetworkData struct {
	NetworkType string `json:"networkType"`
	SockType    string `json:"socktype"`

	Kprobe   string `json:"kprobe,omitempty"`
	Domain   string `json:"domain,omitempty"`
	RemoteIP string `json:"remoteIP,omitempty"`
	HostName string `json:"remoteIP,omitempty"`
	Port     string `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type NetworkGraph struct {
	NData  NetworkData `json:"ndata,omitempty"`
	ID     string      `json:"id,omitempty"`
	Source NodeFields  `json:"source,omitempty"`
	Target NodeFields  `json:"target,omitempty"`
}
