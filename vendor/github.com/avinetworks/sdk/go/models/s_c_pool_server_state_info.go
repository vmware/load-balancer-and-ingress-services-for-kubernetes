package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SCPoolServerStateInfo s c pool server state info
// swagger:model SCPoolServerStateInfo
type SCPoolServerStateInfo struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.1.1.
	IsServer *bool `json:"is_server,omitempty"`

	//  Field introduced in 17.1.1.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  Field introduced in 17.1.1.
	PoolID *string `json:"pool_id,omitempty"`

	//  Field introduced in 17.1.1.
	ServerStates []*SCServerStateInfo `json:"server_states,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`
}
