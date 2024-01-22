// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertConfig alert config
// swagger:model AlertConfig
type AlertConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// The alert config will trigger the selected alert action, which can send notifications and execute a controlscript. It is a reference to an object of type ActionGroupConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActionGroupRef *string `json:"action_group_ref,omitempty"`

	// list of filters matching on events or client logs used for triggering alerts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AlertRule *AlertRule `json:"alert_rule"`

	// This alert config applies to auto scale alerts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoscaleAlert *bool `json:"autoscale_alert,omitempty"`

	// Determines whether an alert is raised immediately when event occurs (realtime) or after specified number of events occurs within rolling time window. Enum options - REALTIME, ROLLINGWINDOW, WATERMARK. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Category *string `json:"category"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// A custom description field. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable or disable this alert config from generating new alerts. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// An alert is expired and deleted after the expiry time has elapsed.  The original event triggering the alert remains in the event's log. Allowed values are 1-31536000. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExpiryTime *uint32 `json:"expiry_time,omitempty"`

	// Name of the alert configuration. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// UUID of the resource for which alert was raised. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjUUID *string `json:"obj_uuid,omitempty"`

	// The object type to which the Alert Config is associated with. Valid object types are - Virtual Service, Pool, Service Engine. Enum options - VIRTUALSERVICE, POOL, HEALTHMONITOR, NETWORKPROFILE, APPLICATIONPROFILE, HTTPPOLICYSET, DNSPOLICY, SECURITYPOLICY, IPADDRGROUP, STRINGGROUP, SSLPROFILE, SSLKEYANDCERTIFICATE, NETWORKSECURITYPOLICY, APPLICATIONPERSISTENCEPROFILE, ANALYTICSPROFILE, VSDATASCRIPTSET, TENANT, PKIPROFILE, AUTHPROFILE, CLOUD.... Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjectType *string `json:"object_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Recommendation *string `json:"recommendation,omitempty"`

	// Only if the Number of Events is reached or exceeded within the Time Window will an alert be generated. Allowed values are 1-31536000. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RollingWindow *uint32 `json:"rolling_window,omitempty"`

	// Signifies system events or the type of client logsused in this alert configuration. Enum options - CONN_LOGS, APP_LOGS, EVENT_LOGS, METRICS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Source *string `json:"source"`

	// Summary of reason why alert is generated. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Summary *string `json:"summary,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// An alert is created only when the number of events meets or exceeds this number within the chosen time frame. Allowed values are 1-65536. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Threshold *uint32 `json:"threshold,omitempty"`

	// Alerts are suppressed (throttled) for this duration of time since the last alert was raised for this alert config. Allowed values are 0-31536000. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Throttle *uint32 `json:"throttle,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
