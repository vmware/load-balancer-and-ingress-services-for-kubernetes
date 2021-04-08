package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AnomalyEventDetails anomaly event details
// swagger:model AnomalyEventDetails
type AnomalyEventDetails struct {

	// Placeholder for description of property deviation of obj type AnomalyEventDetails field type str  type number
	Deviation *float64 `json:"deviation,omitempty"`

	// metric_id of AnomalyEventDetails.
	// Required: true
	MetricID *string `json:"metric_id"`

	// metric_timestamp of AnomalyEventDetails.
	// Required: true
	MetricTimestamp *string `json:"metric_timestamp"`

	// Deprecated. Enum options - EXPONENTIAL_MOVING_AVG, EXPONENTIAL_WEIGHTED_MOVING_AVG, HOLTWINTERS_AT_AS, HOLTWINTERS_AT_MS.
	Model *string `json:"model,omitempty"`

	//  Enum options - EXPONENTIAL_MOVING_AVG, EXPONENTIAL_WEIGHTED_MOVING_AVG, HOLTWINTERS_AT_AS, HOLTWINTERS_AT_MS.
	Models []string `json:"models,omitempty"`

	// node_id of AnomalyEventDetails.
	NodeID *string `json:"node_id,omitempty"`

	//  Enum options - METRICS_OBJ_TYPE_UNKNOWN. VSERVER_L4_SERVER. VSERVER_L4_CLIENT. VSERVER_L7_SERVER. VSERVER_L7_CLIENT. VM_METRICS_OBJ. SE_METRICS_OBJ. VSERVER_RUM. CONTROLLER_METRICS_OBJ. METRICS_COLLECTION. METRICS_RUM_PREAGG_BROWSER_OBJ. METRICS_RUM_PREAGG_COUNTRY_OBJ. METRICS_RUM_PREAGG_DEVTYPE_OBJ. METRICS_RUM_PREAGG_LANG_OBJ. METRICS_RUM_PREAGG_OS_OBJ. METRICS_RUM_PREAGG_URL_OBJ. METRICS_ANOMALY_OBJ. METRICS_HEALTHSCORE_OBJ. METRICS_RESOURCE_TIMING_BROWSER_OBJ. METRICS_RESOURCE_TIMING_OS_OBJ...
	ObjType *string `json:"obj_type,omitempty"`

	// pool_name of AnomalyEventDetails.
	PoolName *string `json:"pool_name,omitempty"`

	// Unique object identifier of pool.
	PoolUUID *string `json:"pool_uuid,omitempty"`

	//  Enum options - ANZ_PRIORITY_HIGH, ANZ_PRIORITY_MEDIUM, ANZ_PRIORITY_LOW.
	// Required: true
	Priority *string `json:"priority"`

	// server of AnomalyEventDetails.
	Server *string `json:"server,omitempty"`

	// Placeholder for description of property value of obj type AnomalyEventDetails field type str  type number
	// Required: true
	Value *float64 `json:"value"`
}
