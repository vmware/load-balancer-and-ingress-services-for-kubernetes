package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotMapping bot mapping
// swagger:model BotMapping
type BotMapping struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Rules for bot classification. Field introduced in 21.1.1. Minimum of 1 items required.
	MappingRules []*BotMappingRule `json:"mapping_rules,omitempty"`

	// The name of this mapping. Field introduced in 21.1.1.
	// Required: true
	Name *string `json:"name"`

	// The unique identifier of the tenant to which this mapping belongs. It is a reference to an object of type Tenant. Field introduced in 21.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// A unique identifier of this mapping. Field introduced in 21.1.1.
	UUID *string `json:"uuid,omitempty"`
}
