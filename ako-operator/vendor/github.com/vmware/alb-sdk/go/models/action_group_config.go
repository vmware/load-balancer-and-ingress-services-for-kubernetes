// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ActionGroupConfig action group config
// swagger:model ActionGroupConfig
type ActionGroupConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// reference of the action script configuration to be used. It is a reference to an object of type AlertScriptConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActionScriptConfigRef *string `json:"action_script_config_ref,omitempty"`

	// Trigger Notification to AutoScale Manager. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AutoscaleTriggerNotification *bool `json:"autoscale_trigger_notification,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Select the Email Notification configuration to use when sending alerts via email. It is a reference to an object of type AlertEmailConfig. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EmailConfigRef *string `json:"email_config_ref,omitempty"`

	// Generate Alert only to external destinations. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	// Required: true
	ExternalOnly *bool `json:"external_only"`

	// When an alert is generated, mark its priority via the Alert Level. Enum options - ALERT_LOW, ALERT_MEDIUM, ALERT_HIGH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Level *string `json:"level"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Select the SNMP Trap Notification to use when sending alerts via SNMP Trap. It is a reference to an object of type SnmpTrapProfile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SnmpTrapProfileRef *string `json:"snmp_trap_profile_ref,omitempty"`

	// Select the Syslog Notification configuration to use when sending alerts via Syslog. It is a reference to an object of type AlertSyslogConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyslogConfigRef *string `json:"syslog_config_ref,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
