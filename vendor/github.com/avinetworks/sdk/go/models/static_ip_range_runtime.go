package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StaticIPRangeRuntime static Ip range runtime
// swagger:model StaticIpRangeRuntime
type StaticIPRangeRuntime struct {

	// Allocated IPs. Field introduced in 20.1.3.
	AllocatedIps []*StaticIPAllocInfo `json:"allocated_ips,omitempty"`

	// Free IP count. Field introduced in 20.1.3.
	FreeIPCount *int32 `json:"free_ip_count,omitempty"`

	// Total IP count. Field introduced in 20.1.3.
	TotalIPCount *int32 `json:"total_ip_count,omitempty"`

	// Object type (VIP only, Service Engine only, or both) which is using this IP group. Enum options - STATIC_IPS_FOR_SE, STATIC_IPS_FOR_VIP, STATIC_IPS_FOR_VIP_AND_SE. Field introduced in 20.1.3.
	Type *string `json:"type,omitempty"`
}
