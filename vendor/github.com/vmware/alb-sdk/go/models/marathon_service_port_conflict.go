package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MarathonServicePortConflict marathon service port conflict
// swagger:model MarathonServicePortConflict
type MarathonServicePortConflict struct {

	// app_name of MarathonServicePortConflict.
	AppName *string `json:"app_name,omitempty"`

	// cc_id of MarathonServicePortConflict.
	CcID *string `json:"cc_id,omitempty"`

	// marathon_url of MarathonServicePortConflict.
	// Required: true
	MarathonURL *string `json:"marathon_url"`

	// Number of port.
	// Required: true
	Port *int32 `json:"port"`
}
