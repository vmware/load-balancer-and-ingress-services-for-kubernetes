package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ReplicationPolicy replication policy
// swagger:model ReplicationPolicy
type ReplicationPolicy struct {

	// Leader's checkpoint. Follower attempt to replicate configuration till this checkpoint. It is a reference to an object of type FederationCheckpoint. Field introduced in 20.1.1.
	CheckpointRef *string `json:"checkpoint_ref,omitempty"`

	// Replication mode. Enum options - REPLICATION_MODE_CONTINUOUS, REPLICATION_MODE_MANUAL, REPLICATION_MODE_ADAPTIVE. Field introduced in 20.1.1.
	ReplicationMode *string `json:"replication_mode,omitempty"`
}
