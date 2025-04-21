// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Alert alert
// swagger:model Alert
type Alert struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Output of the alert action script. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActionScriptOutput *string `json:"action_script_output,omitempty"`

	//  It is a reference to an object of type AlertConfig. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AlertConfigRef *string `json:"alert_config_ref"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppEvents []*ApplicationLog `json:"app_events,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnEvents []*ConnectionLog `json:"conn_events,omitempty"`

	// alert generation criteria. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// List of event pages this alert is associated with. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EventPages []string `json:"event_pages,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Events []*EventLog `json:"events,omitempty"`

	// Unix Timestamp of the last throttling in seconds. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastThrottleTimestamp *float64 `json:"last_throttle_timestamp,omitempty"`

	// Resolved Alert Type. Enum options - ALERT_LOW, ALERT_MEDIUM, ALERT_HIGH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Level *string `json:"level"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricInfo []*MetricLog `json:"metric_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// UUID of the resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ObjKey *string `json:"obj_key"`

	// Name of the resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjName *string `json:"obj_name,omitempty"`

	// UUID of the resource. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ObjUUID *string `json:"obj_uuid"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Reason *string `json:"reason"`

	// related uuids for the connection log. Only Log agent needs to fill this. Server uuid should be in formatpool_uuid-ip-port. In case of no port is set for server it shouldstill be operational port for the server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RelatedUuids []string `json:"related_uuids,omitempty"`

	// State of the alert. It would be active when createdIt would be changed to state read when read by the admin. Enum options - ALERT_STATE_ON, ALERT_STATE_DISMISSED, ALERT_STATE_THROTTLED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	State *string `json:"state"`

	// summary of alert based on alert config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Summary *string `json:"summary"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Threshold uint32 `json:"threshold,omitempty"`

	// Number of times it was throttled. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ThrottleCount uint32 `json:"throttle_count,omitempty"`

	// Unix Timestamp of the last throttling in seconds. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Timestamp *float64 `json:"timestamp"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
