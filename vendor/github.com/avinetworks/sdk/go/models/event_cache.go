package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// EventCache event cache
// swagger:model EventCache
type EventCache struct {

	// Placeholder for description of property dns_state of obj type EventCache field type str  type boolean
	DNSState *bool `json:"dns_state,omitempty"`

	// Cache the exception strings in the system.
	Exceptions []string `json:"exceptions,omitempty"`
}
