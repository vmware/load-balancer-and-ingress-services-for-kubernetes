// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsMgrDebugFilter metrics mgr debug filter
// swagger:model MetricsMgrDebugFilter
type MetricsMgrDebugFilter struct {

	// Set to ignore skip_eval_period field in metrics_anomaly_option. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DebugSkipEvalPeriod *string `json:"debug_skip_eval_period,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableHwTraining *string `json:"disable_hw_training,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Entity *string `json:"entity,omitempty"`

	// Setting to reduce the grace period for license expiry in hours. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseGracePeriod *string `json:"license_grace_period,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogFirstN *string `json:"log_first_n,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LoggingFreq *string `json:"logging_freq,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricInstanceID *string `json:"metric_instance_id,omitempty"`

	// Setting to change the number of queries being processed by per DB connection by Metrics Manager. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinDbQueriesEachConn *string `json:"min_db_queries_each_conn,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Obj *string `json:"obj,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SkipClusterMapCheck *string `json:"skip_cluster_map_check,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SkipMetricsDbWrites *string `json:"skip_metrics_db_writes,omitempty"`
}
