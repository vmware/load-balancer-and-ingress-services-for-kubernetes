// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PriorityLabels priority labels
// swagger:model PriorityLabels
type PriorityLabels struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// A description of the priority labels.
	Description *string `json:"description,omitempty"`

	// Equivalent priority labels in descending order.
	EquivalentLabels []*EquivalentLabels `json:"equivalent_labels,omitempty"`

	// Key value pairs for granular object access control. Also allows for classification and tagging of similar objects. Field deprecated in 20.1.5. Field introduced in 20.1.2. Maximum of 4 items allowed.
	Labels []*KeyValue `json:"labels,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Basic edition, Essentials edition, Enterprise edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// The name of the priority labels.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the priority labels.
	UUID *string `json:"uuid,omitempty"`
}
