package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHmEventVsDetails se hm event vs details
// swagger:model SeHmEventVsDetails
type SeHmEventVsDetails struct {

	// HA Compromised reason.
	HaReason *string `json:"ha_reason,omitempty"`

	// Reason for Virtual Service Down.
	Reason *string `json:"reason,omitempty"`

	// Service Engine name.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the event generator.
	SrcUUID *string `json:"src_uuid,omitempty"`

	// VIP address.
	Vip6Address *IPAddr `json:"vip6_address,omitempty"`

	// VIP address.
	VipAddress *IPAddr `json:"vip_address,omitempty"`

	// VIP id.
	VipID *string `json:"vip_id,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService.
	VirtualService *string `json:"virtual_service,omitempty"`
}
