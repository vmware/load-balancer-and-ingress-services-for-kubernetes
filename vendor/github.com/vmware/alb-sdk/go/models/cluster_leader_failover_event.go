package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterLeaderFailoverEvent cluster leader failover event
// swagger:model ClusterLeaderFailoverEvent
type ClusterLeaderFailoverEvent struct {

	// Details of the new controller cluster leader node.
	LeaderNode *ClusterNode `json:"leader_node,omitempty"`

	// Details of the previous controller cluster leader.
	PreviousLeaderNode *ClusterNode `json:"previous_leader_node,omitempty"`
}
