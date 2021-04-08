package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerPortalRegistration controller portal registration
// swagger:model ControllerPortalRegistration
type ControllerPortalRegistration struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 18.2.6.
	Asset *ControllerPortalAsset `json:"asset,omitempty"`

	//  Field introduced in 18.2.6.
	// Required: true
	Name *string `json:"name"`

	//  Field introduced in 18.2.6.
	PortalAuth *ControllerPortalAuth `json:"portal_auth,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 18.2.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 18.2.6.
	UUID *string `json:"uuid,omitempty"`
}
