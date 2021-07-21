// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LabelGroup label group
// swagger:model LabelGroup
type LabelGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// List of allowed or suggested labels for the label group. Field introduced in 20.1.5.
	Labels []*RoleMatchOperationMatchLabel `json:"labels,omitempty"`

	// Name of the Label Group. Field introduced in 20.1.5.
	// Required: true
	Name *string `json:"name"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Label Group. Field introduced in 20.1.5.
	UUID *string `json:"uuid,omitempty"`
}
