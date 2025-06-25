package models

// Visualizations for KubeArmor
const (
	PROCESSGRAPH    = "PROCESSGRAPH"
	NETWORKGRAPH    = "NETWORKGRAPH"
	ALERTCOUNTGRAPH = "ALERTCOUNTGRAPH"
	PROFILE         = "PROFILE"
	// FILEPROFILE     = "FILEPROFILE"
	// NETWORKPROFILE  = "NETWORKPROFILE"
	// SYSCALLPROFILE  = "SYSCALLPROFILE"
	// ALERTLIST       = "ALERTLIST"
)

/*
Operation types for kubearmor
*/
const (
	OPERATIONPROCESS = "PROCESS"
	OPERATIONFILE    = "FILE"
	OPERATIONNETWORK = "NETWORK"
	OPERATIONSYSCALL = "SYSCALL"
)

// Define the equivalent Go representation of NodeframeFields
