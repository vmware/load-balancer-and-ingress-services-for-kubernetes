package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPPolicySet HTTP policy set
// swagger:model HTTPPolicySet
type HTTPPolicySet struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Checksum of cloud configuration for Pool. Internally set by cloud connector.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// HTTP request policy for the virtual service.
	HTTPRequestPolicy *HTTPRequestPolicy `json:"http_request_policy,omitempty"`

	// HTTP response policy for the virtual service.
	HTTPResponsePolicy *HTTPResponsePolicy `json:"http_response_policy,omitempty"`

	// HTTP security policy for the virtual service.
	HTTPSecurityPolicy *HttpsecurityPolicy `json:"http_security_policy,omitempty"`

	// Placeholder for description of property is_internal_policy of obj type HTTPPolicySet field type str  type boolean
	IsInternalPolicy *bool `json:"is_internal_policy,omitempty"`

	// Name of the HTTP Policy Set.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the HTTP Policy Set.
	UUID *string `json:"uuid,omitempty"`
}
