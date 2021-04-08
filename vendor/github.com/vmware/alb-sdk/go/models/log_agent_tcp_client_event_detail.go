package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LogAgentTCPClientEventDetail log agent TCP client event detail
// swagger:model LogAgentTCPClientEventDetail
type LogAgentTCPClientEventDetail struct {

	//  Field introduced in 20.1.3.
	ErrorCode *string `json:"error_code,omitempty"`

	//  Field introduced in 20.1.3.
	ErrorReason *string `json:"error_reason,omitempty"`

	//  Field introduced in 20.1.3.
	Host *string `json:"host,omitempty"`

	//  Field introduced in 20.1.3.
	Port *string `json:"port,omitempty"`
}
