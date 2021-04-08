package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkFilter network filter
// swagger:model NetworkFilter
type NetworkFilter struct {

	//  It is a reference to an object of type VIMgrNWRuntime.
	// Required: true
	NetworkRef *string `json:"network_ref"`

	// server_filter of NetworkFilter.
	ServerFilter *string `json:"server_filter,omitempty"`
}
