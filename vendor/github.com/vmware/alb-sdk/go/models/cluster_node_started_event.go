// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterNodeStartedEvent cluster node started event
// swagger:model ClusterNodeStartedEvent
type ClusterNodeStartedEvent struct {

	// IP address of the controller VM.
	IP *IPAddr `json:"ip,omitempty"`

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`
}
