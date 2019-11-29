package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GcpInfo gcp info
// swagger:model GcpInfo
type GcpInfo struct {

	// Hostname of this SE.
	Hostname *string `json:"hostname,omitempty"`

	// Network this SE is assigned.
	// Required: true
	Network *string `json:"network"`

	// Project this SE belongs to.
	// Required: true
	Project *string `json:"project"`

	// Subnet assigned to this SE.
	Subnet *string `json:"subnet,omitempty"`

	// Zone this SE is part of.
	// Required: true
	Zone *string `json:"zone"`
}
