package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ObjSyncConfig obj sync config
// swagger:model ObjSyncConfig
type ObjSyncConfig struct {

	// SE CPU limit for InterSE Object Distribution. Allowed values are 15-80. Field introduced in 20.1.3. Unit is PERCENT.
	ObjsyncCPULimit *int32 `json:"objsync_cpu_limit,omitempty"`

	// Hub election interval for InterSE Object Distribution. Allowed values are 30-300. Field introduced in 20.1.3. Unit is SEC.
	ObjsyncHubElectInterval *int32 `json:"objsync_hub_elect_interval,omitempty"`

	// Reconcile interval for InterSE Object Distribution. Allowed values are 1-120. Field introduced in 20.1.3. Unit is SEC.
	ObjsyncReconcileInterval *int32 `json:"objsync_reconcile_interval,omitempty"`
}
