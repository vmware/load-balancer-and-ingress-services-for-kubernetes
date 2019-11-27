package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CustomIPAMDNSProfile custom ipam Dns profile
// swagger:model CustomIpamDnsProfile
type CustomIPAMDNSProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of the Custom IPAM DNS Profile. Field introduced in 17.1.1.
	// Required: true
	Name *string `json:"name"`

	// Parameters that are always passed to the IPAM/DNS script. Field introduced in 17.1.1.
	ScriptParams []*CustomParams `json:"script_params,omitempty"`

	// Script URI of form controller //ipamdnsscripts/<file-name>. Field introduced in 17.1.1.
	// Required: true
	ScriptURI *string `json:"script_uri"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`
}
