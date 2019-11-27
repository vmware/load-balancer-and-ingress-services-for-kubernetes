package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkRuntime network runtime
// swagger:model NetworkRuntime
type NetworkRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Unique object identifier of se.
	SeUUID []string `json:"se_uuid,omitempty"`

	// Placeholder for description of property subnet_runtime of obj type NetworkRuntime field type str  type object
	SubnetRuntime []*SubnetRuntime `json:"subnet_runtime,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
