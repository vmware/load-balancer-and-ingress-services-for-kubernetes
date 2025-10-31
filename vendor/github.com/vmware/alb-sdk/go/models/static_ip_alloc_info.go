// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// StaticIPAllocInfo static Ip alloc info
// swagger:model StaticIpAllocInfo
type StaticIPAllocInfo struct {

	// IP address. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	IP *IPAddr `json:"ip"`

	// Object metadata. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjInfo *string `json:"obj_info,omitempty"`

	// Object which this IP address is allocated to. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjUUID *string `json:"obj_uuid,omitempty"`
}
