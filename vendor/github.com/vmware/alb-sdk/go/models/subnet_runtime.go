package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SubnetRuntime subnet runtime
// swagger:model SubnetRuntime
type SubnetRuntime struct {

	// Moved to StaticIpRangeRuntime. Field deprecated in 20.1.3.
	FreeIPCount *int32 `json:"free_ip_count,omitempty"`

	// Use allocated_ips in StaticIpRangeRuntime. Field deprecated in 20.1.3.
	IPAlloced []*IPAllocInfo `json:"ip_alloced,omitempty"`

	// Static IP range runtime. Field introduced in 20.1.3.
	IPRangeRuntimes []*StaticIPRangeRuntime `json:"ip_range_runtimes,omitempty"`

	// Placeholder for description of property prefix of obj type SubnetRuntime field type str  type object
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// Moved to StaticIpRangeRuntime. Field deprecated in 20.1.3.
	TotalIPCount *int32 `json:"total_ip_count,omitempty"`

	// Can be derived from total - free in StaticIpRangeRuntime. Field deprecated in 20.1.3.
	UsedIPCount *int32 `json:"used_ip_count,omitempty"`
}
