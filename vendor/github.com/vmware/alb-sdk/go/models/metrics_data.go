package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsData metrics data
// swagger:model MetricsData
type MetricsData struct {

	// Placeholder for description of property application_response_time of obj type MetricsData field type str  type number
	ApplicationResponseTime *float64 `json:"application_response_time,omitempty"`

	// Placeholder for description of property blocking_time of obj type MetricsData field type str  type number
	BlockingTime *float64 `json:"blocking_time,omitempty"`

	// Placeholder for description of property browser_rendering_time of obj type MetricsData field type str  type number
	BrowserRenderingTime *float64 `json:"browser_rendering_time,omitempty"`

	// Placeholder for description of property client_rtt of obj type MetricsData field type str  type number
	ClientRtt *float64 `json:"client_rtt,omitempty"`

	// Placeholder for description of property connection_time of obj type MetricsData field type str  type number
	ConnectionTime *float64 `json:"connection_time,omitempty"`

	// Placeholder for description of property dns_lookup_time of obj type MetricsData field type str  type number
	DNSLookupTime *float64 `json:"dns_lookup_time,omitempty"`

	// Placeholder for description of property dom_content_load_time of obj type MetricsData field type str  type number
	DomContentLoadTime *float64 `json:"dom_content_load_time,omitempty"`

	// Placeholder for description of property is_null of obj type MetricsData field type str  type boolean
	IsNull *bool `json:"is_null,omitempty"`

	// Number of num_samples.
	NumSamples *int32 `json:"num_samples,omitempty"`

	// Placeholder for description of property page_download_time of obj type MetricsData field type str  type number
	PageDownloadTime *float64 `json:"page_download_time,omitempty"`

	// Placeholder for description of property page_load_time of obj type MetricsData field type str  type number
	PageLoadTime *float64 `json:"page_load_time,omitempty"`

	// Placeholder for description of property prediction_interval_high of obj type MetricsData field type str  type number
	PredictionIntervalHigh *float64 `json:"prediction_interval_high,omitempty"`

	// Placeholder for description of property prediction_interval_low of obj type MetricsData field type str  type number
	PredictionIntervalLow *float64 `json:"prediction_interval_low,omitempty"`

	// Placeholder for description of property redirection_time of obj type MetricsData field type str  type number
	RedirectionTime *float64 `json:"redirection_time,omitempty"`

	// Placeholder for description of property rum_client_data_transfer_time of obj type MetricsData field type str  type number
	RumClientDataTransferTime *float64 `json:"rum_client_data_transfer_time,omitempty"`

	// Placeholder for description of property server_rtt of obj type MetricsData field type str  type number
	ServerRtt *float64 `json:"server_rtt,omitempty"`

	// Placeholder for description of property service_time of obj type MetricsData field type str  type number
	ServiceTime *float64 `json:"service_time,omitempty"`

	// timestamp of MetricsData.
	Timestamp *string `json:"timestamp,omitempty"`

	// Placeholder for description of property value of obj type MetricsData field type str  type number
	// Required: true
	Value *float64 `json:"value"`

	//  Field introduced in 17.2.2.
	ValueStr *string `json:"value_str,omitempty"`

	//  Field introduced in 17.2.2.
	ValueStrDesc *string `json:"value_str_desc,omitempty"`

	// Placeholder for description of property waiting_time of obj type MetricsData field type str  type number
	WaitingTime *float64 `json:"waiting_time,omitempty"`
}
