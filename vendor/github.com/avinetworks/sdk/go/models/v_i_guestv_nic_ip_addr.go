package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIGuestvNicIPAddr v i guestv nic IP addr
// swagger:model VIGuestvNicIPAddr
type VIGuestvNicIPAddr struct {

	// ip_addr of VIGuestvNicIPAddr.
	// Required: true
	IPAddr *string `json:"ip_addr"`

	// Number of mask.
	// Required: true
	Mask *int32 `json:"mask"`
}
