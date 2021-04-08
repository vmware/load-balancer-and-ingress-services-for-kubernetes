package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterNodeDbFailedEvent cluster node db failed event
// swagger:model ClusterNodeDbFailedEvent
type ClusterNodeDbFailedEvent struct {

	// Number of failures.
	FailureCount *int32 `json:"failure_count,omitempty"`

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`
}
