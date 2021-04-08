package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbReplicationStats gslb replication stats
// swagger:model GslbReplicationStats
type GslbReplicationStats struct {

	// Last config version acknowledged. Field introduced in 20.1.1.
	AcknowledgedVersion *int64 `json:"acknowledged_version,omitempty"`

	// Number of pending objects. Field introduced in 20.1.1.
	PendingObjectCount *int64 `json:"pending_object_count,omitempty"`

	// Last config version received. On leader, this represents the version received from federated datastore. Field introduced in 20.1.1.
	ReceivedVersion *int64 `json:"received_version,omitempty"`
}
