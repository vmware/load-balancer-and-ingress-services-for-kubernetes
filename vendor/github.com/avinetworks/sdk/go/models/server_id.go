package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerID server Id
// swagger:model ServerId
type ServerID struct {

	// This is the external cloud uuid of the Pool server.
	ExternalUUID *string `json:"external_uuid,omitempty"`

	// Placeholder for description of property ip of obj type ServerId field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	// Number of port.
	// Required: true
	Port *int32 `json:"port"`
}
