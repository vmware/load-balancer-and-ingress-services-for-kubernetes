package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ActionGroupConfig action group config
// swagger:model ActionGroupConfig
type ActionGroupConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// reference of the action script configuration to be used. It is a reference to an object of type AlertScriptConfig.
	ActionScriptConfigRef *string `json:"action_script_config_ref,omitempty"`

	// Trigger Notification to AutoScale Manager.
	AutoscaleTriggerNotification *bool `json:"autoscale_trigger_notification,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// Select the Email Notification configuration to use when sending alerts via email. It is a reference to an object of type AlertEmailConfig.
	EmailConfigRef *string `json:"email_config_ref,omitempty"`

	// Generate Alert only to external destinations.
	// Required: true
	ExternalOnly *bool `json:"external_only"`

	// When an alert is generated, mark its priority via the Alert Level. Enum options - ALERT_LOW, ALERT_MEDIUM, ALERT_HIGH.
	// Required: true
	Level *string `json:"level"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Select the SNMP Trap Notification to use when sending alerts via SNMP Trap. It is a reference to an object of type SnmpTrapProfile.
	SnmpTrapProfileRef *string `json:"snmp_trap_profile_ref,omitempty"`

	// Select the Syslog Notification configuration to use when sending alerts via Syslog. It is a reference to an object of type AlertSyslogConfig.
	SyslogConfigRef *string `json:"syslog_config_ref,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
