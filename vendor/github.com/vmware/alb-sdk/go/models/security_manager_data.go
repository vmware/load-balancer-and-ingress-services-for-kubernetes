package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecurityManagerData security manager data
// swagger:model SecurityManagerData
type SecurityManagerData struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Information about various applications. Field introduced in 20.1.1.
	AppLearningInfo []*DbAppLearningInfo `json:"app_learning_info,omitempty"`

	// Virtualservice Name. Field introduced in 20.1.1.
	// Required: true
	Name *string `json:"name"`

	//  It is a reference to an object of type Tenant. Field introduced in 20.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Virtualservice UUID. Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`
}
