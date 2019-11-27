package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsMgrDebugFilter metrics mgr debug filter
// swagger:model MetricsMgrDebugFilter
type MetricsMgrDebugFilter struct {

	// disable_hw_training of MetricsMgrDebugFilter.
	DisableHwTraining *string `json:"disable_hw_training,omitempty"`

	// entity of MetricsMgrDebugFilter.
	Entity *string `json:"entity,omitempty"`

	// setting to reduce the grace period for license expiry in hours.
	LicenseGracePeriod *string `json:"license_grace_period,omitempty"`

	// log_first_n of MetricsMgrDebugFilter.
	LogFirstN *string `json:"log_first_n,omitempty"`

	// logging_freq of MetricsMgrDebugFilter.
	LoggingFreq *string `json:"logging_freq,omitempty"`

	// metric_instance_id of MetricsMgrDebugFilter.
	MetricInstanceID *string `json:"metric_instance_id,omitempty"`

	// obj of MetricsMgrDebugFilter.
	Obj *string `json:"obj,omitempty"`

	// skip_cluster_map_check of MetricsMgrDebugFilter.
	SkipClusterMapCheck *string `json:"skip_cluster_map_check,omitempty"`

	// skip_metrics_db_writes of MetricsMgrDebugFilter.
	SkipMetricsDbWrites *string `json:"skip_metrics_db_writes,omitempty"`
}
