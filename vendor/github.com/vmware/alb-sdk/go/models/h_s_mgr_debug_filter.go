package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HSMgrDebugFilter h s mgr debug filter
// swagger:model HSMgrDebugFilter
type HSMgrDebugFilter struct {

	// entity of HSMgrDebugFilter.
	Entity *string `json:"entity,omitempty"`

	//  Enum options - VSERVER_METRICS_ENTITY, VM_METRICS_ENTITY, SE_METRICS_ENTITY, CONTROLLER_METRICS_ENTITY, APPLICATION_METRICS_ENTITY, TENANT_METRICS_ENTITY, POOL_METRICS_ENTITY.
	MetricEntity *string `json:"metric_entity,omitempty"`

	// Number of period.
	Period *int32 `json:"period,omitempty"`

	// pool of HSMgrDebugFilter.
	Pool *string `json:"pool,omitempty"`

	// server of HSMgrDebugFilter.
	Server *string `json:"server,omitempty"`

	// Placeholder for description of property skip_hs_db_writes of obj type HSMgrDebugFilter field type str  type boolean
	SkipHsDbWrites *bool `json:"skip_hs_db_writes,omitempty"`
}
