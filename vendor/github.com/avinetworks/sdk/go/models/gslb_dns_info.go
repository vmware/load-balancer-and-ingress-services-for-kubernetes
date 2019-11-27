package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbDNSInfo gslb Dns info
// swagger:model GslbDnsInfo
type GslbDNSInfo struct {

	// This field indicates that atleast one DNS is active at the site.
	DNSActive *bool `json:"dns_active,omitempty"`

	// Placeholder for description of property dns_vs_states of obj type GslbDnsInfo field type str  type object
	DNSVsStates []*GslbPerDNSState `json:"dns_vs_states,omitempty"`

	// This field encapsulates the Gs-status edge-triggered framework. . Field introduced in 17.1.1.
	GsStatus *GslbDNSGsStatus `json:"gs_status,omitempty"`

	// This field is used to track the retry attempts for SE download errors. . Field introduced in 17.1.1.
	RetryCount *int32 `json:"retry_count,omitempty"`

	// This tables holds all the se-related info across all DNS-VS(es). . Field introduced in 17.1.1.
	SeTable []*GslbDNSSeInfo `json:"se_table,omitempty"`
}
