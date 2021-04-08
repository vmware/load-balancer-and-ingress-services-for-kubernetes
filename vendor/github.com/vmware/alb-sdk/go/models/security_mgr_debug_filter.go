package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecurityMgrDebugFilter security mgr debug filter
// swagger:model SecurityMgrDebugFilter
type SecurityMgrDebugFilter struct {

	// Dynamically adapt configuration parameters for Application Learning feature. Field introduced in 20.1.1.
	EnableAdaptiveConfig *bool `json:"enable_adaptive_config,omitempty"`

	// uuid of the entity. It is a reference to an object of type Virtualservice. Field introduced in 18.2.6.
	EntityRef *string `json:"entity_ref,omitempty"`
}
