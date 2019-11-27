package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterNodeShutdownEvent cluster node shutdown event
// swagger:model ClusterNodeShutdownEvent
type ClusterNodeShutdownEvent struct {

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Reason for controller node shutdown.
	Reason *string `json:"reason,omitempty"`
}
