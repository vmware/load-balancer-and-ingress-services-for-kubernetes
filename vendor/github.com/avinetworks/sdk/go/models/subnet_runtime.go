package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SubnetRuntime subnet runtime
// swagger:model SubnetRuntime
type SubnetRuntime struct {

	// Number of free_ip_count.
	FreeIPCount *int32 `json:"free_ip_count,omitempty"`

	// Placeholder for description of property ip_alloced of obj type SubnetRuntime field type str  type object
	IPAlloced []*IPAllocInfo `json:"ip_alloced,omitempty"`

	// Placeholder for description of property prefix of obj type SubnetRuntime field type str  type object
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// Number of total_ip_count.
	TotalIPCount *int32 `json:"total_ip_count,omitempty"`

	// Number of used_ip_count.
	UsedIPCount *int32 `json:"used_ip_count,omitempty"`
}
