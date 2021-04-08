package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertMgrDebugFilter alert mgr debug filter
// swagger:model AlertMgrDebugFilter
type AlertMgrDebugFilter struct {

	// filter debugs for entity uuid.
	AlertObjid *string `json:"alert_objid,omitempty"`

	// filter debugs for an alert id.
	AlertUUID *string `json:"alert_uuid,omitempty"`

	// filter debugs for an alert config.
	CfgUUID *string `json:"cfg_uuid,omitempty"`
}
