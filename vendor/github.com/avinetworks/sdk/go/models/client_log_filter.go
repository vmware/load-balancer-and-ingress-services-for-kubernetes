package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClientLogFilter client log filter
// swagger:model ClientLogFilter
type ClientLogFilter struct {

	// Placeholder for description of property all_headers of obj type ClientLogFilter field type str  type boolean
	AllHeaders *bool `json:"all_headers,omitempty"`

	// Placeholder for description of property client_ip of obj type ClientLogFilter field type str  type object
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	//  Special values are 0 - 'infinite'.
	Duration *int32 `json:"duration,omitempty"`

	// Placeholder for description of property enabled of obj type ClientLogFilter field type str  type boolean
	// Required: true
	Enabled *bool `json:"enabled"`

	// Number of index.
	// Required: true
	Index *int32 `json:"index"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property uri of obj type ClientLogFilter field type str  type object
	URI *StringMatch `json:"uri,omitempty"`
}
