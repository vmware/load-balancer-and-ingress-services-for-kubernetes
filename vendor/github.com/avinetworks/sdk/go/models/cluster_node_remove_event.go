package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterNodeRemoveEvent cluster node remove event
// swagger:model ClusterNodeRemoveEvent
type ClusterNodeRemoveEvent struct {

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Role of the node when it left the controller cluster. Enum options - CLUSTER_LEADER, CLUSTER_FOLLOWER.
	Role *string `json:"role,omitempty"`
}
