package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsQueryResponse metrics query response
// swagger:model MetricsQueryResponse
type MetricsQueryResponse struct {

	// Unique object identifier of entity.
	EntityUUID *string `json:"entity_uuid,omitempty"`

	// returns the ID specified in the query.
	ID *string `json:"id,omitempty"`

	// Number of limit.
	Limit *int32 `json:"limit,omitempty"`

	//  Enum options - VSERVER_METRICS_ENTITY, VM_METRICS_ENTITY, SE_METRICS_ENTITY, CONTROLLER_METRICS_ENTITY, APPLICATION_METRICS_ENTITY, TENANT_METRICS_ENTITY, POOL_METRICS_ENTITY.
	MetricEntity *string `json:"metric_entity,omitempty"`

	// metric_id of MetricsQueryResponse.
	MetricID *string `json:"metric_id,omitempty"`

	// Placeholder for description of property series of obj type MetricsQueryResponse field type str  type object
	Series []*MetricsDataSeries `json:"series,omitempty"`

	// start of MetricsQueryResponse.
	Start *string `json:"start,omitempty"`

	// Number of step.
	Step *int32 `json:"step,omitempty"`

	// stop of MetricsQueryResponse.
	Stop *string `json:"stop,omitempty"`
}
