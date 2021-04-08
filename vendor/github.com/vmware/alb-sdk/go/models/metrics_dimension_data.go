package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsDimensionData metrics dimension data
// swagger:model MetricsDimensionData
type MetricsDimensionData struct {

	// Dimension Type. Enum options - METRICS_DIMENSION_METRIC_TIMESTAMP, METRICS_DIMENSION_COUNTRY, METRICS_DIMENSION_OS, METRICS_DIMENSION_URL, METRICS_DIMENSION_DEVTYPE, METRICS_DIMENSION_LANG, METRICS_DIMENSION_BROWSER, METRICS_DIMENSION_IPGROUP, METRICS_DIMENSION_ATTACK, METRICS_DIMENSION_ASN.
	// Required: true
	Dimension *string `json:"dimension"`

	// Dimension ID.
	// Required: true
	DimensionID *string `json:"dimension_id"`
}
