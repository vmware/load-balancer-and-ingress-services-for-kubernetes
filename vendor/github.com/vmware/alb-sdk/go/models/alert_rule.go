// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertRule alert rule
// swagger:model AlertRule
type AlertRule struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnAppLogRule *AlertFilter `json:"conn_app_log_rule,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventMatchFilter *string `json:"event_match_filter,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsRule []*AlertRuleMetric `json:"metrics_rule,omitempty"`

	//  Enum options - OPERATOR_AND, OPERATOR_OR. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Operator *string `json:"operator,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SysEventRule []*AlertRuleEvent `json:"sys_event_rule,omitempty"`
}
