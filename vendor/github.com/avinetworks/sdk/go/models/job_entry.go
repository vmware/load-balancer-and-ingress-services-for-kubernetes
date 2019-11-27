package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JobEntry job entry
// swagger:model JobEntry
type JobEntry struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// expires_at of JobEntry.
	// Required: true
	ExpiresAt *string `json:"expires_at"`

	//  Field introduced in 18.1.2.
	// Required: true
	Name *string `json:"name"`

	// obj_key of JobEntry.
	// Required: true
	ObjKey *string `json:"obj_key"`

	//  Field introduced in 18.1.1.
	Subjobs []*SubJob `json:"subjobs,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
