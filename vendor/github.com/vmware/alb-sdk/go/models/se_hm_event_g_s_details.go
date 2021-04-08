package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHmEventGSDetails se hm event g s details
// swagger:model SeHmEventGSDetails
type SeHmEventGSDetails struct {

	// GslbService name. It is a reference to an object of type GslbService.
	GslbService *string `json:"gslb_service,omitempty"`

	// HA Compromised reason.
	HaReason *string `json:"ha_reason,omitempty"`

	// Reason Gslb Service is down.
	Reason *string `json:"reason,omitempty"`

	// Service Engine name.
	SeName *string `json:"se_name,omitempty"`

	// UUID of the event generator.
	SrcUUID *string `json:"src_uuid,omitempty"`
}
