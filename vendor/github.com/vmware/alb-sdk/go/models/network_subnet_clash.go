package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSubnetClash network subnet clash
// swagger:model NetworkSubnetClash
type NetworkSubnetClash struct {

	// ip_nw of NetworkSubnetClash.
	// Required: true
	IPNw *string `json:"ip_nw"`

	// networks of NetworkSubnetClash.
	Networks []string `json:"networks,omitempty"`
}
