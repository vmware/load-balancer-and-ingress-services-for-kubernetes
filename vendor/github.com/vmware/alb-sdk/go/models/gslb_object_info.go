// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbObjectInfo gslb object info
// swagger:model GslbObjectInfo
type GslbObjectInfo struct {

	// Indicates the object uuid. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Obj *GslbObj `json:"obj,omitempty"`

	// Indicates the object uuid. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjectUUID *string `json:"object_uuid,omitempty"`

	// Indicates the object type Gslb, GslbService or GslbGeoDbProfile. . Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PbName *string `json:"pb_name,omitempty"`

	// Indicates the state of the object unchanged or changed. This is used in vs-mgr to push just the uuid or uuid + protobuf to the SE-Agent. Enum options - GSLB_OBJECT_CHANGED, GSLB_OBJECT_UNCHANGED, GSLB_OBJECT_DELETE. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`
}
