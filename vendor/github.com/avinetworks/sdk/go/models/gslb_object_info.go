package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbObjectInfo gslb object info
// swagger:model GslbObjectInfo
type GslbObjectInfo struct {

	// Indicates the object uuid. Field introduced in 18.1.5, 18.2.1.
	Obj *GslbObj `json:"obj,omitempty"`

	// Indicates the object uuid. Field introduced in 17.1.1.
	ObjectUUID *string `json:"object_uuid,omitempty"`

	// Indicates the object type Gslb, GslbService or GslbGeoDbProfile. . Field introduced in 17.1.1.
	PbName *string `json:"pb_name,omitempty"`

	// Indicates the state of the object unchanged or changed. This is used in vs-mgr to push just the uuid or uuid + protobuf to the SE-Agent. Enum options - GSLB_OBJECT_CHANGED, GSLB_OBJECT_UNCHANGED, GSLB_OBJECT_DELETE. Field introduced in 17.1.1.
	State *string `json:"state,omitempty"`
}
