package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAllocInfo Ip alloc info
// swagger:model IpAllocInfo
type IPAllocInfo struct {

	// Placeholder for description of property ip of obj type IpAllocInfo field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	// mac of IpAllocInfo.
	// Required: true
	Mac *string `json:"mac"`

	// Unique object identifier of se.
	// Required: true
	SeUUID *string `json:"se_uuid"`
}
