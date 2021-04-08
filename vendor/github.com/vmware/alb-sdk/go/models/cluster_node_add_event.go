package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterNodeAddEvent cluster node add event
// swagger:model ClusterNodeAddEvent
type ClusterNodeAddEvent struct {

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Role of the controller within the cluster. Enum options - CLUSTER_LEADER, CLUSTER_FOLLOWER.
	Role *string `json:"role,omitempty"`
}
