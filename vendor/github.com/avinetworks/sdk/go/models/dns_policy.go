package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSPolicy Dns policy
// swagger:model DnsPolicy
type DNSPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Creator name. Field introduced in 17.1.1.
	CreatedBy *string `json:"created_by,omitempty"`

	//  Field introduced in 17.1.1.
	Description *string `json:"description,omitempty"`

	// Name of the DNS Policy. Field introduced in 17.1.1.
	// Required: true
	Name *string `json:"name"`

	// DNS rules. Field introduced in 17.1.1.
	Rule []*DNSRule `json:"rule,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the DNS Policy. Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`
}
