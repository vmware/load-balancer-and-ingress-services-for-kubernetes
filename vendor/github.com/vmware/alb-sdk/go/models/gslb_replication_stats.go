// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbReplicationStats gslb replication stats
// swagger:model GslbReplicationStats
type GslbReplicationStats struct {

	// Last config version acknowledged. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AcknowledgedVersion uint64 `json:"acknowledged_version,omitempty"`

	// Number of pending objects. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PendingObjectCount uint64 `json:"pending_object_count,omitempty"`

	// Last config version received. On leader, this represents the version received from federated datastore. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReceivedVersion uint64 `json:"received_version,omitempty"`
}
