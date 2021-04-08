package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Alert alert
// swagger:model Alert
type Alert struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Output of the alert action script.
	ActionScriptOutput *string `json:"action_script_output,omitempty"`

	//  It is a reference to an object of type AlertConfig.
	// Required: true
	AlertConfigRef *string `json:"alert_config_ref"`

	// Placeholder for description of property app_events of obj type Alert field type str  type object
	AppEvents []*ApplicationLog `json:"app_events,omitempty"`

	// Placeholder for description of property conn_events of obj type Alert field type str  type object
	ConnEvents []*ConnectionLog `json:"conn_events,omitempty"`

	// alert generation criteria.
	Description *string `json:"description,omitempty"`

	// List of event pages this alert is associated with.
	EventPages []string `json:"event_pages,omitempty"`

	// Placeholder for description of property events of obj type Alert field type str  type object
	Events []*EventLog `json:"events,omitempty"`

	// Unix Timestamp of the last throttling in seconds.
	LastThrottleTimestamp *float64 `json:"last_throttle_timestamp,omitempty"`

	// Resolved Alert Type. Enum options - ALERT_LOW, ALERT_MEDIUM, ALERT_HIGH.
	// Required: true
	Level *string `json:"level"`

	// Placeholder for description of property metric_info of obj type Alert field type str  type object
	MetricInfo []*MetricLog `json:"metric_info,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// UUID of the resource.
	// Required: true
	ObjKey *string `json:"obj_key"`

	// Name of the resource.
	ObjName *string `json:"obj_name,omitempty"`

	// UUID of the resource.
	// Required: true
	ObjUUID *string `json:"obj_uuid"`

	// reason of Alert.
	// Required: true
	Reason *string `json:"reason"`

	// related uuids for the connection log. Only Log agent needs to fill this. Server uuid should be in formatpool_uuid-ip-port. In case of no port is set for server it shouldstill be operational port for the server.
	RelatedUuids []string `json:"related_uuids,omitempty"`

	// State of the alert. It would be active when createdIt would be changed to state read when read by the admin. Enum options - ALERT_STATE_ON, ALERT_STATE_DISMISSED, ALERT_STATE_THROTTLED.
	// Required: true
	State *string `json:"state"`

	// summary of alert based on alert config.
	// Required: true
	Summary *string `json:"summary"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Number of threshold.
	Threshold *int32 `json:"threshold,omitempty"`

	// Number of times it was throttled.
	ThrottleCount *int32 `json:"throttle_count,omitempty"`

	// Unix Timestamp of the last throttling in seconds.
	// Required: true
	Timestamp *float64 `json:"timestamp"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
