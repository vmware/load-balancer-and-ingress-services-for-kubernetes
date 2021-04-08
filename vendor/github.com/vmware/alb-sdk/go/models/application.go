package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Application application
// swagger:model Application
type Application struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VirtualserviceRefs []string `json:"virtualservice_refs,omitempty"`
}
