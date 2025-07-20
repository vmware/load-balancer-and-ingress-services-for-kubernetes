// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDataHeader metrics data header
// swagger:model MetricsDataHeader
type MetricsDataHeader struct {

	// Metrics derivation info. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DerivationData *MetricsDerivationData `json:"derivation_data,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DimensionData []*MetricsDimensionData `json:"dimension_data,omitempty"`

	// entity uuid. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EntityUUID *string `json:"entity_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricDescription *string `json:"metric_description,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsMinScale *float64 `json:"metrics_min_scale,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsSumAggInvalid *bool `json:"metrics_sum_agg_invalid,omitempty"`

	// Missing data intervals. data in these intervals are not used for stats calculation. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MissingIntervals []*MetricsMissingDataInterval `json:"missing_intervals,omitempty"`

	// name of the column. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// object ID of the series when object ID was specified in the metric. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjID *string `json:"obj_id,omitempty"`

	// obj_id_type. Enum options - METRICS_OBJ_ID_TYPE_VIRTUALSERVICE, METRICS_OBJ_ID_TYPE_SERVER, METRICS_OBJ_ID_TYPE_POOL, METRICS_OBJ_ID_TYPE_SERVICEENGINE, METRICS_OBJ_ID_TYPE_VIRTUALMACHINE, METRICS_OBJ_ID_TYPE_CONTROLLER, METRICS_OBJ_ID_TYPE_TENANT, METRICS_OBJ_ID_TYPE_CLUSTER, METRICS_OBJ_ID_TYPE_SE_INTERFACE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjIDType *string `json:"obj_id_type,omitempty"`

	// pool_id for the metric. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Priority *bool `json:"priority,omitempty"`

	// server ip port. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Server *string `json:"server,omitempty"`

	// Service Engine ref or UUID. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceengineUUID *string `json:"serviceengine_uuid,omitempty"`

	// statistics of the metric. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Statistics *MetricStatistics `json:"statistics,omitempty"`

	// Tenant ref or UUID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// units of the column data. Enum options - METRIC_COUNT, BITS_PER_SECOND, MILLISECONDS, SECONDS, PER_SECOND, BYTES, PERCENT, KILO_BYTES, KILO_BYTES_PER_SECOND, BYTES_PER_SECOND, KILO_BITS_PER_SECOND, GIGA_BYTES, MEGA_BYTES, NORMALIZED, STRING, SEC, MIN, DAYS, KB, MB.... Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Units *string `json:"units,omitempty"`
}
