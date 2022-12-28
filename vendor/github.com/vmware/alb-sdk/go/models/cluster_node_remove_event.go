// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterNodeRemoveEvent cluster node remove event
// swagger:model ClusterNodeRemoveEvent
type ClusterNodeRemoveEvent struct {

	// IP address of the controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeName *string `json:"node_name,omitempty"`

	// Role of the node when it left the controller cluster. Enum options - CLUSTER_LEADER, CLUSTER_FOLLOWER. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Role *string `json:"role,omitempty"`
}
