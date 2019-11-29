package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertRule alert rule
// swagger:model AlertRule
type AlertRule struct {

	// Placeholder for description of property conn_app_log_rule of obj type AlertRule field type str  type object
	ConnAppLogRule *AlertFilter `json:"conn_app_log_rule,omitempty"`

	// event_match_filter of AlertRule.
	EventMatchFilter *string `json:"event_match_filter,omitempty"`

	// Placeholder for description of property metrics_rule of obj type AlertRule field type str  type object
	MetricsRule []*AlertRuleMetric `json:"metrics_rule,omitempty"`

	//  Enum options - OPERATOR_AND, OPERATOR_OR.
	Operator *string `json:"operator,omitempty"`

	// Placeholder for description of property sys_event_rule of obj type AlertRule field type str  type object
	SysEventRule []*AlertRuleEvent `json:"sys_event_rule,omitempty"`
}
