// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterNodeShutdownEvent cluster node shutdown event
// swagger:model ClusterNodeShutdownEvent
type ClusterNodeShutdownEvent struct {

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Reason for controller node shutdown.
	Reason *string `json:"reason,omitempty"`
}
