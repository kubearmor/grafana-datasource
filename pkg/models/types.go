package models

type QueryModel struct {
	NamespaceQuery string `json:"NamespaceQuery,omitempty"`
	LabelQuery     string `json:"LabelQuery,omitempty"`
	Operation      string `json:"Operation"`
	BatchSize      int    `json:"BatchSize"`
}

type DataStoreConfig struct {
	Username string
	Password string
	URL      string
}
