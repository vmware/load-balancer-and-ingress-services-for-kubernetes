package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafLog waf log
// swagger:model WafLog
type WafLog struct {

	// Latency (in microseconds) in WAF Request Body Phase. Field introduced in 17.2.2.
	LatencyRequestBodyPhase *int64 `json:"latency_request_body_phase,omitempty"`

	// Latency (in microseconds) in WAF Request Header Phase. Field introduced in 17.2.2.
	LatencyRequestHeaderPhase *int64 `json:"latency_request_header_phase,omitempty"`

	// Latency (in microseconds) in WAF Response Body Phase. Field introduced in 17.2.2.
	LatencyResponseBodyPhase *int64 `json:"latency_response_body_phase,omitempty"`

	// Latency (in microseconds) in WAF Response Header Phase. Field introduced in 17.2.2.
	LatencyResponseHeaderPhase *int64 `json:"latency_response_header_phase,omitempty"`

	//  Field introduced in 17.2.1.
	RuleLogs []*WafRuleLog `json:"rule_logs,omitempty"`

	// Denotes whether WAF is running in detection mode or enforcement mode, whether any rules matched the transaction, and whether transaction is dropped by the WAF module. Enum options - NO_WAF, FLAGGED, PASSED, REJECTED. Field introduced in 17.2.2.
	Status *string `json:"status,omitempty"`
}
