package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafProfile waf profile
// swagger:model WafProfile
type WafProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Config params for WAF. Field introduced in 17.2.1.
	// Required: true
	Config *WafConfig `json:"config"`

	//  Field introduced in 17.2.1.
	Description *string `json:"description,omitempty"`

	// List of Data Files Used for WAF Rules. Field introduced in 17.2.1.
	Files []*WafDataFile `json:"files,omitempty"`

	//  Field introduced in 17.2.1.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.2.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.2.1.
	UUID *string `json:"uuid,omitempty"`
}
