package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
