package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHmEventGslbPoolDetails se hm event gslb pool details
// swagger:model SeHmEventGslbPoolDetails
type SeHmEventGslbPoolDetails struct {

	// GslbService Pool name.
	Gsgroup *string `json:"gsgroup,omitempty"`

	// Gslb service name. It is a reference to an object of type GslbService.
	GslbService *string `json:"gslb_service,omitempty"`

	// GslbService member details.
	Gsmember *SeHmEventGslbPoolMemberDetails `json:"gsmember,omitempty"`

	// HA Compromised reason.
	HaReason *string `json:"ha_reason,omitempty"`

	// Service Engine.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the event generator.
	SrcUUID *string `json:"src_uuid,omitempty"`
}
