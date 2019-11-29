package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PersistenceFilter persistence filter
// swagger:model PersistenceFilter
type PersistenceFilter struct {

	// Persistence cookie.
	PersistenceCookie *string `json:"persistence_cookie,omitempty"`

	// Placeholder for description of property persistence_end_ip of obj type PersistenceFilter field type str  type object
	PersistenceEndIP *IPAddr `json:"persistence_end_ip,omitempty"`

	// Placeholder for description of property persistence_ip of obj type PersistenceFilter field type str  type object
	PersistenceIP *IPAddr `json:"persistence_ip,omitempty"`

	// Number of persistence_mask.
	PersistenceMask *int32 `json:"persistence_mask,omitempty"`

	// Placeholder for description of property server_end_ip of obj type PersistenceFilter field type str  type object
	ServerEndIP *IPAddr `json:"server_end_ip,omitempty"`

	// Placeholder for description of property server_ip of obj type PersistenceFilter field type str  type object
	ServerIP *IPAddr `json:"server_ip,omitempty"`

	// Number of server_mask.
	ServerMask *int32 `json:"server_mask,omitempty"`

	// Number of server_port.
	ServerPort *int32 `json:"server_port,omitempty"`
}
