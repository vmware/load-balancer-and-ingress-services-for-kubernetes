package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SCServerStateInfo s c server state info
// swagger:model SCServerStateInfo
type SCServerStateInfo struct {

	//  Field introduced in 17.1.1.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Field introduced in 17.1.1.
	ServerIP *IPAddr `json:"server_ip,omitempty"`

	//  Allowed values are 1-65535. Field introduced in 17.1.1.
	ServerPort *int32 `json:"server_port,omitempty"`
}
