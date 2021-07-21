// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ConfigDeleteDetails config delete details
// swagger:model ConfigDeleteDetails
type ConfigDeleteDetails struct {

	// Error message if request failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// API path.
	Path *string `json:"path,omitempty"`

	// Deleted data of the resource.
	ResourceData *string `json:"resource_data,omitempty"`

	// Name of the deleted resource.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the deleted resource.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status.
	Status *string `json:"status,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
