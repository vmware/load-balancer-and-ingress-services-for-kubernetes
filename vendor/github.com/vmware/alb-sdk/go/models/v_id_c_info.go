package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIDCInfo v ID c info
// swagger:model VIDCInfo
type VIDCInfo struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// managed_object_id of VIDCInfo.
	// Required: true
	ManagedObjectID *string `json:"managed_object_id"`

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
}
