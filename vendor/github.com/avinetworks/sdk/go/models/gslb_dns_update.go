package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbDNSUpdate gslb Dns update
// swagger:model GslbDnsUpdate
type GslbDNSUpdate struct {

	// Number of clear_on_max_retries.
	ClearOnMaxRetries *int32 `json:"clear_on_max_retries,omitempty"`

	// Gslb, GslbService objects that is pushed on a per Dns basis. Field introduced in 17.1.1.
	ObjInfo []*GslbObjectInfo `json:"obj_info,omitempty"`

	// Number of send_interval.
	SendInterval *int32 `json:"send_interval,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
