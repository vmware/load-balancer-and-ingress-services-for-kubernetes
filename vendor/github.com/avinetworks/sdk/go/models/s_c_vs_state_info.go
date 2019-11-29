package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SCVsStateInfo s c vs state info
// swagger:model SCVsStateInfo
type SCVsStateInfo struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.1.1.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.1.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.1.1.
	UUID *string `json:"uuid,omitempty"`

	//  Field introduced in 17.1.1.
	VipID *string `json:"vip_id,omitempty"`

	//  Field introduced in 17.1.1.
	VsID *string `json:"vs_id,omitempty"`
}
