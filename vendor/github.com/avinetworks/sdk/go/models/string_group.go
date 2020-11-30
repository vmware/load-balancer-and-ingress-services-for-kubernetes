package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StringGroup *string group
// swagger:model StringGroup
type StringGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Configure Key Value in the *string group.
	Kv []*KeyValue `json:"kv,omitempty"`

	// Key value pairs for granular object access control. Also allows for classification and tagging of similar objects. Field introduced in 20.1.3.
	Labels []*KeyValue `json:"labels,omitempty"`

	// Enable the longest match, default is the shortest match. Field introduced in 18.2.8.
	LongestMatch *bool `json:"longest_match,omitempty"`

	// Name of the *string group.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Type of StringGroup. Enum options - SG_TYPE_STRING, SG_TYPE_KEYVAL.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the *string group.
	UUID *string `json:"uuid,omitempty"`
}
