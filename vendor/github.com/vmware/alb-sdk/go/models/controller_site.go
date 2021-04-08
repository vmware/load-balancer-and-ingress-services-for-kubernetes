package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerSite controller site
// swagger:model ControllerSite
type ControllerSite struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// IP Address or a DNS resolvable, fully qualified domain name of the Site Controller Cluster. Field introduced in 18.2.5.
	// Required: true
	Address *string `json:"address"`

	// Name for the Site Controller Cluster. Field introduced in 18.2.5.
	// Required: true
	Name *string `json:"name"`

	// The Controller Site Cluster's REST API port number. Allowed values are 1-65535. Field introduced in 18.2.5.
	Port *int32 `json:"port,omitempty"`

	// Reference for the Tenant. It is a reference to an object of type Tenant. Field introduced in 18.2.5.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Reference for the Site Controller Cluster. Field introduced in 18.2.5.
	UUID *string `json:"uuid,omitempty"`
}
