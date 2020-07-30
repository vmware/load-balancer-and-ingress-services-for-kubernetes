package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsError vs error
// swagger:model VsError
type VsError struct {

	// The time at which the error occurred. Field introduced in 18.2.10, 20.1.1.
	EventTimestamp *TimeStamp `json:"event_timestamp,omitempty"`

	// reason of VsError.
	Reason []string `json:"reason,omitempty"`

	//  Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY.
	SeGroupHaMode *string `json:"se_group_ha_mode,omitempty"`

	//  It is a reference to an object of type ServiceEngineGroup.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	// The SE on which the VS errored during scale-in/scale-out operations. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.10, 20.1.1.
	SeRef *string `json:"se_ref,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Enum options - TRAFFIC_DISRUPTED, TRAFFIC_NOT_DISRUPTED.
	TrafficStatus *string `json:"traffic_status,omitempty"`

	// vip_id of VsError.
	VipID *string `json:"vip_id,omitempty"`

	//  It is a reference to an object of type VirtualService.
	VsRef *string `json:"vs_ref,omitempty"`
}
