package models

type QueryModel struct {
	NamespaceQuery string `json:"NamespaceQuery,omitempty"`
	LabelQuery     string `json:"LabelQuery,omitempty"`
	Operation      string `json:"Operation"`
	BatchSize      int    `json:"BatchSize"`
}

type DataStoreConfig struct {
	Index    string
	Username string `json:"basicAuthUser"` // basicAuthUser
	Password string `json:"-"`             // Basicsuth Password
	URL      string `json:"-"`             // Datasource URL

	// TLS settings
	TLSAuth           bool   `json:"tlsAuth"`           // TLS Client Auth
	TLSAuthWithCACert bool   `json:"tlsAuthWithCACert"` // With CA Cert
	TLSSkipVerify     bool   `json:"tlsSkipVerify"`     // Skip TLS Verify
	CACert            []byte `json:"-"`                 //CA cert

	// OAuth forwarding
	ForwardOAuthIdentity bool `json:"forwardOauthIdentity"` // Forward OAuth Identity

	// Credentials forwarding
	WithCredentials bool `json:"withCredentials"` // With Credentials
}
