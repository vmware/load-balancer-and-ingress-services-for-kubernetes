package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StaticIPAllocInfo static Ip alloc info
// swagger:model StaticIpAllocInfo
type StaticIPAllocInfo struct {

	// IP address. Field introduced in 20.1.3.
	// Required: true
	IP *IPAddr `json:"ip"`

	// Object metadata. Field introduced in 20.1.3.
	ObjInfo *string `json:"obj_info,omitempty"`

	// Object which this IP address is allocated to. Field introduced in 20.1.3.
	ObjUUID *string `json:"obj_uuid,omitempty"`
}
