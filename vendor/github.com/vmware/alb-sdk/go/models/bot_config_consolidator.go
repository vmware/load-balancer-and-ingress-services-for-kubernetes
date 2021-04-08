package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotConfigConsolidator bot config consolidator
// swagger:model BotConfigConsolidator
type BotConfigConsolidator struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Human-readable description of this consolidator. Field introduced in 21.1.1.
	Description *string `json:"description,omitempty"`

	// The name of this consolidator. Field introduced in 21.1.1.
	// Required: true
	Name *string `json:"name"`

	// Script that consolidates results from all components. Field introduced in 21.1.1.
	Script *string `json:"script,omitempty"`

	// The unique identifier of the tenant to which this mapping belongs. It is a reference to an object of type Tenant. Field introduced in 21.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// A unique identifier to this consolidator. Field introduced in 21.1.1.
	UUID *string `json:"uuid,omitempty"`
}
