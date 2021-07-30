// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ClusterServiceFailedEvent cluster service failed event
// swagger:model ClusterServiceFailedEvent
type ClusterServiceFailedEvent struct {

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Name of the controller service.
	ServiceName *string `json:"service_name,omitempty"`
}
