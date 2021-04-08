package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSSrvRdata Dns srv rdata
// swagger:model DnsSrvRdata
type DNSSrvRdata struct {

	// Service port. Allowed values are 0-65535.
	// Required: true
	Port *int32 `json:"port"`

	// Priority of the target hosting the service, low value implies higher priority for this service record. Allowed values are 0-65535.
	Priority *int32 `json:"priority,omitempty"`

	// Canonical hostname, of the machine hosting the service, with no trailing period. 'default.host' is valid but not 'default.host.'.
	Target *string `json:"target,omitempty"`

	// Relative weight for service records with same priority, high value implies higher preference for this service record. Allowed values are 0-65535.
	Weight *int32 `json:"weight,omitempty"`
}
