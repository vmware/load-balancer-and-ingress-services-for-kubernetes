// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterNodeDbFailedEvent cluster node db failed event
// swagger:model ClusterNodeDbFailedEvent
type ClusterNodeDbFailedEvent struct {

	// Number of failures.
	FailureCount *int32 `json:"failure_count,omitempty"`

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`
}
