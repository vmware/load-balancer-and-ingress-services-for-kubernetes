package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MicroServiceGroup micro service group
// swagger:model MicroServiceGroup
type MicroServiceGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Name of the MicroService group.
	// Required: true
	Name *string `json:"name"`

	// Configure MicroService(es). It is a reference to an object of type MicroService.
	ServiceRefs []string `json:"service_refs,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the MicroService group.
	UUID *string `json:"uuid,omitempty"`
}
