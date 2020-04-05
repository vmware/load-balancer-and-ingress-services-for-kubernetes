package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPSMLocation waf p s m location
// swagger:model WafPSMLocation
type WafPSMLocation struct {

	// Free-text comment about this location. Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Location index, this is used to determine the order of the locations. Field introduced in 18.2.3.
	// Required: true
	Index *int32 `json:"index"`

	// Apply these rules only if the request is matching this description. Field introduced in 18.2.3.
	Match *WafPSMLocationMatch `json:"match,omitempty"`

	// User defined name for this location, it must be unique in the group. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`

	// A list of rules which should be applied on this location. Field introduced in 18.2.3.
	Rules []*WafPSMRule `json:"rules,omitempty"`
}
