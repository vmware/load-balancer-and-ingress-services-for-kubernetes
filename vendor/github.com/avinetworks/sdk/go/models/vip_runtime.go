package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VipRuntime vip runtime
// swagger:model VipRuntime
type VipRuntime struct {

	// ev of VipRuntime.
	Ev []string `json:"ev,omitempty"`

	// Placeholder for description of property ev_status of obj type VipRuntime field type str  type object
	EvStatus *VsEvStatus `json:"ev_status,omitempty"`

	// Placeholder for description of property first_se_assigned_time of obj type VipRuntime field type str  type object
	FirstSeAssignedTime *TimeStamp `json:"first_se_assigned_time,omitempty"`

	// Placeholder for description of property first_time_placement of obj type VipRuntime field type str  type boolean
	FirstTimePlacement *bool `json:"first_time_placement,omitempty"`

	// fsm_state_id of VipRuntime.
	FsmStateID *string `json:"fsm_state_id,omitempty"`

	// fsm_state_name of VipRuntime.
	FsmStateName *string `json:"fsm_state_name,omitempty"`

	// Placeholder for description of property last_changed_time of obj type VipRuntime field type str  type object
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Placeholder for description of property last_scale_status of obj type VipRuntime field type str  type object
	LastScaleStatus *ScaleStatus `json:"last_scale_status,omitempty"`

	// Placeholder for description of property marked_for_delete of obj type VipRuntime field type str  type boolean
	MarkedForDelete *bool `json:"marked_for_delete,omitempty"`

	//  Enum options - METRICS_MGR_PORT_0, METRICS_MGR_PORT_1, METRICS_MGR_PORT_2, METRICS_MGR_PORT_3.
	MetricsMgrPort *string `json:"metrics_mgr_port,omitempty"`

	// Placeholder for description of property migrate_in_progress of obj type VipRuntime field type str  type boolean
	MigrateInProgress *bool `json:"migrate_in_progress,omitempty"`

	// Placeholder for description of property migrate_request of obj type VipRuntime field type str  type object
	MigrateRequest *VsMigrateParams `json:"migrate_request,omitempty"`

	// Placeholder for description of property migrate_scalein_pending of obj type VipRuntime field type str  type boolean
	MigrateScaleinPending *bool `json:"migrate_scalein_pending,omitempty"`

	// Placeholder for description of property migrate_scaleout_pending of obj type VipRuntime field type str  type boolean
	MigrateScaleoutPending *bool `json:"migrate_scaleout_pending,omitempty"`

	// Number of num_additional_se.
	NumAdditionalSe *int32 `json:"num_additional_se,omitempty"`

	//  Enum options - METRICS_MGR_PORT_0, METRICS_MGR_PORT_1, METRICS_MGR_PORT_2, METRICS_MGR_PORT_3.
	PrevMetricsMgrPort *string `json:"prev_metrics_mgr_port,omitempty"`

	// Number of progress_percent.
	ProgressPercent *int32 `json:"progress_percent,omitempty"`

	// Placeholder for description of property requested_resource of obj type VipRuntime field type str  type object
	RequestedResource *VirtualServiceResource `json:"requested_resource,omitempty"`

	// Placeholder for description of property scale_status of obj type VipRuntime field type str  type object
	ScaleStatus *ScaleStatus `json:"scale_status,omitempty"`

	// Placeholder for description of property scalein_in_progress of obj type VipRuntime field type str  type boolean
	ScaleinInProgress *bool `json:"scalein_in_progress,omitempty"`

	// Placeholder for description of property scalein_request of obj type VipRuntime field type str  type object
	ScaleinRequest *VsScaleinParams `json:"scalein_request,omitempty"`

	// Placeholder for description of property scaleout_in_progress of obj type VipRuntime field type str  type boolean
	ScaleoutInProgress *bool `json:"scaleout_in_progress,omitempty"`

	// Placeholder for description of property se_list of obj type VipRuntime field type str  type object
	SeList []*SeList `json:"se_list,omitempty"`

	// Placeholder for description of property servers_configured of obj type VipRuntime field type str  type boolean
	ServersConfigured *bool `json:"servers_configured,omitempty"`

	// Placeholder for description of property supp_runtime_status of obj type VipRuntime field type str  type object
	SuppRuntimeStatus *OperationalStatus `json:"supp_runtime_status,omitempty"`

	// Placeholder for description of property user_scaleout_pending of obj type VipRuntime field type str  type boolean
	UserScaleoutPending *bool `json:"user_scaleout_pending,omitempty"`

	// vip_id of VipRuntime.
	VipID *string `json:"vip_id,omitempty"`

	// VIP finished resyncing with resource manager. Field introduced in 18.1.4, 18.2.1.
	WarmstartResyncDone *bool `json:"warmstart_resync_done,omitempty"`

	// RPC sent to resource manager for warmstart resync. Field introduced in 18.1.4, 18.2.1.
	WarmstartResyncSent *bool `json:"warmstart_resync_sent,omitempty"`
}
