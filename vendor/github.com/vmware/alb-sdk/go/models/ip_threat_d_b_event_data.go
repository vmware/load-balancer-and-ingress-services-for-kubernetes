package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPThreatDBEventData IP threat d b event data
// swagger:model IPThreatDBEventData
type IPThreatDBEventData struct {

	// Reason for IPThreatDb transaction failure. Field introduced in 20.1.1.
	Reason *string `json:"reason,omitempty"`

	// Status of IPThreatDb transaction. Field introduced in 20.1.1.
	Status *string `json:"status,omitempty"`

	// Last synced version of the IPThreatDB. Field introduced in 20.1.1.
	Version *string `json:"version,omitempty"`
}
