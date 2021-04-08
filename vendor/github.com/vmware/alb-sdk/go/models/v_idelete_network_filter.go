package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIdeleteNetworkFilter v idelete network filter
// swagger:model VIDeleteNetworkFilter
type VIdeleteNetworkFilter struct {

	//  Field introduced in 17.1.3.
	CloudUUID *string `json:"cloud_uuid,omitempty"`

	//  Field introduced in 17.1.3.
	Datacenter *string `json:"datacenter,omitempty"`

	//  Field introduced in 17.1.3.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	//  Field introduced in 17.1.3.
	VcenterURL *string `json:"vcenter_url,omitempty"`
}
