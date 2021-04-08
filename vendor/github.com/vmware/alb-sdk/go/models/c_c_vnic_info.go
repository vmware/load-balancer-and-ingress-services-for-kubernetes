package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CCVnicInfo c c vnic info
// swagger:model CC_VnicInfo
type CCVnicInfo struct {

	// mac_address of CC_VnicInfo.
	MacAddress *string `json:"mac_address,omitempty"`

	// Unique object identifier of network.
	// Required: true
	NetworkUUID *string `json:"network_uuid"`

	// Unique object identifier of port.
	PortUUID *string `json:"port_uuid,omitempty"`

	//  Enum options - SYSERR_SUCCESS. SYSERR_FAILURE. SYSERR_OUT_OF_MEMORY. SYSERR_NO_ENT. SYSERR_INVAL. SYSERR_ACCESS. SYSERR_FAULT. SYSERR_IO. SYSERR_TIMEOUT. SYSERR_NOT_SUPPORTED. SYSERR_NOT_READY. SYSERR_UPGRADE_IN_PROGRESS. SYSERR_WARM_START_IN_PROGRESS. SYSERR_TRY_AGAIN. SYSERR_NOT_UPGRADING. SYSERR_PENDING. SYSERR_EVENT_GEN_FAILURE. SYSERR_CONFIG_PARAM_MISSING. SYSERR_RANGE. SYSERR_BAD_REQUEST...
	Status *string `json:"status,omitempty"`

	// status_string of CC_VnicInfo.
	StatusString *string `json:"status_string,omitempty"`

	// Unique object identifier of subnet.
	SubnetUUID *string `json:"subnet_uuid,omitempty"`

	// Unique object identifier of vrf.
	VrfUUID *string `json:"vrf_uuid,omitempty"`
}
