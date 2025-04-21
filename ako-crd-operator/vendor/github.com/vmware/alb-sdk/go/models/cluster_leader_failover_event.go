// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterLeaderFailoverEvent cluster leader failover event
// swagger:model ClusterLeaderFailoverEvent
type ClusterLeaderFailoverEvent struct {

	// Details of the new controller cluster leader node. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LeaderNode *ClusterNode `json:"leader_node,omitempty"`

	// Details of the previous controller cluster leader. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreviousLeaderNode *ClusterNode `json:"previous_leader_node,omitempty"`
}
