// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ReplicationPolicy replication policy
// swagger:model ReplicationPolicy
type ReplicationPolicy struct {

	// Leader's checkpoint. Follower attempt to replicate configuration till this checkpoint. It is a reference to an object of type FederationCheckpoint. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CheckpointRef *string `json:"checkpoint_ref,omitempty"`

	// Replication mode. Enum options - REPLICATION_MODE_CONTINUOUS, REPLICATION_MODE_MANUAL, REPLICATION_MODE_ADAPTIVE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ReplicationMode *string `json:"replication_mode,omitempty"`
}
