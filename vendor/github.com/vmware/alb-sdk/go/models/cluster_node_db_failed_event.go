// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterNodeDbFailedEvent cluster node db failed event
// swagger:model ClusterNodeDbFailedEvent
type ClusterNodeDbFailedEvent struct {

	// Number of failures. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FailureCount *uint32 `json:"failure_count,omitempty"`

	// IPv4 address of the controller VM. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP *IPAddr `json:"ip,omitempty"`

	// IPv6 address of the controller VM. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6 *IPAddr `json:"ip6,omitempty"`

	// Name of controller node. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NodeName *string `json:"node_name,omitempty"`
}
