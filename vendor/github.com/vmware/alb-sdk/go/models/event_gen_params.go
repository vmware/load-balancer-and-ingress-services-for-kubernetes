package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// EventGenParams event gen params
// swagger:model EventGenParams
type EventGenParams struct {

	// cluster of EventGenParams.
	Cluster *string `json:"cluster,omitempty"`

	//  Enum options - VINFRA_DISC_DC. VINFRA_DISC_HOST. VINFRA_DISC_CLUSTER. VINFRA_DISC_VM. VINFRA_DISC_NW. MGMT_NW_NAME_CHANGED. DISCOVERY_DATACENTER_DEL. VM_ADDED. VM_REMOVED. VINFRA_DISC_COMPLETE. VCENTER_ADDRESS_ERROR. SE_GROUP_CLUSTER_DEL. SE_GROUP_MGMT_NW_DEL. MGMT_NW_DEL. VCENTER_BAD_CREDENTIALS. ESX_HOST_UNREACHABLE. SERVER_DELETED. SE_GROUP_HOST_DEL. VINFRA_DISC_FAILURE. ESX_HOST_POWERED_DOWN...
	Events []string `json:"events,omitempty"`

	// pool of EventGenParams.
	Pool *string `json:"pool,omitempty"`

	// sslkeyandcertificate of EventGenParams.
	Sslkeyandcertificate *string `json:"sslkeyandcertificate,omitempty"`

	// virtualservice of EventGenParams.
	Virtualservice *string `json:"virtualservice,omitempty"`
}
