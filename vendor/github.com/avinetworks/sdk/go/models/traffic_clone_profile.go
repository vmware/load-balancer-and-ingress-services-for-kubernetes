package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TrafficCloneProfile traffic clone profile
// swagger:model TrafficCloneProfile
type TrafficCloneProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.1.1.
	CloneServers []*CloneServer `json:"clone_servers,omitempty"`

	//  It is a reference to an object of type Cloud. Field introduced in 17.1.1.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Name for the Traffic Clone Profile. Field introduced in 17.1.1.
	// Required: true
	Name *string `json:"name"`

	// Specifies if client IP needs to be preserved to clone destination. Field introduced in 17.1.1.
	PreserveClientIP *bool `json:"preserve_client_ip,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Traffic Clone Profile. Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`
}
