package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsDataHeader metrics data header
// swagger:model MetricsDataHeader
type MetricsDataHeader struct {

	// Metrics derivation info.
	DerivationData *MetricsDerivationData `json:"derivation_data,omitempty"`

	// Placeholder for description of property dimension_data of obj type MetricsDataHeader field type str  type object
	DimensionData []*MetricsDimensionData `json:"dimension_data,omitempty"`

	// entity uuid.
	EntityUUID *string `json:"entity_uuid,omitempty"`

	// metric_description of MetricsDataHeader.
	MetricDescription *string `json:"metric_description,omitempty"`

	// Placeholder for description of property metrics_min_scale of obj type MetricsDataHeader field type str  type number
	MetricsMinScale *float64 `json:"metrics_min_scale,omitempty"`

	// Placeholder for description of property metrics_sum_agg_invalid of obj type MetricsDataHeader field type str  type boolean
	MetricsSumAggInvalid *bool `json:"metrics_sum_agg_invalid,omitempty"`

	// Missing data intervals. data in these intervals are not used for stats calculation.
	MissingIntervals []*MetricsMissingDataInterval `json:"missing_intervals,omitempty"`

	// name of the column.
	// Required: true
	Name *string `json:"name"`

	// object ID of the series when object ID was specified in the metric.
	ObjID *string `json:"obj_id,omitempty"`

	// obj_id_type. Enum options - METRICS_OBJ_ID_TYPE_VIRTUALSERVICE, METRICS_OBJ_ID_TYPE_SERVER, METRICS_OBJ_ID_TYPE_POOL, METRICS_OBJ_ID_TYPE_SERVICEENGINE, METRICS_OBJ_ID_TYPE_VIRTUALMACHINE, METRICS_OBJ_ID_TYPE_CONTROLLER, METRICS_OBJ_ID_TYPE_TENANT, METRICS_OBJ_ID_TYPE_CLUSTER, METRICS_OBJ_ID_TYPE_SE_INTERFACE.
	ObjIDType *string `json:"obj_id_type,omitempty"`

	// pool_id for the metric.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	// Placeholder for description of property priority of obj type MetricsDataHeader field type str  type boolean
	Priority *bool `json:"priority,omitempty"`

	// server ip port.
	Server *string `json:"server,omitempty"`

	// Service Engine ref or UUID. Field introduced in 17.2.8.
	ServiceengineUUID *string `json:"serviceengine_uuid,omitempty"`

	// statistics of the metric.
	Statistics *MetricStatistics `json:"statistics,omitempty"`

	// Tenant ref or UUID.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// units of the column data. Enum options - METRIC_COUNT, BITS_PER_SECOND, MILLISECONDS, SECONDS, PER_SECOND, BYTES, PERCENT, KILO_BYTES, KILO_BYTES_PER_SECOND, BYTES_PER_SECOND, KILO_BITS_PER_SECOND, GIGA_BYTES, MEGA_BYTES, NORMALIZED, STRING, SEC, MIN, DAYS, KB, MB, GB, MBPS, GHZ, RATIO, WORD, MICROSECONDS, HEALTH.
	Units *string `json:"units,omitempty"`
}
